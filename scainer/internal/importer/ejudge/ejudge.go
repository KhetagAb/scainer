package ejudge

import (
	"context"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"

	"scainer/internal/domain"
	"scainer/internal/importer"
	"scainer/internal/progress"
	"scainer/internal/store"
	ejudgeapi "scainer/pkg/ejudge"
)

func init() { importer.Register("ejudge", newFromConfig) }

const (
	cursorKeyPrefix = "ejudge:cursor:"
	// ejudge не принимает «∞» в list-runs.
	lastRunOpen = 1_000_000_000
)

type Importer struct {
	cfg ejudgeapi.Config
	env *ejudgeapi.Env
}

var _ importer.Importer = (*Importer)(nil)

func (i *Importer) Name() string { return "ejudge" }

func newFromConfig(node *yaml.Node) (importer.Importer, error) {
	var cfg ejudgeapi.Config
	if node != nil && node.Kind != 0 {
		if err := node.Decode(&cfg); err != nil {
			return nil, err
		}
	}
	if cfg.ContestID <= 0 {
		return nil, fmt.Errorf("ejudge importer: contest_id должен быть > 0")
	}

	env, err := ejudgeapi.LoadEnv()
	if err != nil {
		return nil, err
	}
	return &Importer{cfg: cfg, env: env}, nil
}

func cursorKey(contestID int) string {
	return cursorKeyPrefix + strconv.Itoa(contestID)
}

func (i *Importer) Import(ctx context.Context, s store.Store) ([]domain.Submission, error) {
	if i.env == nil || i.env.Client == nil {
		return nil, fmt.Errorf("ejudge importer: клиент не инициализирован")
	}
	contestID := i.cfg.ContestID
	if contestID <= 0 {
		return nil, fmt.Errorf("ejudge importer: contest_id должен быть > 0")
	}

	firstRun := 0
	cur, ok, err := s.GetCursor(ctx, cursorKey(contestID))
	if err != nil {
		return nil, err
	}
	if ok && cur != "" {
		n, err := strconv.Atoi(cur)
		if err != nil {
			return nil, fmt.Errorf("ejudge importer: невалидный курсор %q: %w", cur, err)
		}
		firstRun = n + 1
	}

	lastRun := lastRunOpen
	listResp, err := i.env.Client.ListRunsWithResponse(ctx, ejudgeapi.ListRunsParams(contestID, &firstRun, &lastRun))
	if err != nil {
		return nil, err
	}
	if listResp.StatusCode() != 200 || listResp.JSON200 == nil {
		return nil, fmt.Errorf("ejudge list-runs: HTTP %d: %s", listResp.StatusCode(), truncate(listResp.Body, 200))
	}
	if err := ejudgeapi.EnsureOK(listResp.JSON200.Ok, listResp.JSON200.Error); err != nil {
		return nil, err
	}
	if listResp.JSON200.Result == nil || listResp.JSON200.Result.Runs == nil {
		return nil, nil
	}
	runs := *listResp.JSON200.Result.Runs
	if len(runs) == 0 {
		return nil, nil
	}

	contestName := ""
	if info, err := i.env.Client.ContestStatus(ctx, contestID); err != nil {
		return nil, fmt.Errorf("ejudge contest-status: %w", err)
	} else {
		contestName = info.Name
	}

	subs := make([]domain.Submission, 0, len(runs))
	maxRunID := -1
	base := countContestSubmissions(ctx, s, domain.ContestID(strconv.Itoa(contestID)))
	batch := len(runs)
	total := base + batch

	// Абсолютный прогресс: уже в store + текущая догрузка (в случае инкрементального импорта).
	progress.Report(ctx, progress.Event{Phase: "importing", Done: base, Total: total})
	for idx, run := range runs {
		if run.RunId == nil {
			return nil, fmt.Errorf("ejudge: ран без run_id")
		}
		runID := *run.RunId
		// Без участника пропускаем. Курсор двигаем только после успешного импорта —
		// иначе пропуски «сжигают» курсор, и повторный импорт уже ничего не видит.
		if (run.UserLogin == nil || *run.UserLogin == "") && (run.UserName == nil || *run.UserName == "") {
			progress.Report(ctx, progress.Event{Phase: "importing", Done: base + idx + 1, Total: total})
			continue
		}

		dl, err := i.env.Client.DownloadRunWithResponse(ctx, ejudgeapi.DownloadRunParams(contestID, runID))
		if err != nil {
			return nil, err
		}
		if dl.StatusCode() != 200 {
			return nil, fmt.Errorf("ejudge download-run run_id=%d: HTTP %d", runID, dl.StatusCode())
		}

		sub, err := mapToSubmission(contestID, run, dl.Body, i.env, contestName)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
		if runID > maxRunID {
			maxRunID = runID
		}
		progress.Report(ctx, progress.Event{Phase: "importing", Done: base + idx + 1, Total: total})
	}

	if maxRunID >= 0 {
		if err := s.SetCursor(ctx, cursorKey(contestID), strconv.Itoa(maxRunID)); err != nil {
			return nil, err
		}
	}
	return subs, nil
}

func countContestSubmissions(ctx context.Context, s store.Store, contest domain.ContestID) int {
	byProblem, err := s.ByProblem(ctx, contest)
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
