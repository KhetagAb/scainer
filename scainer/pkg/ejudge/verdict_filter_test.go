package ejudge

import "testing"

// Коды из REVIEW_VERDICT_OPTIONS на фронте — должны парситься и мапиться из ejudge status.
var reviewFilterVerdicts = []struct {
	verdict string
	status  int
}{
	{verdict: "PR", status: 16},
	{verdict: "OK", status: 0},
	{verdict: "WA", status: 5},
	{verdict: "TL", status: 3},
	{verdict: "ML", status: 12},
	{verdict: "CE", status: 1},
	{verdict: "RJ", status: 17},
	{verdict: "CF", status: 6},
	{verdict: "DQ", status: 10},
}

func TestReviewFilterVerdictsParseFromStatusStr(t *testing.T) {
	for _, tc := range reviewFilterVerdicts {
		if got := ParseVerdict(tc.verdict); got != Verdict(tc.verdict) {
			t.Fatalf("ParseVerdict(%q) = %q, want %q", tc.verdict, got, tc.verdict)
		}
	}
}

func TestReviewFilterVerdictsFromStatus(t *testing.T) {
	for _, tc := range reviewFilterVerdicts {
		if got := VerdictFromStatus(tc.status); got != Verdict(tc.verdict) {
			t.Fatalf("VerdictFromStatus(%d) = %q, want %q", tc.status, got, tc.verdict)
		}
	}
}

func TestReviewFilterVerdictsStatusCodeRoundTrip(t *testing.T) {
	for _, tc := range reviewFilterVerdicts {
		code, ok := StatusCode(Verdict(tc.verdict))
		if !ok {
			// CE не имеет обратного StatusCode (только чтение из ejudge).
			if tc.verdict == "CE" {
				continue
			}
			t.Fatalf("StatusCode(%q) not supported", tc.verdict)
		}
		if code != tc.status {
			t.Fatalf("StatusCode(%q) = %d, want %d", tc.verdict, code, tc.status)
		}
	}
}

func TestStyleViolationMapsToRJ(t *testing.T) {
	if got := ParseVerdict("SV"); got != VerdictRJ {
		t.Fatalf("ParseVerdict(SV) = %q, want RJ", got)
	}
	if VerdictFromStatus(14) != VerdictRJ {
		t.Fatal("status 14 (SV) -> RJ")
	}
}
