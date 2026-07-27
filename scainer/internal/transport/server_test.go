package transport_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"scainer/internal/domain"
	"scainer/internal/services/analyze"
	"scainer/internal/services/contests"
	"scainer/internal/services/analyze/detect"
	"scainer/internal/services/ejudge/gateway"
	"scainer/internal/services/importer"
	"scainer/internal/services/review"
	"scainer/internal/services/scoring"
	"scainer/internal/services/teachers"
	"scainer/internal/transport"
	"scainer/pkg/auth"
	"scainer/pkg/jobs"
	"scainer/pkg/store"
)

func emptyOrchestrator() *analyze.Orchestrator {
	return analyze.NewOrchestrator(analyze.NewRegistry(detect.NewLimiter(4)))
}

type stubImporter struct{}

func (stubImporter) Name() string { return "stub" }

func (stubImporter) Import(context.Context, importer.Store) (importer.Result, error) {
	return importer.Result{}, nil
}

var _ importer.Importer = stubImporter{}

func init() {
	importer.Register("stub", func(node *yaml.Node) (importer.Importer, error) {
		return stubImporter{}, nil
	})
}

type fakeRegistry struct {
	byID map[domain.ContestID]contests.ContestRecord
}

func newFakeRegistry() *fakeRegistry {
	return &fakeRegistry{byID: make(map[domain.ContestID]contests.ContestRecord)}
}

func (r *fakeRegistry) Put(_ context.Context, rec contests.ContestRecord) error {
	r.byID[rec.Contest.ID] = rec
	return nil
}

func (r *fakeRegistry) Get(_ context.Context, id domain.ContestID) (contests.ContestRecord, bool, error) {
	rec, ok := r.byID[id]
	return rec, ok, nil
}

func (r *fakeRegistry) Delete(_ context.Context, id domain.ContestID) error {
	delete(r.byID, id)
	return nil
}

func (r *fakeRegistry) List(context.Context) ([]contests.ContestRecord, error) {
	out := make([]contests.ContestRecord, 0, len(r.byID))
	for _, rec := range r.byID {
		out = append(out, rec)
	}
	return out, nil
}

var _ contests.ContestRegistry = (*fakeRegistry)(nil)

type fakeAnalysisRepository struct {
	byID map[domain.ContestID]contests.AnalysisSnapshot
}

func newFakeAnalysisRepository() *fakeAnalysisRepository {
	return &fakeAnalysisRepository{byID: make(map[domain.ContestID]contests.AnalysisSnapshot)}
}

func (f *fakeAnalysisRepository) Put(_ context.Context, snap contests.AnalysisSnapshot) error {
	f.byID[snap.ContestID] = snap
	return nil
}

func (f *fakeAnalysisRepository) Get(_ context.Context, id domain.ContestID) (contests.AnalysisSnapshot, bool, error) {
	snap, ok := f.byID[id]
	return snap, ok, nil
}

func (f *fakeAnalysisRepository) Delete(_ context.Context, id domain.ContestID) error {
	delete(f.byID, id)
	return nil
}

var _ contests.AnalysisRepository = (*fakeAnalysisRepository)(nil)

type noopTeacherRepository struct{}

func (noopTeacherRepository) Get(context.Context, string) (teachers.Record, bool, error) {
	return teachers.Record{}, false, nil
}

func (noopTeacherRepository) Upsert(context.Context, string, string) error {
	return nil
}

func testTeachers() *teachers.Service {
	return teachers.NewService(noopTeacherRepository{}, nil)
}

func testAuth() auth.Service {
	return auth.New("jwt-secret", time.Hour)
}

