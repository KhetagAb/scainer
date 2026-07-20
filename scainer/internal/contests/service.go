package contests

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"

	"scainer/internal/detect"
	"scainer/internal/domain"
	"scainer/internal/importer"
	"scainer/internal/jobs"
	"scainer/internal/scoring"
	"scainer/internal/store"
	ejudgeapi "scainer/pkg/ejudge"
)

var ErrContestNotFound = errors.New("contest not found")
var ErrDuplicateContest = errors.New("contest already registered")

const lastImportCursorPrefix = "http:last_import:"

type SourceSpec struct {
	Type   string
	Config yaml.Node
}

type Registration struct {
	ID               domain.ContestID
	ParallelID       domain.ParallelID
	ParallelName     string
	Source           *SourceSpec
	ExcludedProblems []domain.ProblemID
}

type ContestStatistic struct {
	SubmissionCount int
	ProblemCount    int
	FindingsCount   int
}

type Contest struct {
	ID               domain.ContestID   `bson:"id"`
	Name             string             `bson:"name"`
	ParallelID       domain.ParallelID  `bson:"parallel_id"`
	ParallelName     string             `bson:"parallel_name"`
	ExcludedProblems []domain.ProblemID `bson:"excluded_problems"`
	LastImportedAt   *time.Time         `bson:"-"`
	Statistic        ContestStatistic   `bson:"-"`
}

type ProblemInfo struct {
	ID              domain.ProblemID
	Name            string
	Excluded        bool
	SubmissionCount int
}

type ImportResult struct {
	ImportedCount  int
	LastImportedAt time.Time
}

type RuntimeConfig struct {
	Pool           *jobs.Pool
	AnalyzeLimiter *detect.Limiter
}

type Service struct {
	store              store.Store
	scorer             scoring.Scorer
	defaultJudgeSystem string
	dets               []detect.Detector[domain.ProblemUnit]
	registry           ContestRegistry
	findingsStore      FindingsStore
	pool               *jobs.Pool
	analyzeLimiter     *detect.Limiter
}

func New(st store.Store, scorer scoring.Scorer, defaultJudgeSystem string, registry ContestRegistry, findingsStore FindingsStore, cfg RuntimeConfig, dets ...detect.Detector[domain.ProblemUnit]) *Service {
	return &Service{
		store:              st,
		scorer:             scorer,
		defaultJudgeSystem: defaultJudgeSystem,
		registry:           registry,
		findingsStore:      findingsStore,
		pool:               cfg.Pool,
		analyzeLimiter:     cfg.AnalyzeLimiter,
		dets:               dets,
	}
}

func (s *Service) get(ctx context.Context, id domain.ContestID) (ContestRecord, error) {
	rec, ok, err := s.registry.Get(ctx, id)
	if err != nil {
		return ContestRecord{}, err
	}
	if !ok {
		return ContestRecord{}, ErrContestNotFound
	}
	return rec, nil
}

func (s *Service) buildRuntime(id domain.ContestID, source SourceSpec) (detect.Stage, []importer.Importer, error) {
	imp, err := importer.Build(source.Type, &source.Config)
	if err != nil {
		return nil, nil, fmt.Errorf("build importer: %w", err)
	}
	stage := detect.NewStage(detect.ProblemSelector{Contest: id}, s.analyzeLimiter, s.dets...)
	return stage, []importer.Importer{imp}, nil
}

func (s *Service) Register(ctx context.Context, reg Registration) (Contest, error) {
	if _, err := s.get(ctx, reg.ID); err == nil {
		return Contest{}, ErrDuplicateContest
	} else if !errors.Is(err, ErrContestNotFound) {
		return Contest{}, err
	}

	source := reg.Source
	if source == nil {
		source = &SourceSpec{
			Type:   s.defaultJudgeSystem,
			Config: yaml.Node{},
		}
	}

	// ejudge: contest_id всегда = ID контеста в scainer; PostContests Source не заполняет.
	contestName := ""
	if source.Type == "ejudge" {
		cfgNode, err := ejudgeConfigNode(reg.ID)
		if err != nil {
			return Contest{}, err
		}
		source = &SourceSpec{Type: "ejudge", Config: *cfgNode}

		contestName, err = ejudgeContestName(ctx, reg.ID)
		if err != nil {
			return Contest{}, err
		}
	}

	if _, _, err := s.buildRuntime(reg.ID, *source); err != nil {
		return Contest{}, err
	}

	contest := Contest{
		ID:               reg.ID,
		Name:             contestName,
		ParallelID:       reg.ParallelID,
		ParallelName:     reg.ParallelName,
		ExcludedProblems: reg.ExcludedProblems,
	}

	if err := s.registry.Put(ctx, ContestRecord{Contest: contest, Source: *source}); err != nil {
		return Contest{}, fmt.Errorf("persist contest: %w", err)
	}

	return contest, nil
}

