package auth_test

import (
	"context"
	"testing"
	"time"

	"scainer/pkg/auth"
)

func TestWithLoginRoundtrip(t *testing.T) {
	ctx := auth.WithLogin(context.Background(), "alice")
	login, ok := auth.LoginFrom(ctx)
	if !ok || login != "alice" {
		t.Fatalf("login=%q ok=%v", login, ok)
	}
}

func TestIssueParseRoundtrip(t *testing.T) {
	svc := auth.New("jwt-secret", time.Hour)
	token, expiresIn, err := svc.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}
	if expiresIn != 3600 {
		t.Fatalf("expiresIn: got %d want 3600", expiresIn)
	}

	claims, err := svc.ParseToken("Bearer " + token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Username != "admin" {
		t.Fatalf("username: got %q", claims.Username)
	}

	claims, err = svc.ParseToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Username != "admin" {
		t.Fatalf("username without Bearer: got %q", claims.Username)
	}
}

func TestParseExpiredToken(t *testing.T) {
	svc := auth.New("jwt-secret", time.Millisecond)
	token, _, err := svc.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Millisecond)

	if _, err := svc.ParseToken(token); err == nil {
		t.Fatal("ожидали ошибку для истёкшего токена")
	}
}

func TestParseInvalidToken(t *testing.T) {
	svc := auth.New("jwt-secret", time.Hour)
	if _, err := svc.ParseToken("not-a-jwt"); err == nil {
		t.Fatal("ожидали ошибку")
	}
	if _, err := svc.ParseToken(""); err == nil {
		t.Fatal("ожидали ошибку для пустого токена")
	}

	other := auth.New("other-secret", time.Hour)
	token, _, err := other.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ParseToken(token); err == nil {
		t.Fatal("токен с другим секретом должен отклоняться")
	}
}
