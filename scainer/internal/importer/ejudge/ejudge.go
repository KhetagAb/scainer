package ejudge

import (
	"context"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"

	"scainer/internal/domain"
	"scainer/internal/importer"
	"scainer/internal/progress"
	ejudgeapi "scainer/pkg/ejudge"
)

func Factory(client *ejudgeapi.Client) importer.Factory {
	return func(node *yaml.Node) (importer.Importer, error) {
		if client == nil {
			return nil, fmt.Errorf("ejudge importer: клиент не задан")
		}
		var cfg Config
		if node != nil && node.Kind != 0 {
			if err := node.Decode(&cfg); err != nil {
				return nil, err
			}
		}
		if cfg.ContestID <= 0 {
			return nil, fmt.Errorf("ejudge importer: contest_id должен быть > 0")
		}
		return &Importer{cfg: cfg, client: client}, nil
	}
}

const (
	cursorKeyPrefix = "ejudge:cursor:"
	// ejudge не принимает «∞» в list-runs.
	lastRunOpen = 1_000_000_000
)

type Importer struct {
	cfg    Config
	client *ejudgeapi.Client
}

var _ importer.Importer = (*Importer)(nil)

func (i *Importer) Name() string { return "ejudge" }

func cursorKey(contestID int) string {
	return cursorKeyPrefix + strconv.Itoa(contestID)
}

func (i *Importer) Import(ctx context.Context, store importer.Store) (importer.Result, error) {
	contestID := i.cfg.ContestID
	if contestID <= 0 {
		return importer.Result{}, fmt.Errorf("ejudge importer: contest_id должен быть > 0")
	}

	info, err := i.client.ContestStatus(ctx, contestID)
	if err != nil {
		return importer.Result{}, fmt.Errorf("ejudge contest-status: %w", err)
	}
	contestName := info.Name

	firstRun := 0
	cur, ok, err := store.GetCursor(ctx, cursorKey(contestID))
	if err != nil {
		return importer.Result{}, err
	}
	if ok && cur != "" {
		n, err := strconv.Atoi(cur)
		if err != nil {
			return importer.Result{}, fmt.Errorf("ejudge importer: невалидный курсор %q: %w", cur, err)
		}
		firstRun = n + 1
	}

	lastRun := lastRunOpen
	listReply, err := i.client.ListRuns(ctx, contestID, &firstRun, &lastRun)
	if err != nil {
		return importer.Result{}, err
	}
	if listReply.Result == nil || listReply.Result.Runs == nil || len(*listReply.Result.Runs) == 0 {
		return importer.Result{ContestName: contestName}, nil
	}
	runs := *listReply.Result.Runs

	subs := make([]domain.Submission, 0, len(runs))
	maxRunID := -1
	base := countContestSubmissions(ctx, store, domain.ContestID(strconv.Itoa(contestID)))
	batch := len(runs)
	total := base + batch

	// Абсолютный прогресс: уже в store + текущая догрузка (в случае инкрементального импорта).
	progress.Report(ctx, progress.Event{Phase: "importing", Done: base, Total: total})
	for idx, run := range runs {
		if run.RunId == nil {
			return importer.Result{}, fmt.Errorf("ejudge: ран без run_id")
		}
		runID := *run.RunId
		// Без участника пропускаем. Курсор двигаем только после успешного импорта —
		// иначе пропуски «сжигают» курсор, и повторный импорт уже ничего не видит.
		if (run.UserLogin == nil || *run.UserLogin == "") && (run.UserName == nil || *run.UserName == "") {
			progress.Report(ctx, progress.Event{Phase: "importing", Done: base + idx + 1, Total: total})
			continue
		}

		dl, err := i.client.DownloadRunWithResponse(ctx, ejudgeapi.DownloadRunParams(contestID, runID))
		if err != nil {
			return importer.Result{}, err
		}
		if dl.StatusCode() != 200 {
			return importer.Result{}, fmt.Errorf("ejudge download-run run_id=%d: HTTP %d", runID, dl.StatusCode())
		}

		sub, err := mapToSubmission(contestID, run, dl.Body, contestName)
		if err != nil {
			return importer.Result{}, err
		}
		subs = append(subs, sub)
		if runID > maxRunID {
			maxRunID = runID
		}
		progress.Report(ctx, progress.Event{Phase: "importing", Done: base + idx + 1, Total: total})
	}

	if maxRunID >= 0 {
		if err := store.SetCursor(ctx, cursorKey(contestID), strconv.Itoa(maxRunID)); err != nil {
			return importer.Result{}, err
		}
	}
	return importer.Result{Submissions: subs, ContestName: contestName}, nil
}

func countContestSubmissions(ctx context.Context, store importer.Store, contest domain.ContestID) int {
	byProblem, err := store.ByProblem(ctx, contest)
	if err != nil {
		return 0
	}
	n := 0
	for _, list := range byProblem {
		n += len(list)
	}
	return n
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "…"
}