func (s *Service) List(ctx context.Context) ([]Contest, error) {
	recs, err := s.registry.List(ctx)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(recs, func(a, b ContestRecord) int {
		return compareContestID(a.Contest.ID, b.Contest.ID)
	})

	out := make([]Contest, 0, len(recs))
	for _, rec := range recs {
		contest := rec.Contest
		raw, ok, err := s.store.GetCursor(ctx, lastImportKey(rec.Contest.ID))
		if err != nil {
			return nil, err
		}
		if ok && raw != "" {
			t, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				return nil, fmt.Errorf("parse lastImportedAt for %s: %w", rec.Contest.ID, err)
			}
			contest.LastImportedAt = &t
		}
		if err := s.enrichStats(ctx, &contest); err != nil {
			return nil, err
		}
		out = append(out, contest)
	}
	return out, nil
}

func (s *Service) enrichStats(ctx context.Context, contest *Contest) error {
	subs, problems, err := countStore(ctx, s.store, contest.ID)
	if err != nil {
		return err
	}
	contest.Statistic.SubmissionCount = subs
	contest.Statistic.ProblemCount = problems

	snap, ok, err := s.findingsStore.Get(ctx, contest.ID)
	if err != nil {
		return err
	}
	if !ok {
		contest.Statistic.FindingsCount = 0
		return nil
	}
	contest.Statistic.FindingsCount = len(snap.Findings)
	return nil
}

func countStore(ctx context.Context, st store.Store, id domain.ContestID) (submissions int, problems int, err error) {
	byProblem, err := st.ByProblem(ctx, id)
	if err != nil {
		return 0, 0, fmt.Errorf("stats ByProblem %s: %w", id, err)
	}
	problems = len(byProblem)
	for _, list := range byProblem {
		submissions += len(list)
	}
	return submissions, problems, nil
}

func (s *Service) SetParallel(ctx context.Context, id domain.ContestID, parallelID domain.ParallelID, parallelName string) error {
	rec, err := s.get(ctx, id)
	if err != nil {
		return err
	}
	contest := rec.Contest
	contest.ParallelID = parallelID
	contest.ParallelName = parallelName

	if err := s.registry.Put(ctx, ContestRecord{Contest: contest, Source: rec.Source}); err != nil {
		return fmt.Errorf("persist contest: %w", err)
	}
	return nil
}

func (s *Service) RemoveContest(ctx context.Context, id domain.ContestID) error {
	if _, err := s.get(ctx, id); err != nil {
		return err
	}
	if err := s.registry.Delete(ctx, id); err != nil {
		return fmt.Errorf("persist contest removal: %w", err)
	}
	if err := s.findingsStore.Delete(ctx, id); err != nil {
		return fmt.Errorf("persist findings removal: %w", err)
	}
	return nil
}

func (s *Service) SetExcludedProblems(ctx context.Context, id domain.ContestID, problems []domain.ProblemID) error {
	rec, err := s.get(ctx, id)
	if err != nil {
		return err
	}
	contest := rec.Contest
	contest.ExcludedProblems = problems

	if err := s.registry.Put(ctx, ContestRecord{Contest: contest, Source: rec.Source}); err != nil {
		return fmt.Errorf("persist contest: %w", err)
	}
	return nil
}

func (s *Service) SubmitImport(ctx context.Context, id domain.ContestID) (string, error) {
	if _, err := s.get(ctx, id); err != nil {
		return "", err
	}
	jobID := s.pool.Submit(ctx, func(jobCtx context.Context) error {
		_, _, _, err := s.runImport(jobCtx, id)
		return err
	})
	return jobID, nil
}

func (s *Service) JobStatus(id string) (jobs.State, bool) {
	return s.pool.Get(id)
}

func (s *Service) SubscribeJob(id string) (<-chan jobs.State, func(), bool) {
	return s.pool.Subscribe(id)
}

func (s *Service) runImport(ctx context.Context, id domain.ContestID) (ImportResult, []domain.Finding, map[domain.SubmissionID]domain.Submission, error) {
	rec, err := s.get(ctx, id)
	if err != nil {
		return ImportResult{}, nil, nil, err
	}

	stage, imps, err := s.buildRuntime(id, rec.Source)
	if err != nil {
		return ImportResult{}, nil, nil, err
	}

	imported := 0
	for _, imp := range imps {
		subs, err := imp.Import(ctx, s.store)
		if err != nil {
			return ImportResult{}, nil, nil, err
		}
		if err := s.store.Put(ctx, subs); err != nil {
			return ImportResult{}, nil, nil, err
		}
		imported += len(subs)
	}

	now := time.Now().UTC().Truncate(time.Second)
	if err := s.store.SetCursor(ctx, lastImportKey(id), now.Format(time.RFC3339)); err != nil {
		return ImportResult{}, nil, nil, err
	}
	findings, subs, err := s.recomputeFindings(ctx, id, rec.Contest, stage)
	if err != nil {
		return ImportResult{}, nil, nil, fmt.Errorf("recompute findings: %w", err)
	}
	return ImportResult{ImportedCount: imported, LastImportedAt: now}, findings, subs, nil
}

