package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"scainer/internal/analyze"
	"scainer/internal/contests"
	"scainer/internal/detect"
	"scainer/internal/detect/aiusage"
	"scainer/internal/detect/jplag"
	"scainer/internal/domain"
	"scainer/internal/jobs"
	"scainer/internal/llm"
	"scainer/internal/repository"
	"scainer/internal/scoring"
	"scainer/internal/store"
	"scainer/internal/transport"
	"scainer/pkg/auth"

	_ "scainer/internal/importer/ejudge"
	_ "scainer/internal/importer/folder"
)

const defaultJudgeSystem = "ejudge"

const defaultStoreDir = "./data"

const defaultMongoDatabase = "scainer"

const defaultJobsMaxConcurrent = 4

// JPlag — java-подпроцесс; слишком большое число запустит слишком много JVM разом.
const defaultAnalyzeConcurrency = 4

const (
	envAIUsageEnabled = "AIUSAGE_ENABLED"
)

func main() {
	_ = godotenv.Load()

	addr := flag.String("addr", ":8080", "адрес HTTP-сервера")
	flag.Parse()

	if err := run(*addr); err != nil {
		fmt.Fprintln(os.Stderr, "scainer:", err)
		os.Exit(1)
	}
}

func run(addr string) error {
	username := os.Getenv("ADMIN_USERNAME")
	password := os.Getenv("ADMIN_PASSWORD")
	secret := os.Getenv("JWT_SECRET")
	if username == "" || password == "" || secret == "" {
		return fmt.Errorf("нужны ADMIN_USERNAME, ADMIN_PASSWORD и JWT_SECRET в окружении")
	}
	ttl := 24 * time.Hour
	if raw := os.Getenv("JWT_TTL"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return fmt.Errorf("JWT_TTL: %w", err)
		}
		ttl = d
	}

	storeDir := os.Getenv("STORE_DIR")
	if storeDir == "" {
		storeDir = defaultStoreDir
	}
	st, err := store.NewFS(storeDir)
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}

	det, err := jplag.NewFromEnv(st)
	if err != nil {
		return fmt.Errorf("jplag: %w", err)
	}

	mongoURI, err := repository.URIFromEnv()
	if err != nil {
		return err
	}
	mongoDB := os.Getenv("MONGODB_DATABASE")
	if mongoDB == "" {
		mongoDB = defaultMongoDatabase
	}
	mongoClient, db, err := repository.Connect(context.Background(), mongoURI, mongoDB)
	if err != nil {
		return fmt.Errorf("mongo: %w", err)
	}
	defer mongoClient.Disconnect(context.Background())

	registry := repository.NewContestRepository(db)
	findingsStore := repository.NewFindingsRepository(db)

	jobsMaxConcurrent, err := intEnv("JOBS_MAX_CONCURRENT", defaultJobsMaxConcurrent)
	if err != nil {
		return err
	}
	analyzeConcurrency, err := intEnv("ANALYZE_CONCURRENCY", defaultAnalyzeConcurrency)
	if err != nil {
		return err
	}
	limiter := detect.NewLimiter(analyzeConcurrency)
	// Один Limiter на процесс: иначе JOBS_MAX_CONCURRENT job'ов перемножили бы параллелизм JPlag/LLM.
	factories := []detect.StageFactory{
		func(id domain.ContestID) detect.Stage {
			return detect.NewStage(detect.ProblemSelector{Contest: id}, limiter, det)
		},
	}

	if os.Getenv(envAIUsageEnabled) == "1" {
		model, err := llm.NewOpenAIFromEnv()
		if err != nil {
			return fmt.Errorf("aiusage: %w", err)
		}
		analyzer := &aiusage.Analyzer{Model: model}
		taskDet := aiusage.NewTaskDetector(analyzer)
		factories = append(factories, func(id domain.ContestID) detect.Stage {
			return detect.NewStage(
				detect.OkWithLastSelector{Contest: id},
				limiter,
				taskDet,
			)
		})
		fmt.Fprintln(os.Stderr, "scainer: aiusage enabled")
	}

	pipeline := detect.Compose(factories...)
	pool := jobs.NewPool(jobsMaxConcurrent)
	scorer := scoring.NewWeighted()
	contestsSvc := contests.NewService(registry, findingsStore, defaultJudgeSystem)
	reader := contests.NewContestReader(registry, st, findingsStore)
	analyzeSvc := analyze.New(pool, analyze.NewRunner(registry, st, findingsStore, scorer, pipeline))

	authSvc := auth.New(username, password, secret, ttl)
	e := transport.New(contestsSvc, reader, analyzeSvc, authSvc).Echo()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		fmt.Fprintf(os.Stderr, "scainer: listening on %s\n", addr)
		errCh <- e.Start(addr)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

func intEnv(name string, defaultVal int) (int, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return defaultVal, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	return n, nil
}
