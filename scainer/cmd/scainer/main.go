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

	"scainer/internal/analyze"
	"scainer/internal/configs"
	"scainer/internal/contests"
	"scainer/internal/detect"
	"scainer/internal/detect/aiusage"
	"scainer/internal/detect/jplag"
	"scainer/internal/domain"
	"scainer/internal/importer"
	ejimporter "scainer/internal/importer/ejudge"
	"scainer/pkg/jobs"
	"scainer/pkg/llm"
	"scainer/internal/repository"
	"scainer/internal/review"
	ejreview "scainer/internal/review/ejudge"
	"scainer/internal/scoring"
	"scainer/internal/store"
	"scainer/internal/transport"
	"scainer/pkg/auth"
	ejudgeapi "scainer/pkg/ejudge"
	"scainer/pkg/openai"

	_ "scainer/internal/importer/folder"
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

	ejClient, err := openEjudge(cfg.Ejudge)
	if err != nil {
		return err
	}
	if ejClient != nil {
		importer.Register("ejudge", ejimporter.Factory(ejClient))
	}

	pipeline, pool, err := buildAnalyzeRuntime(cfg, store)
	if err != nil {
		return err
	}

	svcs := wireServices(cfg, store, mongo, ejClient, pipeline, pool)
	e := transport.New(svcs.contests, svcs.reader, svcs.analyze, svcs.review, svcs.auth).Echo()
	return serveHTTP(cfg.HTTP, e)
}

type mongoDeps struct {
	client        *mongo.Client
	registry      contests.ContestRegistry
	findingsStore contests.FindingsStore
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
		registry:      repository.NewContestRepository(db),
		findingsStore: repository.NewFindingsRepository(db),
	}, nil
}

func openEjudge(ec configs.EjudgeConfig) (*ejudgeapi.Client, error) {
	if !ec.Enabled() {
		fmt.Fprintln(os.Stderr, "scainer: ejudge disabled (нет base_url/api_key)")
		return nil, nil
	}
	client, err := ejudgeapi.New(ec.BaseURL, ec.APIKey, ec.Timeout)
	if err != nil {
		return nil, fmt.Errorf("ejudge: %w", err)
	}
	return client, nil
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
	auth     auth.Service
}

func wireServices(
	cfg *configs.Config,
	store *store.FS,
	mongo mongoDeps,
	ejClient *ejudgeapi.Client,
	pipeline detect.Pipeline,
	pool *jobs.Pool,
) appServices {
	scorer := scoring.NewWeighted()
	return appServices{
		contests: contests.NewService(mongo.registry, mongo.findingsStore, defaultJudgeSystem),
		reader:   contests.NewContestReader(mongo.registry, store, mongo.findingsStore),
		analyze:  analyze.New(pool, analyze.NewRunner(mongo.registry, store, mongo.findingsStore, scorer, pipeline)),
		review: review.New(store, &ejreview.Comments{Client: ejClient}, &ejreview.Status{Client: ejClient}),
		auth: auth.New(
			cfg.Admin.Username,
			cfg.Admin.Password,
			cfg.Admin.JWTSecret,
			cfg.Admin.JWTTTL,
		),
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
