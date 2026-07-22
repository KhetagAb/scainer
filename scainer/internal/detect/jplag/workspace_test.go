package jplag

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"scainer/internal/domain"
	"scainer/internal/store"
	pkgfs "scainer/pkg/fs"
)

func TestPrepareSubmissions_SymlinkFromFS(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	fs, err := store.NewFS(root)
	if err != nil {
		t.Fatal(err)
	}

	subs := []domain.Submission{
		{ID: "ejudge:1:1", Participant: "alice", Problem: "A", Contest: "1", Lang: domain.LangCPP, Source: []byte("src-a"), Verdict: domain.VerdictOK},
		{ID: "ejudge:1:2", Participant: "bob", Problem: "A", Contest: "1", Lang: domain.LangCPP, Source: []byte("src-b"), Verdict: domain.VerdictOK},
	}
	if err := fs.Put(ctx, subs); err != nil {
		t.Fatal(err)
	}

	workDir := t.TempDir()
	byDir, err := prepareSubmissions(workDir, subs, ".cpp", fs)
	if err != nil {
		t.Fatal(err)
	}
	if len(byDir) != 2 {
		t.Fatalf("byDir = %d", len(byDir))
	}

	link := filepath.Join(workDir, "submissions", pkgfs.SanitizeFileName(string(subs[0].ID)), "main.cpp")
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s должен быть symlink, mode=%v", link, info.Mode())
	}
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatal(err)
	}
	srcPath, ok := fs.SourcePath(subs[0])
	if !ok {
		t.Fatal("SourcePath")
	}
	abs, err := filepath.Abs(srcPath)
	if err != nil {
		t.Fatal(err)
	}
	if target != abs {
		t.Fatalf("symlink target = %q, want %q", target, abs)
	}
	raw, err := os.ReadFile(link)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "src-a" {
		t.Fatalf("через symlink прочитали %q", raw)
	}
}

func TestPrepareSubmissions_WriteFileWithoutSources(t *testing.T) {
	workDir := t.TempDir()
	subs := []domain.Submission{
		{ID: "a", Participant: "alice", Problem: "A", Contest: "1", Lang: domain.LangPython, Source: []byte("print(1)")},
	}
	if _, err := prepareSubmissions(workDir, subs, ".py", nil); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(workDir, "submissions", "a", "main.py")
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("без Sources ожидается обычный файл, не symlink")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "print(1)" {
		t.Fatalf("got %q", raw)
	}
}

func TestNew_SetsSourcesAndWorkRoot(t *testing.T) {
	root := t.TempDir()
	fs, err := store.NewFS(root)
	if err != nil {
		t.Fatal(err)
	}
	d, err := New("bin/jplag.jar", fs)
	if err != nil {
		t.Fatal(err)
	}
	if d.Sources == nil {
		t.Fatal("Sources")
	}
	want := filepath.Join(root, ".jplag")
	if d.WorkRoot != want {
		t.Fatalf("WorkRoot = %q, want %q", d.WorkRoot, want)
	}
}

func TestPrepareWorkDir_UnderWorkRoot(t *testing.T) {
	fs, err := store.NewFS(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	d, err := New("bin/jplag.jar", fs)
	if err != nil {
		t.Fatal(err)
	}
	u := domain.ProblemUnit{Problem: "A", Lang: domain.LangCPP}
	dir, cleanup, err := d.prepareWorkDir(u, "50501")
	if err != nil {
		t.Fatal(err)
	}
	if cleanup == nil {
		t.Fatal("ожидается cleanup для run-dir")
	}
	defer cleanup()

	prefix := filepath.Join(d.WorkRoot, "50501", "A", "cpp") + string(filepath.Separator)
	if len(dir) <= len(prefix) || dir[:len(prefix)] != prefix {
		t.Fatalf("workDir = %q, want under %q", dir, prefix)
	}
	base := filepath.Base(dir)
	if len(base) < 4 || base[:4] != "run-" {
		t.Fatalf("run dir name = %q, want run-*", base)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatal(err)
	}

	dir2, cleanup2, err := d.prepareWorkDir(u, "50501")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup2()
	if dir == dir2 {
		t.Fatal("параллельные/повторные prepareWorkDir должны давать разные run-dir")
	}
}

func TestPrepareWorkDir_CleanupRemovesRunDir(t *testing.T) {
	fs, err := store.NewFS(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	d, err := New("bin/jplag.jar", fs)
	if err != nil {
		t.Fatal(err)
	}
	dir, cleanup, err := d.prepareWorkDir(domain.ProblemUnit{Problem: "A", Lang: domain.LangCPP}, "50501")
	if err != nil {
		t.Fatal(err)
	}
	cleanup()
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("после cleanup run-dir должен исчезнуть, err=%v", err)
	}
}
