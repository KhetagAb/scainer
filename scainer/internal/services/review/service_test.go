package review_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"scainer/internal/domain"
	"scainer/internal/services/review"
	"scainer/pkg/store"
)

type fakeComments struct {
	list  []review.Comment
	err   error
	posts []string
}

func (f *fakeComments) List(_ context.Context, _ domain.Submission) ([]review.Comment, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.list, nil
}

func (f *fakeComments) Post(_ context.Context, _ domain.Submission, text string) error {
	if f.err != nil {
		return f.err
	}
	f.posts = append(f.posts, text)
	return nil
}

type fakeStatus struct {
	verdict domain.Verdict
	syncErr error
	setErr  error
	sets    []domain.Verdict
}

func (f *fakeStatus) Sync(_ context.Context, sub domain.Submission) (domain.Submission, error) {
	if f.syncErr != nil {
		return domain.Submission{}, f.syncErr
	}
	out := sub
	out.Verdict = f.verdict
	return out, nil
}

func (f *fakeStatus) SetVerdict(_ context.Context, sub domain.Submission, v domain.Verdict) (domain.Submission, error) {
	if f.setErr != nil {
		return domain.Submission{}, f.setErr
	}
	f.sets = append(f.sets, v)
	out := sub
	out.Verdict = v
	return out, nil
}

func putSub(t *testing.T, st *store.Mem, verdict domain.Verdict) domain.Submission {
	t.Helper()
	sub := domain.Submission{
		ID:          "ejudge:42:7",
		Contest:     "42",
		Problem:     "A",
		Participant: "alice",
		Lang:        domain.LangCPP,
		Source:      []byte("int main(){}"),
		SubmittedAt: time.Unix(100, 0),
		Verdict:     verdict,
		Meta:        map[string]any{"run_id": 7, "contest_id": 42},
	}
	if err := st.Put(context.Background(), []domain.Submission{sub}); err != nil {
		t.Fatal(err)
	}
	return sub
}

func TestList(t *testing.T) {
	st := store.NewMem()
	_ = putSub(t, st, domain.VerdictPR)
	_ = st.Put(context.Background(), []domain.Submission{{
		ID: "ejudge:42:8", Contest: "42", Problem: "B", Participant: "bob",
		SubmittedAt: time.Unix(50, 0), Verdict: domain.VerdictOK,
		Meta: map[string]any{"run_id": 8},
	}})
	svc := review.New(st, &fakeComments{}, &fakeStatus{})
	items, err := svc.List(context.Background(), "42")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len=%d", len(items))
	}
	if items[0].ID != "ejudge:42:8" || items[1].ID != "ejudge:42:7" {
		t.Fatalf("order=%v %v", items[0].ID, items[1].ID)
	}
}

func TestLoadComments_SyncAndList(t *testing.T) {
	st := store.NewMem()
	_ = putSub(t, st, domain.VerdictPR)
	comments := &fakeComments{list: []review.Comment{{ID: "1", Text: "hi"}}}
	status := &fakeStatus{verdict: domain.VerdictOK}
	svc := review.New(st, comments, status)

	res, err := svc.LoadComments(context.Background(), "ejudge:42:7")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusStale || res.CommentsError != "" || res.Verdict != domain.VerdictOK || len(res.Comments) != 1 {
		t.Fatalf("%+v", res)
	}
	got, _ := st.GetByID(context.Background(), "ejudge:42:7")
	if got.Verdict != domain.VerdictOK {
		t.Fatalf("cache verdict=%s", got.Verdict)
	}
}

func TestLoadComments_PartialErrors(t *testing.T) {
	st := store.NewMem()
	_ = putSub(t, st, domain.VerdictPR)
	svc := review.New(st, &fakeComments{err: errors.New("msgs down")}, &fakeStatus{syncErr: errors.New("sync down")})
	res, err := svc.LoadComments(context.Background(), "ejudge:42:7")
	if err != nil {
		t.Fatal(err)
	}
	if !res.StatusStale || res.StatusError == "" || res.CommentsError == "" || res.Verdict != domain.VerdictPR {
		t.Fatalf("%+v", res)
	}
}

func TestDecide_CommentThenStatus(t *testing.T) {
	st := store.NewMem()
	_ = putSub(t, st, domain.VerdictPR)
	comments := &fakeComments{}
	status := &fakeStatus{}
	svc := review.New(st, comments, status)

	if err := svc.Decide(context.Background(), "ejudge:42:7", review.DecideRequest{
		Verdict: domain.VerdictRJ,
		Comment: "fix",
	}); err != nil {
		t.Fatal(err)
	}
	if len(comments.posts) != 1 || comments.posts[0] != "fix" {
		t.Fatalf("posts=%v", comments.posts)
	}
	if len(status.sets) != 1 || status.sets[0] != domain.VerdictRJ {
		t.Fatalf("sets=%v", status.sets)
	}
	got, _ := st.GetByID(context.Background(), "ejudge:42:7")
	if got.Verdict != domain.VerdictRJ {
		t.Fatalf("cache=%s", got.Verdict)
	}
}

func TestDecide_StatusFailDoesNotUpdateCache(t *testing.T) {
	st := store.NewMem()
	_ = putSub(t, st, domain.VerdictPR)
	comments := &fakeComments{}
	status := &fakeStatus{setErr: errors.New("write failed")}
	svc := review.New(st, comments, status)

	err := svc.Decide(context.Background(), "ejudge:42:7", review.DecideRequest{
		Verdict: domain.VerdictOK,
		Comment: "ok",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if len(comments.posts) != 1 {
		t.Fatal("comment should be sent before status")
	}
	got, _ := st.GetByID(context.Background(), "ejudge:42:7")
	if got.Verdict != domain.VerdictPR {
		t.Fatalf("cache must stay PR, got %s", got.Verdict)
	}
}

func TestDecide_CommentFailStops(t *testing.T) {
	st := store.NewMem()
	_ = putSub(t, st, domain.VerdictPR)
	comments := &fakeComments{err: errors.New("post fail")}
	status := &fakeStatus{}
	svc := review.New(st, comments, status)

	err := svc.Decide(context.Background(), "ejudge:42:7", review.DecideRequest{
		Verdict: domain.VerdictOK,
		Comment: "x",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if len(status.sets) != 0 {
		t.Fatal("status must not be set")
	}
}
