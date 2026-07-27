package selectors_test

import (
	"context"
	"testing"
	"time"

	"scainer/internal/domain"
	"scainer/internal/services/analyze/selectors"
	"scainer/pkg/store"
)

func TestSubmissionsSelector(t *testing.T) {
	ctx := context.Background()
	st := store.NewMem()
	_ = st.Put(ctx, []domain.Submission{
		{ID: "1", Contest: "c", Problem: "B", Participant: "bob", SubmittedAt: time.Unix(2, 0)},
		{ID: "2", Contest: "c", Problem: "A", Participant: "alice", SubmittedAt: time.Unix(1, 0)},
		{ID: "3", Contest: "c", Problem: "A", Participant: "carol", SubmittedAt: time.Unix(3, 0)},
	})

	units, err := selectors.Submissions("c")(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 3 {
		t.Fatalf("units: got %d, want 3", len(units))
	}
	if units[0].Sub.ID != "2" || units[1].Sub.ID != "3" || units[2].Sub.ID != "1" {
		t.Fatalf("order: %#v", units)
	}
}
