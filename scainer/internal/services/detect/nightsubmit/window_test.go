package nightsubmit_test

import (
	"testing"
	"time"

	"scainer/internal/services/detect/nightsubmit"
)

func mskTime(hour, min int) time.Time {
	return time.Date(2026, 7, 23, hour, min, 0, 0, time.FixedZone("MSK", 3*60*60))
}

func TestInNightWindow_Boundaries(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
		want bool
	}{
		{"22:29 outside", mskTime(22, 29), false},
		{"22:30 inclusive", mskTime(22, 30), true},
		{"23:59 inside", mskTime(23, 59), true},
		{"00:00 inside", mskTime(0, 0), true},
		{"06:30 inclusive", mskTime(6, 30), true},
		{"06:31 outside", mskTime(6, 31), false},
		{"12:00 outside", mskTime(12, 0), false},
		{"zero time", time.Time{}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := nightsubmit.InNightWindow(tc.t); got != tc.want {
				t.Fatalf("InNightWindow(%v) = %v, want %v", tc.t, got, tc.want)
			}
		})
	}
}

func TestWindowLabel(t *testing.T) {
	if got := nightsubmit.WindowLabel(); got != "22:30–06:30" {
		t.Fatalf("WindowLabel() = %q", got)
	}
}

func TestEvidenceDescription(t *testing.T) {
	got := nightsubmit.EvidenceDescription(mskTime(2, 14))
	want := "Отправлено в 02:14 МСК · ночной интервал 22:30–06:30"
	if got != want {
		t.Fatalf("EvidenceDescription() = %q, want %q", got, want)
	}
}

func TestInNightWindow_UTCInput(t *testing.T) {
	// 22:30 MSK = 19:30 UTC
	utc := time.Date(2026, 7, 23, 19, 30, 0, 0, time.UTC)
	if !nightsubmit.InNightWindow(utc) {
		t.Fatal("expected night window for 19:30 UTC (= 22:30 MSK)")
	}
}
