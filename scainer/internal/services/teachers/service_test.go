package teachers_test

import (
	"context"
	"errors"
	"testing"

	"scainer/internal/services/teachers"
)

type fakeRepository struct {
	rec teachers.Record
	ok  bool
	err error
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

func TestLogin_OK(t *testing.T) {
	svc := teachers.NewService(&fakeRepository{
		ok: true,
		rec: teachers.Record{Login: "alice", Password: "secret"},
	})
	if err := svc.Login(context.Background(), "alice", "secret"); err != nil {
		t.Fatal(err)
	}
}

func TestLogin_NotFound(t *testing.T) {
	svc := teachers.NewService(&fakeRepository{})
	err := svc.Login(context.Background(), "alice", "secret")
	if !errors.Is(err, teachers.ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	svc := teachers.NewService(&fakeRepository{
		ok: true,
		rec: teachers.Record{Login: "alice", Password: "secret"},
	})
	err := svc.Login(context.Background(), "alice", "wrong")
	if !errors.Is(err, teachers.ErrInvalidCredentials) {
		t.Fatalf("err = %v", err)
	}
}
