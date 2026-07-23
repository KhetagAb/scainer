package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"

	"scainer/internal/configs"
	"scainer/internal/domain"
	"scainer/internal/repository"
	"scainer/internal/services/analyze"
	"scainer/internal/services/contests"
	"scainer/internal/services/detect"
	"scainer/internal/services/detect/aiusage"
	"scainer/internal/services/detect/jplag"
	"scainer/internal/services/ejudge"
	ejgateway "scainer/internal/services/ejudge/gateway"
	"scainer/internal/services/importer"
	"scainer/internal/services/review"
	"scainer/internal/services/scoring"
	"scainer/internal/services/teachers"
	"scainer/internal/transport"
	"scainer/pkg/auth"
	"scainer/pkg/jobs"
	"scainer/pkg/llm"
	"scainer/pkg/llm/openai"
	"scainer/pkg/store"
)

const defaultJudgeSystem = "ejudge"

func main() {
	configPath := flag.String("config", "", "путь к config.yaml (или CONFIG_PATH)")
	flag.Parse()

	cfg, err := configs.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "scainer:", err)
		os.Exit(1)
	}

	if err := run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "scainer:", err)
		os.Exit(1)
	}
}

func run(cfg *configs.Config) error {
	store, err := store.NewFS(cfg.Store.Dir)
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}

	mongo, err := openMongo(cfg.MongoDB)
	if err != nil {
		return err
	}
	defer mongo.client.Disconnect(context.Background())

	if err := seedTeachers(context.Background(), cfg.TeachersPasswords, mongo.teachers); err != nil {
		return fmt.Errorf("teachers: %w", err)
	}

	teachersSvc := teachers.NewService(mongo.teachers)
	ejGateway := ejgateway.New(mongo.teachers, mongo.credentials, cfg.Ejudge.BaseURL, cfg.Ejudge.Timeout)
	if cfg.Ejudge.Enabled() {
		importer.Register("ejudge", ejudge.ImporterFactory(ejGateway))
		fmt.Fprintln(os.Stderr, "scainer: ejudge enabled")
	} else {
		fmt.Fprintln(os.Stderr, "scainer: ejudge disabled (нет base_url)")
	}

	pipeline, pool, err := buildAnalyzeRuntime(cfg, store)
	if err != nil {
		return err
	}

	svcs := wireServices(cfg, store, mongo, ejGateway, teachersSvc, pipeline, pool)
	e := transport.New(svcs.contests, svcs.reader, svcs.analyze, svcs.review, svcs.teachers, svcs.auth).Echo()
	return serveHTTP(cfg.HTTP, e)
}

type mongoDeps struct {
	client        *mongo.Client
	teachers      *repository.TeachersRepository
	credentials   *repository.EjudgeCredentialsRepository
	registry      contests.ContestRegistry
	findingsRepo contests.FindingsRepository
}

func seedTeachers(ctx context.Context, passwords map[string]string, repo *repository.TeachersRepository) error {
	if len(passwords) == 0 {
		return nil
	}
	if err := repo.UpsertPasswords(ctx, passwords); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "scainer: seeded %d teacher(s) from TEACHERS_PASSWORDS\n", len(passwords))
	return nil
}

func openMongo(mc configs.MongoDBConfig) (mongoDeps, error) {
	uri, err := mc.URI()
	if err != nil {
		return mongoDeps{}, err
	}
	client, db, err := repository.Connect(context.Background(), uri, mc.Database)
	if err != nil {
		return mongoDeps{}, fmt.Errorf("mongo: %w", err)
	}
	return mongoDeps{
		client:        client,
		teachers:      repository.NewTeachersRepository(db),
		credentials:   repository.NewEjudgeCredentialsRepository(db),
		registry:      repository.NewContestRepository(db),
		findingsRepo: repository.NewFindingsRepository(db),
	}, nil
}

