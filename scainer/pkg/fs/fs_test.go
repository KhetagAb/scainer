package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomicNoLeftoverTemp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")

	if err := WriteFileAtomic(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Fatalf("got %q want %q", got, "hello")
	}

	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("временный файл не должен оставаться: err=%v", err)
	}
}

func TestWriteFileAtomicOverwrites(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")

	if err := WriteFileAtomic(path, []byte("first"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomic(path, []byte("second"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "second" {
		t.Fatalf("got %q want %q", got, "second")
	}
}

func TestSanitizeFileName(t *testing.T) {
	cases := map[string]string{
		"ejudge:50501:12":        "ejudge_50501_12",
		"c/A/p2":                 "c_A_p2",
		"plain":                  "plain",
		"a:b/c":                  "a_b_c",
	}
	for in, want := range cases {
		if got := SanitizeFileName(in); got != want {
			t.Errorf("SanitizeFileName(%q) = %q, want %q", in, got, want)
		}
	}
}
