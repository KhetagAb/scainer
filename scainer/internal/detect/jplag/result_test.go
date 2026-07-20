package jplag

import (
	"path/filepath"
	"testing"
)

func TestParseResult_SampleFixture(t *testing.T) {
	path := filepath.Join("testdata", "sample.jplag")
	ov, comps, err := parseResult(path)
	if err != nil {
		t.Fatal(err)
	}
	if ov.TotalComparisons != 1 {
		t.Fatalf("total_comparisons = %d", ov.TotalComparisons)
	}
	if len(ov.TopComparisons) != 1 {
		t.Fatalf("top_comparisons len = %d", len(ov.TopComparisons))
	}
	top := ov.TopComparisons[0]
	if top.FirstSubmission != "alice" || top.SecondSubmission != "bob" {
		t.Fatalf("pair = %q/%q", top.FirstSubmission, top.SecondSubmission)
	}
	if got := top.avgSimilarity(); got != 1.0 {
		t.Fatalf("AVG = %v", got)
	}

	name := comparisonFileName(*ov, "alice", "bob")
	if name != "alice-bob.json" {
		t.Fatalf("comparison file = %q", name)
	}
	c, ok := comps[name]
	if !ok {
		t.Fatalf("missing comparison %q in map %#v", name, comps)
	}
	if c.ID1 != "alice" || c.ID2 != "bob" {
		t.Fatalf("ids = %q/%q", c.ID1, c.ID2)
	}
	if len(c.Matches) == 0 {
		t.Fatal("expected matches")
	}
	m := c.Matches[0]
	if m.Start1 <= 0 || m.End1 < m.Start1 || m.Start2 <= 0 || m.End2 < m.Start2 {
		t.Fatalf("match lines = %+v", m)
	}
}