func buildAnalyzeRuntime(cfg *configs.Config, st *store.FS) (detect.Pipeline, *jobs.Pool, error) {
	det, err := jplag.New(cfg.JPlag.JarPath, st)
	if err != nil {
		return nil, nil, fmt.Errorf("jplag: %w", err)
	}
	if err := det.CheckRuntime(); err != nil {
		return nil, nil, err
	}

	// Один Limiter на процесс: иначе jobs_max_concurrent job'ов перемножили бы параллелизм JPlag/LLM.
	limiter := detect.NewLimiter(cfg.Analyze.AnalyzeConcurrency)
	factories := []detect.StageFactory{
		func(id domain.ContestID) detect.Stage {
			return detect.NewStage(detect.ProblemSelector{Contest: id}, limiter, det)
		},
	}
	if cfg.AIUsage.Enabled {
		model, err := openAIModel(cfg.OpenAI)
		if err != nil {
			return nil, nil, fmt.Errorf("aiusage: %w", err)
		}
		analyzer := &aiusage.Analyzer{Model: model}
		taskDet := aiusage.NewTaskDetector(analyzer)
		factories = append(factories, func(id domain.ContestID) detect.Stage {
			return detect.NewStage(detect.OkWithLastSelector{Contest: id}, limiter, taskDet)
		})
		fmt.Fprintln(os.Stderr, "scainer: aiusage enabled")
	}

	return detect.Compose(factories...), jobs.NewPool(cfg.Analyze.JobsMaxConcurrent), nil
}

func openAIModel(oc configs.OpenAIConfig) (llm.IntelligenceModel, error) {
	baseURL := oc.BaseURL
	if baseURL == "" {
		baseURL = llm.DefaultBaseURL
	}
	var (
		client *openai.Client
		err    error
	)
	if oc.Username != "" || oc.Password != "" {
		client, err = openai.New(baseURL, oc.Username, oc.Password, oc.Timeout)
	} else if oc.APIKey != "" {
		client, err = openai.NewWithAPIKey(baseURL, oc.APIKey, oc.Timeout)
	} else {
		return nil, fmt.Errorf("задайте openai.username/password или openai.api_key")
	}
	if err != nil {
		return nil, err
	}
	return llm.NewOpenAI(client, oc.Model)
}

type appServices struct {
	contests *contests.Service
	reader   *contests.ContestReader
	analyze  *analyze.Service
	review   *review.Service
	teachers *teachers.Service
	auth     auth.Service
}

func wireServices(
	cfg *configs.Config,
	store *store.FS,
	mongo mongoDeps,
	ejGateway *ejgateway.Gateway,
	teachersSvc *teachers.Service,
	pipeline detect.Pipeline,
	pool *jobs.Pool,
) appServices {
	scorer := scoring.NewWeighted()

	var comments review.CommentsProvider
	var status review.StatusProvider
	if cfg.Ejudge.Enabled() {
		comments = ejudge.NewComments(ejGateway)
		status = ejudge.NewStatus(ejGateway)
	}

	return appServices{
		contests: contests.NewService(mongo.registry, mongo.findingsRepo, defaultJudgeSystem),
		reader:   contests.NewContestReader(mongo.registry, store, mongo.findingsRepo),
		analyze:  analyze.New(pool, analyze.NewRunner(mongo.registry, store, mongo.findingsRepo, scorer, pipeline)),
		review:   review.New(store, comments, status),
		teachers: teachersSvc,
		auth:     auth.New(cfg.Admin.JWTSecret, cfg.Admin.JWTTTL),
	}
}

func serveHTTP(hc configs.HTTPConfig, e *echo.Echo) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		fmt.Fprintf(os.Stderr, "scainer: listening on %s\n", hc.Addr)
		errCh <- e.Start(hc.Addr)
	}()

	select {
	case <-ctx.Done():
		timeout := hc.ShutdownTimeout
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := e.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		return nil
	case err := <-errCh:
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