func TestPostContestImportNotFound(t *testing.T) {
	authSvc := testAuth()
	token, _, err := authSvc.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}

	reg := newFakeRegistry()
	fs := newFakeAnalysisRepository()
	st := store.NewMem()
	orch := emptyOrchestrator()
	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, st, fs, scoring.NewWeighted())
	analyzeSvc := analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, orch))

	if _, err := svc.Register(context.Background(), contests.Registration{
		ID:     "contest01",
		Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	e := transport.New(svc, reader, analyzeSvc, nil, testTeachers(), authSvc, nil).Echo()

	req := httptest.NewRequest(http.MethodPost, "/api/contests/missing/import", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want 404 body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["error"] != "contest not found" {
		t.Fatalf("error: got %q", body["error"])
	}
}

func TestPostContestImportSubmitsJob(t *testing.T) {
	authSvc := testAuth()
	token, _, err := authSvc.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}

	reg := newFakeRegistry()
	fs := newFakeAnalysisRepository()
	st := store.NewMem()
	orch := emptyOrchestrator()
	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, st, fs, scoring.NewWeighted())
	analyzeSvc := analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, orch))

	if _, err := svc.Register(context.Background(), contests.Registration{
		ID:     "contest01",
		Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	e := transport.New(svc, reader, analyzeSvc, nil, testTeachers(), authSvc, nil).Echo()

	req := httptest.NewRequest(http.MethodPost, "/api/contests/contest01/import", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status: got %d want 202 body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		JobID string `json:"jobId"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.JobID == "" {
		t.Fatal("ожидали непустой jobId")
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		statusReq := httptest.NewRequest(http.MethodGet, "/api/jobs/"+body.JobID, nil)
		statusReq.Header.Set("Authorization", "Bearer "+token)
		statusRec := httptest.NewRecorder()
		e.ServeHTTP(statusRec, statusReq)
		if statusRec.Code != http.StatusOK {
			t.Fatalf("GetJob status: got %d want 200 body=%s", statusRec.Code, statusRec.Body.String())
		}
		var st struct {
			Status   string `json:"status"`
			Progress struct {
				Phase string `json:"phase"`
			} `json:"progress"`
		}
		if err := json.Unmarshal(statusRec.Body.Bytes(), &st); err != nil {
			t.Fatal(err)
		}
		if st.Status == "succeeded" || st.Status == "failed" {
			if st.Progress.Phase == "analyzing" {
				t.Fatalf("import job must not enter analyzing phase, got %+v", st)
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("import job did not finish in time")
}

func TestPostContestSync(t *testing.T) {
	authSvc := testAuth()
	token, _, err := authSvc.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}

	reg := newFakeRegistry()
	fs := newFakeAnalysisRepository()
	st := store.NewMem()
	orch := emptyOrchestrator()
	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, st, fs, scoring.NewWeighted())
	analyzeSvc := analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, orch))

	if _, err := svc.Register(context.Background(), contests.Registration{
		ID:     "contest01",
		Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	e := transport.New(svc, reader, analyzeSvc, nil, testTeachers(), authSvc, nil).Echo()

	req := httptest.NewRequest(http.MethodPost, "/api/contests/contest01/sync", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status: got %d want 202 body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		JobID string `json:"jobId"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.JobID == "" {
		t.Fatal("ожидали непустой jobId")
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		statusReq := httptest.NewRequest(http.MethodGet, "/api/jobs/"+body.JobID, nil)
		statusReq.Header.Set("Authorization", "Bearer "+token)
		statusRec := httptest.NewRecorder()
		e.ServeHTTP(statusRec, statusReq)
		if statusRec.Code != http.StatusOK {
			t.Fatalf("GetJob status: got %d want 200 body=%s", statusRec.Code, statusRec.Body.String())
		}
		var st struct {
			Status   string `json:"status"`
			Progress struct {
				Phase string `json:"phase"`
			} `json:"progress"`
		}
		if err := json.Unmarshal(statusRec.Body.Bytes(), &st); err != nil {
			t.Fatal(err)
		}
		if st.Status == "succeeded" || st.Status == "failed" {
			if st.Status == "failed" {
				t.Fatalf("sync job failed: %+v", st)
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("sync job did not finish in time")
}

func TestPostContestAnalyzeWithoutImport(t *testing.T) {
	authSvc := testAuth()
	token, _, err := authSvc.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}

	reg := newFakeRegistry()
	fs := newFakeAnalysisRepository()
	st := store.NewMem()
	orch := emptyOrchestrator()
	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, st, fs, scoring.NewWeighted())
	analyzeSvc := analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, orch))

	if _, err := svc.Register(context.Background(), contests.Registration{
		ID:     "contest01",
		Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	e := transport.New(svc, reader, analyzeSvc, nil, testTeachers(), authSvc, nil).Echo()

	req := httptest.NewRequest(http.MethodPost, "/api/contests/contest01/analyze", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400 body=%s", rec.Code, rec.Body.String())
	}
}

func TestPostContestAnalyzeSubmitsJob(t *testing.T) {
	authSvc := testAuth()
	token, _, err := authSvc.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}

	reg := newFakeRegistry()
	fs := newFakeAnalysisRepository()
	st := store.NewMem()
	orch := emptyOrchestrator()
	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, st, fs, scoring.NewWeighted())
	analyzeSvc := analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, orch))

	ctx := context.Background()
	if _, err := svc.Register(ctx, contests.Registration{
		ID:     "contest01",
		Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}
	now := time.Now().UTC()
	rec, ok, err := reg.Get(ctx, "contest01")
	if err != nil || !ok {
		t.Fatalf("Get contest: %v ok=%v", err, ok)
	}
	rec.Contest.LastImportedAt = &now
	if err := reg.Put(ctx, rec); err != nil {
		t.Fatal(err)
	}

	e := transport.New(svc, reader, analyzeSvc, nil, testTeachers(), authSvc, nil).Echo()

	req := httptest.NewRequest(http.MethodPost, "/api/contests/contest01/analyze", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	httpRec := httptest.NewRecorder()
	e.ServeHTTP(httpRec, req)

	if httpRec.Code != http.StatusAccepted {
		t.Fatalf("status: got %d want 202 body=%s", httpRec.Code, httpRec.Body.String())
	}
	var body struct {
		JobID string `json:"jobId"`
	}
	if err := json.Unmarshal(httpRec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.JobID == "" {
		t.Fatal("ожидали непустой jobId")
	}
}

func TestGetJobNotFound(t *testing.T) {
	authSvc := testAuth()
	token, _, err := authSvc.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}

	reg := newFakeRegistry()
	fs := newFakeAnalysisRepository()
	st := store.NewMem()
	orch := emptyOrchestrator()
	e := transport.New(
		contests.NewService(reg, fs, "stub"),
		contests.NewContestReader(reg, st, fs, scoring.NewWeighted()),
		analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, orch)),
		nil,
		testTeachers(),
		authSvc,
		nil,
	).Echo()

	req := httptest.NewRequest(http.MethodGet, "/api/jobs/does-not-exist", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want 404 body=%s", rec.Code, rec.Body.String())
	}
}

