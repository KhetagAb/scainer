package lksh_test

import (
	"os"
	"path/filepath"
	"testing"

	"scainer/pkg/lksh"
)

func TestExtractRawConfig_parallel5(t *testing.T) {
	html := readFixture(t, "parallel_5.html")
	cfg, err := lksh.ExtractRawConfig(html)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaseURL != "/5/" {
		t.Fatalf("base_url: got %q want /5/", cfg.BaseURL)
	}
	lessons := lksh.CollectLessons(cfg.Layout)
	if len(lessons) < 10 {
		t.Fatalf("lessons: got %d want >= 10", len(lessons))
	}
}

func TestFindLessonByContestID_parallel5(t *testing.T) {
	html := readFixture(t, "parallel_5.html")
	cfg, err := lksh.ExtractRawConfig(html)
	if err != nil {
		t.Fatal(err)
	}
	lessons := lksh.CollectLessons(cfg.Layout)
	lesson, err := lksh.FindLessonByContestID(lessons, "50501")
	if err != nil {
		t.Fatal(err)
	}
	if lesson.StatementsURL != "statements/01.pdf" {
		t.Fatalf("statements_url: got %q", lesson.StatementsURL)
	}
}

func TestStatementPDFURLFromPage_parallel5(t *testing.T) {
	html := readFixture(t, "parallel_5.html")
	got, err := lksh.StatementPDFURLFromPage("https://ejudge.lksh.ru", "5", "50501", html)
	if err != nil {
		t.Fatal(err)
	}
	want := "https://ejudge.lksh.ru/5/statements/01.pdf"
	if got != want {
		t.Fatalf("pdf url: got %q want %q", got, want)
	}
}

func TestBuildStatementPDFURL_letterParallel(t *testing.T) {
	lesson := lksh.Lesson{StatementsURL: "statements/09_2.pdf"}
	got, err := lksh.BuildStatementPDFURL("https://ejudge.lksh.ru", "X", "/x/", lesson)
	if err != nil {
		t.Fatal(err)
	}
	want := "https://ejudge.lksh.ru/x/statements/09_2.pdf"
	if got != want {
		t.Fatalf("pdf url: got %q want %q", got, want)
	}
}

func TestParseEjudgeContestID(t *testing.T) {
	t.Parallel()
	cases := []struct {
		raw  string
		want string
		ok   bool
	}{
		{"//ejudge.lksh.ru/cgi-bin/new-client?contest_id=50501&amp;locale_id=1", "50501", true},
		{"https://codeforces.com/group/fEGCJZSa7c", "", false},
		{"", "", false},
	}
	for _, tc := range cases {
		got, ok := lksh.ParseEjudgeContestID(tc.raw)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("ParseEjudgeContestID(%q) = %q,%v want %q,%v", tc.raw, got, ok, tc.want, tc.ok)
		}
	}
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("testdata", name)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
