package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"

	"scainer/internal/configs"
	"scainer/internal/cron"
	"scainer/internal/repository"
	"scainer/internal/services/analyze"
	"scainer/internal/services/contests"
	"scainer/internal/services/explain"
	"scainer/internal/services/ejudge"
	ejgateway "scainer/internal/services/ejudge/gateway"
	"scainer/internal/services/importer"
	"scainer/internal/services/refresh"
	"scainer/internal/services/review"
	"scainer/internal/services/scoring"
	"scainer/internal/services/statements"
	"scainer/internal/services/teachers"
	"scainer/internal/transport"
	"scainer/pkg/auth"
	"scainer/pkg/jobs"
	"scainer/pkg/llm"
	"scainer/pkg/lksh"
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

	if err := seedTeacherLogins(context.Background(), teacherLoginsToSeed(cfg), mongo.teachers); err != nil {
		return fmt.Errorf("teachers: %w", err)
	}

	var ejAuth teachers.EjudgeAuth
	if cfg.Ejudge.Enabled() {
		ejAuth = teachers.EjudgeAuth{
			BaseURL: cfg.Ejudge.BaseURL,
			Timeout: cfg.Ejudge.Timeout,
		}
	}
	teachersSvc := teachers.NewService(mongo.teachers, ejAuth)
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

	runner := analyze.NewRunner(mongo.registry, store, mongo.analysisRepo, orchestrator)
	statementsSvc := wireStatements(cfg, mongo.registry)
	explainSvc := wireExplain(statementsSvc, cfg, store)
	refreshOrch := analyze.NewRefreshOrchestrator(runner, statementsSvc)
	svcs := wireServices(cfg, store, mongo, ejGateway, teachersSvc, refreshOrch, pool)

	cronCtx, stopCron := context.WithCancel(context.Background())
	defer stopCron()
	if cfg.ImportCron.Enabled {
		masterLogin := cfg.ImportCron.MasterLogin
		if cfg.Ejudge.Enabled() {
			warmCtx, cancel := context.WithTimeout(auth.WithLogin(context.Background(), masterLogin), 30*time.Second)
			if err := ejGateway.EnsureAPIKey(warmCtx); err != nil {
				fmt.Fprintf(os.Stderr, "scainer: import_cron ejudge warmup for masterlogin %s: %v\n", masterLogin, err)
			}
			cancel()
		}
		cancel, err := cron.StartImportCron(cronCtx, mongo.registry, svcs.analyze, cfg.ImportCron.Interval, masterLogin)
		if err != nil {
			return fmt.Errorf("import_cron: %w", err)
		}
		defer cancel()
		fmt.Fprintf(os.Stderr, "scainer: import_cron every %s as masterlogin %s\n", cfg.ImportCron.Interval, masterLogin)
	}

	var ejudgeGw *ejgateway.Gateway
	if cfg.Ejudge.Enabled() {
		ejudgeGw = ejGateway
	}
	e := transport.New(svcs.contests, svcs.reader, svcs.analyze, svcs.review, svcs.teachers, svcs.auth, ejudgeGw, statementsSvc, explainSvc).Echo()
	return serveHTTP(cfg.HTTP, e)
}

func teacherLoginsToSeed(cfg *configs.Config) []string {
	logins := append([]string(nil), cfg.TeachersLogins...)
	ml := strings.TrimSpace(cfg.ImportCron.MasterLogin)
	if ml == "" {
		return logins
	}
	for _, l := range logins {
		if l == ml {
			return logins
		}
	}
	return append(logins, ml)
}

type mongoDeps struct {
	client       *mongo.Client
	teachers     *repository.TeachersRepository
	credentials  *repository.EjudgeCredentialsRepository
	registry     contests.ContestRegistry
	analysisRepo contests.AnalysisRepository
}

func seedTeacherLogins(ctx context.Context, logins []string, repo *repository.TeachersRepository) error {
	if len(logins) == 0 {
		return nil
	}
	for _, login := range logins {
		if err := repo.EnsureLogin(ctx, login); err != nil {
			return fmt.Errorf("%s: %w", login, err)
		}
	}
	fmt.Fprintf(os.Stderr, "scainer: ensured %d teacher login(s) from TEACHERS_LOGINS\n", len(logins))
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
	refreshOrch *refresh.Orchestrator,
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
		analyze:  analyze.New(pool, refreshOrch, mongo.registry),
		review:   review.New(store, comments, status),
		teachers: teachersSvc,
		auth:     auth.New(cfg.Admin.JWTSecret, cfg.Admin.JWTTTL),
	}
}

func wireStatements(cfg *configs.Config, registry contests.ContestRegistry) *statements.Service {
	lkshClient := lksh.NewClient(cfg.Ejudge.BaseURL, cfg.Ejudge.Timeout)
	providers := map[string]statements.Provider{
		statements.SourceLksh: ejudge.NewLkshStatementProvider(lkshClient),
	}
	problems := statements.NewProblemStore(cfg.Store.Dir)
	return statements.NewService(registry, providers, problems)
}

func wireExplain(statementsSvc *statements.Service, cfg *configs.Config, submissionStore *store.FS) *explain.Service {
	problems := statements.NewProblemStore(cfg.Store.Dir)
	var model llm.Model = llm.Unconfigured{}
	if m, err := llm.NewOpenAIFromEnv(); err == nil {
		model = m
	} else {
		fmt.Fprintf(os.Stderr, "scainer: llm disabled: %v\n", err)
	}
	return explain.New(
		statementsSvc,
		problems,
		explain.NewSubmissionLabelResolver(submissionStore),
		model,
	)
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