type fakeReviewComments struct {
	list []review.Comment
}

func (f *fakeReviewComments) List(context.Context, domain.Submission) ([]review.Comment, error) {
	return f.list, nil
}

func (f *fakeReviewComments) Post(context.Context, domain.Submission, string) error { return nil }

type fakeReviewStatus struct{}

func (fakeReviewStatus) Sync(_ context.Context, sub domain.Submission) (domain.Submission, error) {
	return sub, nil
}

func (fakeReviewStatus) SetVerdict(_ context.Context, sub domain.Submission, v domain.Verdict) (domain.Submission, error) {
	out := sub
	out.Verdict = v
	return out, nil
}

func TestGetContestSubmissions(t *testing.T) {
	authSvc := testAuth()
	token, _, err := authSvc.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}

	reg := newFakeRegistry()
	fs := newFakeAnalysisRepository()
	st := store.NewMem()
	_ = st.Put(context.Background(), []domain.Submission{{
		ID: "ejudge:contest01:1", Contest: "contest01", Problem: "A", Participant: "alice",
		Lang: domain.LangCPP, SubmittedAt: time.Unix(10, 0), Verdict: domain.VerdictPR,
		Meta: map[string]any{"run_id": 1},
	}})
	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, st, fs, scoring.NewWeighted())
	orch := emptyOrchestrator()
	analyzeSvc := analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, orch))
	reviewSvc := review.New(st, &fakeReviewComments{}, fakeReviewStatus{})

	if _, err := svc.Register(context.Background(), contests.Registration{
		ID: "contest01", Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatal(err)
	}

	e := transport.New(svc, reader, analyzeSvc, reviewSvc, testTeachers(), authSvc, nil).Echo()
	req := httptest.NewRequest(http.MethodGet, "/api/contests/contest01/submissions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var items []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0]["id"] != "ejudge:contest01:1" {
		t.Fatalf("%v", items)
	}
}

