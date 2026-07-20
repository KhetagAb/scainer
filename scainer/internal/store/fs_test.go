package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"scainer/internal/domain"
)

func TestFSSourcePath(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	f, err := NewFS(root)
	if err != nil {
		t.Fatal(err)
	}
	sub := domain.Submission{
		ID:          "ejudge:50501:12",
		Participant: "alice",
		Problem:     "A",
		Contest:     "50501",
		Lang:        domain.LangCPP,
		Source:      []byte("int main(){}"),
	}
	if _, ok := f.SourcePath(sub); ok {
		t.Fatal("до Put SourcePath должен быть ok=false")
	}
	if err := f.Put(ctx, []domain.Submission{sub}); err != nil {
		t.Fatal(err)
	}
	path, ok := f.SourcePath(sub)
	if !ok {
		t.Fatal("после Put SourcePath должен найти файл")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "int main(){}" {
		t.Fatalf("содержимое source = %q", raw)
	}
	if filepath.Base(path) != "source.cpp" {
		t.Fatalf("basename = %q, want source.cpp", filepath.Base(path))
	}
	if f.Root() != root {
		t.Fatalf("Root = %q, want %q", f.Root(), root)
	}
}

func TestFSGet(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	f, err := NewFS(root)
	if err != nil {
		t.Fatal(err)
	}
	err = f.Put(ctx, []domain.Submission{
		{ID: "a", Participant: "alice", Problem: "A", Contest: "c1", Lang: domain.LangCPP, Source: []byte("int main(){}")},
		{ID: "b", Participant: "bob", Problem: "A", Contest: "c1", Lang: domain.LangPython, Source: []byte("print(1)")},
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := f.Get(ctx, []domain.SubmissionID{"a", "missing", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("ожидали 2 посылки, получили %d", len(got))
	}

	empty, err := f.Get(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 0 {
		t.Fatalf("пустой список ID: ожидали 0, получили %d", len(empty))
	}
}

func TestFSPersistsAcrossProcesses(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	f1, err := NewFS(root)
	if err != nil {
		t.Fatal(err)
	}
	subs := []domain.Submission{
		{ID: "ejudge:50501:12", Participant: "alice", Problem: "A", Contest: "50501", Lang: domain.LangCPP, Source: []byte("src-a")},
		{ID: "ejudge:50501:13", Participant: "bob", Problem: "A", Contest: "50501", Lang: domain.LangGo, Source: []byte("src-b")},
		{ID: "ejudge:50501:14", Participant: "alice", Problem: "B", Contest: "50501", Lang: domain.LangJava, Source: []byte("src-c")},
	}
	if err := f1.Put(ctx, subs); err != nil {
		t.Fatal(err)
	}
	if err := f1.SetCursor(ctx, "ejudge:cursor:50501", "14"); err != nil {
		t.Fatal(err)
	}

	f2, err := NewFS(root)
	if err != nil {
		t.Fatal(err)
	}

	got, err := f2.Get(ctx, []domain.SubmissionID{"ejudge:50501:12", "ejudge:50501:13", "ejudge:50501:14"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("ожидали 3 посылки после рестарта, получили %d", len(got))
	}
	bySrc := map[string]bool{}
	for _, s := range got {
		bySrc[string(s.Source)] = true
	}
	for _, want := range []string{"src-a", "src-b", "src-c"} {
		if !bySrc[want] {
			t.Fatalf("не нашли исходник %q после рестарта", want)
		}
	}

	byProblem, err := f2.ByProblem(ctx, "50501")
	if err != nil {
		t.Fatal(err)
	}
	if len(byProblem["A"]) != 2 || len(byProblem["B"]) != 1 {
		t.Fatalf("неожиданная группировка ByProblem: %+v", byProblem)
	}

	byParticipant, err := f2.ByParticipant(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(byParticipant["alice"]) != 2 || len(byParticipant["bob"]) != 1 {
		t.Fatalf("неожиданная группировка ByParticipant: %+v", byParticipant)
	}

	v, ok, err := f2.GetCursor(ctx, "ejudge:cursor:50501")
	if err != nil {
		t.Fatal(err)
	}
	if !ok || v != "14" {
		t.Fatalf("курсор не пережил рестарт: ok=%v v=%q", ok, v)
	}
}

func TestFSSanitizeIDNoCollision(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	f, err := NewFS(root)
	if err != nil {
		t.Fatal(err)
	}
	subs := []domain.Submission{
		{ID: "ejudge:1:1", Participant: "p1", Problem: "A", Contest: "1", Lang: domain.LangCPP, Source: []byte("one")},
		{ID: "c/A/p2", Participant: "p2", Problem: "A", Contest: "1", Lang: domain.LangCPP, Source: []byte("two")},
	}
	if err := f.Put(ctx, subs); err != nil {
		t.Fatal(err)
	}

	f2, err := NewFS(root)
	if err != nil {
		t.Fatal(err)
	}
	got, err := f2.Get(ctx, []domain.SubmissionID{"ejudge:1:1", "c/A/p2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("ожидали 2 посылки, получили %d", len(got))
	}
}

func TestFSPutNoTempFilesLeftOver(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	f, err := NewFS(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Put(ctx, []domain.Submission{
		{ID: "a", Participant: "alice", Problem: "A", Contest: "c1", Lang: domain.LangCPP, Source: []byte("src")},
	}); err != nil {
		t.Fatal(err)
	}
	if err := f.SetCursor(ctx, "k", "v"); err != nil {
		t.Fatal(err)
	}

	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".tmp" {
			t.Fatalf("остался временный файл: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestFS_ConcurrentPutAndRead(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	f, err := NewFS(root)
	if err != nil {
		t.Fatal(err)
	}

	const writers = 4
	const readers = 4
	const opsPerGoroutine = 50

	var wg sync.WaitGroup
	wg.Add(writers + readers)

	for w := 0; w < writers; w++ {
		w := w
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				id := domain.SubmissionID(fmt.Sprintf("w%d-%d", w, i))
				_ = f.Put(ctx, []domain.Submission{
					{ID: id, Contest: "c", Problem: "A", Participant: "p", Lang: domain.LangCPP, Source: []byte("x")},
				})
				_ = f.SetCursor(ctx, "k", fmt.Sprintf("%d", i))
			}
		}()
	}
	for r := 0; r < readers; r++ {
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				_, _ = f.Get(ctx, []domain.SubmissionID{"w0-0"})
				_, _ = f.ByProblem(ctx, "c")
				_, _ = f.ByParticipant(ctx)
				_, _, _ = f.GetCursor(ctx, "k")
			}
		}()
	}
	wg.Wait()
}
