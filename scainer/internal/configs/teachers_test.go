package configs_test

import (
	"testing"

	"scainer/internal/configs"
)

func TestParseTeachersPasswords(t *testing.T) {
	got, err := configs.ParseTeachersPasswords("alice:secret;bob:pass2")
	if err != nil {
		t.Fatal(err)
	}
	if got["alice"] != "secret" || got["bob"] != "pass2" {
		t.Fatalf("got=%v", got)
	}
}

func TestParseTeachersPasswords_PasswordWithColon(t *testing.T) {
	got, err := configs.ParseTeachersPasswords("alice:pa:ss:word")
	if err != nil {
		t.Fatal(err)
	}
	if got["alice"] != "pa:ss:word" {
		t.Fatalf("got=%q", got["alice"])
	}
}

func TestParseTeachersPasswords_Empty(t *testing.T) {
	got, err := configs.ParseTeachersPasswords("")
	if err != nil || got != nil {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestParseTeachersPasswords_Invalid(t *testing.T) {
	cases := []string{
		"alice",
		":secret",
		"alice:",
		"alice:secret;alice:other",
	}
	for _, raw := range cases {
		if _, err := configs.ParseTeachersPasswords(raw); err == nil {
			t.Fatalf("expected error for %q", raw)
		}
	}
}