func TestGetSubmissionComments(t *testing.T) {
	authSvc := testAuth()
	token, _, err := authSvc.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}

	reg := newFakeRegistry()
	fs := newFakeAnalysisRepository()
	st := store.NewMem()
	_ = st.Put(context.Background(), []domain.Submission{{
		ID: "ejudge:contest01:1", Contest: "contest01", Problem: "A", Participant: "alice",
		Verdict: domain.VerdictPR, Meta: map[string]any{"run_id": 1},
	}})
	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, st, fs, scoring.NewWeighted())
	orch := emptyOrchestrator()
	analyzeSvc := analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, orch))
	reviewSvc := review.New(st, &fakeReviewComments{list: []review.Comment{{ID: "9", Text: "hi", Time: time.Unix(1, 0)}}}, fakeReviewStatus{})

	if _, err := svc.Register(context.Background(), contests.Registration{
		ID: "contest01", Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatal(err)
	}

	e := transport.New(svc, reader, analyzeSvc, reviewSvc, testTeachers(), authSvc, nil).Echo()
	req := httptest.NewRequest(http.MethodGet, "/api/contests/contest01/submissions/ejudge:contest01:1/comments", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["verdict"] != "PR" || body["status_stale"] != false {
		t.Fatalf("%v", body)
	}
	comments, _ := body["comments"].([]any)
	if len(comments) != 1 {
		t.Fatalf("comments=%v", comments)
	}
}

type stubEjudgeGateway struct {
	browserLogin gateway.BrowserLogin
	ok           bool
	err          error
}

func (s stubEjudgeGateway) BrowserLogin(_ context.Context, contestID int) (gateway.BrowserLogin, bool, error) {
	if contestID <= 0 {
		return gateway.BrowserLogin{}, false, nil
	}
	login := s.browserLogin
	if login.ContestID == 0 {
		login.ContestID = contestID
	}
	return login, s.ok, s.err
}

func (stubEjudgeGateway) EnsureAPIKey(context.Context) error { return nil }

func TestGetAuthMe_UsernameOnly(t *testing.T) {
	authSvc := testAuth()
	token, _, err := authSvc.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}

	reg := newFakeRegistry()
	fs := newFakeAnalysisRepository()
	st := store.NewMem()
	orch := emptyOrchestrator()
	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, st, fs, scoring.NewWeighted())
	analyzeSvc := analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, orch))
	e := transport.New(svc, reader, analyzeSvc, nil, testTeachers(), authSvc, nil).Echo()

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["username"] != "admin" {
		t.Fatalf("username=%v", body["username"])
	}
	if _, ok := body["ejudge"]; ok {
		t.Fatalf("ejudge should be absent, got %v", body["ejudge"])
	}
}

func TestGetContestEjudgeLogin_OK(t *testing.T) {
	authSvc := testAuth()
	token, _, err := authSvc.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}

	reg := newFakeRegistry()
	fs := newFakeAnalysisRepository()
	st := store.NewMem()
	orch := emptyOrchestrator()
	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, st, fs, scoring.NewWeighted())
	analyzeSvc := analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, orch))
	e := transport.New(svc, reader, analyzeSvc, nil, testTeachers(), authSvc, stubEjudgeGateway{
		browserLogin: gateway.BrowserLogin{
			BaseURL:   "https://ejudge.example",
			Login:     "admin",
			Password:  "secret",
			ContestID: 50506,
		},
		ok: true,
	}).Echo()

	req := httptest.NewRequest(http.MethodGet, "/api/contests/50506/ejudge-login", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["base_url"] != "https://ejudge.example" {
		t.Fatalf("body=%v", body)
	}
	if body["login"] != "admin" || body["password"] != "secret" || body["contest_id"] != float64(50506) {
		t.Fatalf("browser login fields=%v", body)
	}
}
