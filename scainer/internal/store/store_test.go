package store

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"scainer/internal/domain"
)

func TestMemGet(t *testing.T) {
	ctx := context.Background()
	m := NewMem()
	_ = m.Put(ctx, []domain.Submission{
		{ID: "a", Participant: "alice", Problem: "A"},
		{ID: "b", Participant: "bob", Problem: "A"},
	})

	got, err := m.GetByIDs(ctx, []domain.SubmissionID{"a", "missing", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("ожидали 2 посылки, получили %d", len(got))
	}
	if got[0].ID != "a" || got[1].ID != "b" {
		t.Fatalf("неожиданный порядок: %v, %v", got[0].ID, got[1].ID)
	}

	_, err = m.GetByID(ctx, "missing")
	if !errors.Is(err, domain.ErrSubmissionNotFound) {
		t.Fatalf("ожидали ErrSubmissionNotFound, получили %v", err)
	}

	one, err := m.GetByID(ctx, "a")
	if err != nil {
		t.Fatal(err)
	}
	if one.ID != "a" {
		t.Fatalf("ID = %v", one.ID)
	}

	empty, err := m.GetByIDs(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 0 {
		t.Fatalf("пустой список ID: ожидали 0, получили %d", len(empty))
	}
}

func TestMem_ConcurrentPutAndRead(t *testing.T) {
	ctx := context.Background()
	m := NewMem()

	const writers = 4
	const readers = 4
	const opsPerGoroutine = 200

	var wg sync.WaitGroup
	wg.Add(writers + readers)

	for w := 0; w < writers; w++ {
		w := w
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				id := domain.SubmissionID(fmt.Sprintf("w%d-%d", w, i))
				_ = m.Put(ctx, []domain.Submission{{ID: id, Contest: "c", Problem: "A", Participant: "p"}})
				_ = m.SetCursor(ctx, "k", fmt.Sprintf("%d", i))
			}
		}()
	}
	for r := 0; r < readers; r++ {
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				_, _ = m.GetByID(ctx, "w0-0")
				_, _ = m.ByProblem(ctx, "c")
				_, _ = m.ByParticipant(ctx)
				_, _, _ = m.GetCursor(ctx, "k")
			}
		}()
	}
	wg.Wait()
}
