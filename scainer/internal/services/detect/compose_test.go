package detect_test

import (
	"testing"

	"scainer/internal/services/detect"
	"scainer/internal/domain"
)

func TestComposeForContest(t *testing.T) {
	var calls []domain.ContestID
	p := detect.Compose(
		func(id domain.ContestID) detect.Stage {
			calls = append(calls, id)
			return nil
		},
		nil,
		func(id domain.ContestID) detect.Stage {
			calls = append(calls, id+"-2")
			return nil
		},
	)
	stages := p.ForContest("c1")
	if len(stages) != 2 {
		t.Fatalf("stages: got %d want 2", len(stages))
	}
	if len(calls) != 2 || calls[0] != "c1" || calls[1] != "c1-2" {
		t.Fatalf("calls: %v", calls)
	}
}
