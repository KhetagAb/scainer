package auth_test

import (
	"testing"
	"time"

	"github.com/lksh/scainer/pkg/auth"
)

func TestValidateCredentials(t *testing.T) {
	svc := auth.New("admin", "secret", "jwt-secret", time.Hour)

	if !svc.ValidateCredentials("admin", "secret") {
		t.Fatal("ожидали успех для верных кредов")
	}
	if svc.ValidateCredentials("admin", "wrong") {
		t.Fatal("неверный пароль должен отклоняться")
	}
	if svc.ValidateCredentials("other", "secret") {
		t.Fatal("неверный username должен отклоняться")
	}
	if svc.ValidateCredentials("adminx", "secret") {
		t.Fatal("частичное совпадение username недопустимо")
	}
}

func TestIssueParseRoundtrip(t *testing.T) {
	svc := auth.New("admin", "secret", "jwt-secret", time.Hour)
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
	svc := auth.New("admin", "secret", "jwt-secret", time.Millisecond)
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
	svc := auth.New("admin", "secret", "jwt-secret", time.Hour)
	if _, err := svc.ParseToken("not-a-jwt"); err == nil {
		t.Fatal("ожидали ошибку")
	}
	if _, err := svc.ParseToken(""); err == nil {
		t.Fatal("ожидали ошибку для пустого токена")
	}

	other := auth.New("admin", "secret", "other-secret", time.Hour)
	token, _, err := other.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ParseToken(token); err == nil {
		t.Fatal("токен с другим секретом должен отклоняться")
	}
}