func (s *Service) recomputeFindings(ctx context.Context, id domain.ContestID, contest Contest, stage detect.Stage) ([]domain.Finding, map[domain.SubmissionID]domain.Submission, error) {
	excludedMap := make(map[domain.ProblemID]bool, len(contest.ExcludedProblems))
	for _, p := range contest.ExcludedProblems {
		excludedMap[p] = true
	}

	policy := detect.AnalysisPolicy{ExcludedProblems: excludedMap}
	signals, err := stage.Run(ctx, s.store, policy)
	if err != nil {
		return nil, nil, fmt.Errorf("analyze: %w", err)
	}
	findings := s.scorer.Score(signals)
	if err := s.findingsStore.Put(ctx, FindingsSnapshot{
		ContestID: id, Findings: findings, ComputedAt: time.Now().UTC(),
	}); err != nil {
		return nil, nil, err
	}
	subs, err := loadReferencedSubmissions(ctx, s.store, findings)
	if err != nil {
		return nil, nil, err
	}
	return findings, subs, nil
}

func (s *Service) GetFindings(ctx context.Context, id domain.ContestID) ([]domain.Finding, map[domain.SubmissionID]domain.Submission, error) {
	if _, err := s.get(ctx, id); err != nil {
		return nil, nil, err
	}

	snap, ok, err := s.findingsStore.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, nil
	}

	subs, err := loadReferencedSubmissions(ctx, s.store, snap.Findings)
	if err != nil {
		return nil, nil, err
	}
	return snap.Findings, subs, nil
}

func (s *Service) Problems(ctx context.Context, id domain.ContestID) ([]ProblemInfo, error) {
	rec, err := s.get(ctx, id)
	if err != nil {
		return nil, err
	}

	problems, err := s.store.ByProblem(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get problems: %w", err)
	}

	excluded := make(map[domain.ProblemID]bool)
	for _, p := range rec.Contest.ExcludedProblems {
		excluded[p] = true
	}

	out := make([]ProblemInfo, 0, len(problems))
	for probID, subs := range problems {
		problemName := ""
		if len(subs) > 0 && len(subs[0].Meta) > 0 {
			if name, ok := subs[0].Meta["problem_name"]; ok {
				if nameStr, ok := name.(string); ok {
					problemName = nameStr
				}
			}
		}

		out = append(out, ProblemInfo{
			ID:              probID,
			Name:            problemName,
			Excluded:        excluded[probID],
			SubmissionCount: len(subs),
		})
	}
	return out, nil
}

func lastImportKey(id domain.ContestID) string {
	return lastImportCursorPrefix + string(id)
}

func compareContestID(a, b domain.ContestID) int {
	ai, aErr := strconv.Atoi(string(a))
	bi, bErr := strconv.Atoi(string(b))
	if aErr == nil && bErr == nil {
		return cmp.Compare(ai, bi)
	}
	return cmp.Compare(string(a), string(b))
}

func ejudgeConfigNode(contestID domain.ContestID) (*yaml.Node, error) {
	cid, err := strconv.Atoi(string(contestID))
	if err != nil {
		return nil, fmt.Errorf("invalid contest_id for ejudge: %s", contestID)
	}

	raw, err := yaml.Marshal(map[string]int{"contest_id": cid})
	if err != nil {
		return nil, fmt.Errorf("marshal ejudge config: %w", err)
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("unmarshal ejudge config: %w", err)
	}
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		return doc.Content[0], nil
	}
	return &doc, nil
}

func ejudgeContestName(ctx context.Context, contestID domain.ContestID) (string, error) {
	cid, err := strconv.Atoi(string(contestID))
	if err != nil {
		return "", fmt.Errorf("invalid contest_id for ejudge: %s", contestID)
	}

	env, err := ejudgeapi.LoadEnv()
	if err != nil {
		return "", fmt.Errorf("load ejudge env: %w", err)
	}
	if env == nil || env.Client == nil {
		return "", errors.New("ejudge client not initialized")
	}

	info, err := env.Client.ContestStatus(ctx, cid)
	if err != nil {
		return "", fmt.Errorf("get contest status: %w", err)
	}
	return info.Name, nil
}

func loadReferencedSubmissions(ctx context.Context, st store.Store, findings []domain.Finding) (map[domain.SubmissionID]domain.Submission, error) {
	seen := make(map[domain.SubmissionID]bool)
	var ids []domain.SubmissionID
	for _, f := range findings {
		for _, sig := range f.Signals {
			for _, ev := range sig.Evidence {
				for _, sp := range ev.Spans {
					if seen[sp.Submission] {
						continue
					}
					seen[sp.Submission] = true
					ids = append(ids, sp.Submission)
				}
			}
		}
	}
	list, err := st.Get(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make(map[domain.SubmissionID]domain.Submission, len(list))
	for _, sub := range list {
		out[sub.ID] = sub
	}
	return out, nil
}
