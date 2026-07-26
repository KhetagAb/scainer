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
	"scainer/internal/cron"
	"scainer/internal/repository"
	"scainer/internal/services/analyze"
	"scainer/internal/services/contests"
	"scainer/internal/services/ejudge"
	ejgateway "scainer/internal/services/ejudge/gateway"
	"scainer/internal/services/importer"
	"scainer/internal/services/review"
	"scainer/internal/services/scoring"
	"scainer/internal/services/teachers"
	"scainer/internal/transport"
	"scainer/pkg/auth"
	"scainer/pkg/jobs"
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

	teachersSvc := teachers.NewService(mongo.teachers, mongo.loginAudit)
	ejGateway := ejgateway.New(mongo.teachers, mongo.credentials, cfg.Ejudge.BaseURL, cfg.Ejudge.Timeout)
	if cfg.Ejudge.Enabled() {
		importer.Register("ejudge", ejudge.ImporterFactory(ejGateway))
		fmt.Fprintln(os.Stderr, "scainer: ejudge enabled")
	} else {
		fmt.Fprintln(os.Stderr, "scainer: ejudge disabled (нет base_url)")
	}

	orchestrator, pool, err := analyze.NewRuntime(cfg, store)
	if err != nil {
		return err
	}

	svcs := wireServices(cfg, store, mongo, ejGateway, teachersSvc, orchestrator, pool)

	cronCtx, stopCron := context.WithCancel(context.Background())
	defer stopCron()
	if cfg.ImportCron.Enabled {
		cancel, err := cron.StartImportCron(cronCtx, mongo.registry, svcs.analyze, cfg.ImportCron.Interval)
		if err != nil {
			return fmt.Errorf("import_cron: %w", err)
		}
		defer cancel()
		fmt.Fprintf(os.Stderr, "scainer: import_cron every %s\n", cfg.ImportCron.Interval)
	}

	var ejudgeGw *ejgateway.Gateway
	if cfg.Ejudge.Enabled() {
		ejudgeGw = ejGateway
	}
	e := transport.New(svcs.contests, svcs.reader, svcs.analyze, svcs.review, svcs.teachers, svcs.auth, ejudgeGw).Echo()
	return serveHTTP(cfg.HTTP, e)
}

type mongoDeps struct {
	client       *mongo.Client
	teachers     *repository.TeachersRepository
	credentials  *repository.EjudgeCredentialsRepository
	registry     contests.ContestRegistry
	analysisRepo contests.AnalysisRepository
	loginAudit   *repository.LoginAuditRepository
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
		client:       client,
		teachers:     repository.NewTeachersRepository(db),
		credentials:  repository.NewEjudgeCredentialsRepository(db),
		registry:     repository.NewContestRepository(db),
		analysisRepo: repository.NewAnalysisRepository(db),
		loginAudit:   repository.NewLoginAuditRepository(db),
	}, nil
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
	orchestrator *analyze.Orchestrator,
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
		contests: contests.NewService(mongo.registry, mongo.analysisRepo, defaultJudgeSystem),
		reader:   contests.NewContestReader(mongo.registry, store, mongo.analysisRepo, scorer),
		analyze:  analyze.New(pool, analyze.NewRunner(mongo.registry, store, mongo.analysisRepo, orchestrator)),
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
