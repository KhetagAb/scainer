package teachers_test

import (
	"context"
	"errors"
	"testing"

	"scainer/internal/services/teachers"
)

type fakeRepository struct {
	rec       teachers.Record
	ok        bool
	err       error
	upsertErr error
	upserted  map[string]string
}

func (f *fakeRepository) Get(_ context.Context, login string) (teachers.Record, bool, error) {
	if f.err != nil {
		return teachers.Record{}, false, f.err
	}
	if !f.ok || f.rec.Login != login {
		return teachers.Record{}, false, nil
	}
	return f.rec, true, nil
}

func (f *fakeRepository) Upsert(_ context.Context, login, password string) error {
	if f.upsertErr != nil {
		return f.upsertErr
	}
	if f.upserted == nil {
		f.upserted = make(map[string]string)
	}
	f.upserted[login] = password
	return nil
}

type fakeEjudge struct {
	err error
}

func (f fakeEjudge) Validate(context.Context, string, string) error {
	return f.err
}

func TestLogin_OK(t *testing.T) {
	repo := &fakeRepository{
		ok:  true,
		rec: teachers.Record{Login: "alice", Password: ""},
	}
	svc := teachers.NewService(repo, fakeEjudge{})
	if err := svc.Login(context.Background(), "alice", "secret"); err != nil {
		t.Fatal(err)
	}
	if repo.upserted["alice"] != "secret" {
		t.Fatalf("upserted=%v", repo.upserted)
	}
}

func TestLogin_NotFound(t *testing.T) {
	svc := teachers.NewService(&fakeRepository{}, fakeEjudge{})
	err := svc.Login(context.Background(), "alice", "secret")
	if !errors.Is(err, teachers.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	svc := teachers.NewService(&fakeRepository{
		ok:  true,
		rec: teachers.Record{Login: "alice", Password: ""},
	}, fakeEjudge{err: errors.New("ejudge rejected")})
	err := svc.Login(context.Background(), "alice", "wrong")
	if !errors.Is(err, teachers.ErrInvalidCredentials) {
		t.Fatalf("err = %v", err)
	}
}

func TestLogin_NoEjudge(t *testing.T) {
	svc := teachers.NewService(&fakeRepository{
		ok:  true,
		rec: teachers.Record{Login: "alice", Password: ""},
	}, nil)
	err := svc.Login(context.Background(), "alice", "secret")
	if !errors.Is(err, teachers.ErrInvalidCredentials) {
		t.Fatalf("err = %v", err)
	}
}
