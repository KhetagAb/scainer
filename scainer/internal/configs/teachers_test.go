package configs_test

import (
	"testing"

	"scainer/internal/configs"
)

func TestParseTeachersLogins(t *testing.T) {
	got, err := configs.ParseTeachersLogins("alice;bob")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "alice" || got[1] != "bob" {
		t.Fatalf("got=%v", got)
	}
}

func TestParseTeachersLogins_TrimsSpaces(t *testing.T) {
	got, err := configs.ParseTeachersLogins(" alice ; bob ")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "alice" || got[1] != "bob" {
		t.Fatalf("got=%v", got)
	}
}

func TestParseTeachersLogins_Empty(t *testing.T) {
	got, err := configs.ParseTeachersLogins("")
	if err != nil || got != nil {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestParseTeachersLogins_Invalid(t *testing.T) {
	cases := []string{
		";",
		"alice;alice",
	}
	for _, raw := range cases {
		if _, err := configs.ParseTeachersLogins(raw); err == nil {
			t.Fatalf("expected error for %q", raw)
		}
	}
}
