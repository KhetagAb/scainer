package analyze

import (
	"context"
	"fmt"
	"os"
	"time"

	"scainer/internal/configs"
	"scainer/internal/domain"
	"scainer/internal/services/analyze/detect"
	"scainer/internal/services/analyze/detect/aiusage"
	"scainer/internal/services/analyze/detect/jplag"
	"scainer/internal/services/analyze/detect/nightsubmit"
	"scainer/internal/services/analyze/selectors"
	"scainer/internal/services/contests"
	"scainer/pkg/jobs"
	"scainer/pkg/store"
)

type contestRun struct {
	ctx       context.Context
	contestID domain.ContestID
	store     selectors.Store
}

type Registry struct {
	limiter *detect.Limiter
	runs    []func(contestRun, *contests.AnalysisSnapshot) error
}

type Orchestrator struct {
	registry *Registry
}

type Result struct {
	ImportedCount  int
	LastImportedAt time.Time
}

type Runner struct {
	registry        contests.ContestRegistry
	submissionStore contests.SubmissionStore
	analysisRepo    contests.AnalysisRepository
	orchestrator    *Orchestrator
}

func NewRegistry(limiter *detect.Limiter) *Registry {
	return &Registry{limiter: limiter}
}

func (r *Registry) Runs() []func(contestRun, *contests.AnalysisSnapshot) error {
	return r.runs
}

func Register[U domain.Unit](
	r *Registry,
	det detect.Detector[U],
	policy InvalidationPolicy[U],
	selector func(domain.ContestID) selectors.Selector[U],
) {
	binding := detectorRun[U]{
		name:     det.Name(),
		det:      det,
		policy:   policy,
		selector: selector,
		limiter:  r.limiter,
	}
	r.runs = append(r.runs, binding.run)
}

func NewRuntime(cfg *configs.Config, st *store.FS) (*Orchestrator, *jobs.Pool, error) {
	jplagDet, err := jplag.NewFromConfig(cfg.JPlag, st)
	if err != nil {
		return nil, nil, err
	}
	aiUsage, err := aiusage.NewFromConfig(cfg.AIUsage, cfg.OpenAI)
	if err != nil {
		return nil, nil, err
	}
	if aiUsage != nil {
		fmt.Fprintln(os.Stderr, "scainer: aiusage enabled")
	}

	limiter := detect.NewLimiter(cfg.Analyze.AnalyzeConcurrency)
	r := NewRegistry(limiter)
	RegisterDetectors(r, nightsubmit.NewDetector(), jplagDet, aiUsage)
	return NewOrchestrator(r), jobs.NewPool(cfg.Analyze.JobsMaxConcurrent), nil
}

func RegisterDetectors(
	r *Registry,
	nightSubmit detect.Detector[domain.StandaloneUnit],
	jplagDet detect.Detector[domain.ProblemUnit],
	aiUsage detect.Detector[domain.ProblemParticipantUnit],
) {
	Register(r, nightSubmit, Once[domain.StandaloneUnit]{Key: ScopeKeySubmission}, selectors.Submissions)
	Register(r, jplagDet, UnlessChanged[domain.ProblemUnit]{
		Key:    ScopeKeyProblem,
		SubIDs: SubmissionIDsFromProblemUnit,
	}, selectors.Problems)
	if aiUsage != nil {
		Register(r, aiUsage, ManualOnly[domain.ProblemParticipantUnit]{Key: ScopeKeyProblemParticipant}, selectors.OkWithLast)
	}
}
