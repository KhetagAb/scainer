package store

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/lksh/scainer/internal/domain"
)

func TestMemGet(t *testing.T) {
	ctx := context.Background()
	m := NewMem()
	_ = m.Put(ctx, []domain.Submission{
		{ID: "a", Participant: "alice", Problem: "A"},
		{ID: "b", Participant: "bob", Problem: "A"},
	})

	got, err := m.Get(ctx, []domain.SubmissionID{"a", "missing", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("ожидали 2 посылки, получили %d", len(got))
	}
	ids := map[domain.SubmissionID]bool{}
	for _, s := range got {
		ids[s.ID] = true
	}
	if !ids["a"] || !ids["b"] {
		t.Fatalf("неожиданные ID: %v", ids)
	}

	empty, err := m.Get(ctx, nil)
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
				_, _ = m.Get(ctx, []domain.SubmissionID{"w0-0"})
				_, _ = m.ByProblem(ctx, "c")
				_, _ = m.ByParticipant(ctx)
				_, _, _ = m.GetCursor(ctx, "k")
			}
		}()
	}
	wg.Wait()
}
