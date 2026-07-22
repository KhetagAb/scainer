package transport_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"scainer/internal/analyze"
	"scainer/internal/contests"
	"scainer/internal/detect"
	"scainer/internal/domain"
	"scainer/internal/importer"
	"scainer/pkg/jobs"
	"scainer/internal/review"
	"scainer/internal/scoring"
	"scainer/internal/store"
	"scainer/internal/transport"
	"scainer/pkg/auth"
)

func emptyPipeline() detect.Pipeline {
	return detect.Compose()
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

type fakeFindingsStore struct {
	byID map[domain.ContestID]contests.FindingsSnapshot
}

func newFakeFindingsStore() *fakeFindingsStore {
	return &fakeFindingsStore{byID: make(map[domain.ContestID]contests.FindingsSnapshot)}
}

func (f *fakeFindingsStore) Put(_ context.Context, snap contests.FindingsSnapshot) error {
	f.byID[snap.ContestID] = snap
	return nil
}

func (f *fakeFindingsStore) Get(_ context.Context, id domain.ContestID) (contests.FindingsSnapshot, bool, error) {
	snap, ok := f.byID[id]
	return snap, ok, nil
}

func (f *fakeFindingsStore) Delete(_ context.Context, id domain.ContestID) error {
	delete(f.byID, id)
	return nil
}

var _ contests.FindingsStore = (*fakeFindingsStore)(nil)

func TestPostContestImportNotFound(t *testing.T) {
	authSvc := auth.New("admin", "secret", "jwt-secret", time.Hour)
	token, _, err := authSvc.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}

	reg := newFakeRegistry()
	fs := newFakeFindingsStore()
	st := store.NewMem()
	pipeline := emptyPipeline()
	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, st, fs)
	analyzeSvc := analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, scoring.NewWeighted(), pipeline))

	if _, err := svc.Register(context.Background(), contests.Registration{
		ID:     "contest01",
		Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	e := transport.New(svc, reader, analyzeSvc, nil, authSvc).Echo()

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
	authSvc := auth.New("admin", "secret", "jwt-secret", time.Hour)
	token, _, err := authSvc.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}

	reg := newFakeRegistry()
	fs := newFakeFindingsStore()
	st := store.NewMem()
	pipeline := emptyPipeline()
	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, st, fs)
	analyzeSvc := analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, scoring.NewWeighted(), pipeline))

	if _, err := svc.Register(context.Background(), contests.Registration{
		ID:     "contest01",
		Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	e := transport.New(svc, reader, analyzeSvc, nil, authSvc).Echo()

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

	statusReq := httptest.NewRequest(http.MethodGet, "/api/jobs/"+body.JobID, nil)
	statusReq.Header.Set("Authorization", "Bearer "+token)
	statusRec := httptest.NewRecorder()
	e.ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusOK {
		t.Fatalf("GetJob status: got %d want 200 body=%s", statusRec.Code, statusRec.Body.String())
	}
}

func TestGetJobNotFound(t *testing.T) {
	authSvc := auth.New("admin", "secret", "jwt-secret", time.Hour)
	token, _, err := authSvc.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}

	reg := newFakeRegistry()
	fs := newFakeFindingsStore()
	st := store.NewMem()
	pipeline := emptyPipeline()
	e := transport.New(
		contests.NewService(reg, fs, "stub"),
		contests.NewContestReader(reg, st, fs),
		analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, scoring.NewWeighted(), pipeline)),
		nil,
		authSvc,
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
	authSvc := auth.New("admin", "secret", "jwt-secret", time.Hour)
	token, _, err := authSvc.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}

	reg := newFakeRegistry()
	fs := newFakeFindingsStore()
	st := store.NewMem()
	_ = st.Put(context.Background(), []domain.Submission{{
		ID: "ejudge:contest01:1", Contest: "contest01", Problem: "A", Participant: "alice",
		Lang: domain.LangCPP, SubmittedAt: time.Unix(10, 0), Verdict: domain.VerdictPR,
		Meta: map[string]any{"run_id": 1},
	}})
	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, st, fs)
	pipeline := emptyPipeline()
	analyzeSvc := analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, scoring.NewWeighted(), pipeline))
	reviewSvc := review.New(st, &fakeReviewComments{}, fakeReviewStatus{})

	if _, err := svc.Register(context.Background(), contests.Registration{
		ID: "contest01", Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatal(err)
	}

	e := transport.New(svc, reader, analyzeSvc, reviewSvc, authSvc).Echo()
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
	authSvc := auth.New("admin", "secret", "jwt-secret", time.Hour)
	token, _, err := authSvc.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}

	reg := newFakeRegistry()
	fs := newFakeFindingsStore()
	st := store.NewMem()
	_ = st.Put(context.Background(), []domain.Submission{{
		ID: "ejudge:contest01:1", Contest: "contest01", Problem: "A", Participant: "alice",
		Verdict: domain.VerdictPR, Meta: map[string]any{"run_id": 1},
	}})
	svc := contests.NewService(reg, fs, "stub")
	reader := contests.NewContestReader(reg, st, fs)
	pipeline := emptyPipeline()
	analyzeSvc := analyze.New(jobs.NewPool(4), analyze.NewRunner(reg, st, fs, scoring.NewWeighted(), pipeline))
	reviewSvc := review.New(st, &fakeReviewComments{list: []review.Comment{{ID: "9", Text: "hi", Time: time.Unix(1, 0)}}}, fakeReviewStatus{})

	if _, err := svc.Register(context.Background(), contests.Registration{
		ID: "contest01", Source: &contests.SourceSpec{Type: "stub"},
	}); err != nil {
		t.Fatal(err)
	}

	e := transport.New(svc, reader, analyzeSvc, reviewSvc, authSvc).Echo()
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
