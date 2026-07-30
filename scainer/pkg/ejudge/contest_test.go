package ejudge_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"scainer/pkg/ejudge"
)

func TestContestStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/new-master" || r.URL.Query().Get("action") != "contest-status-json" {
			t.Fatalf("unexpected %s %s", r.URL.Path, r.URL.RawQuery)
		}
		if r.Header.Get("Authorization") == "" {
			t.Fatal("missing auth")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"ok":true,
			"result":{
				"contest":{"id":50601,"name":"День 01"},
				"problems":[
					{"id":1,"short_name":"A","internal_name":"find-cycle","name":"Find cycle"},
					{"id":2,"short_name":"B","name":"Task B"}
				]
			}
		}`))
	}))
	t.Cleanup(srv.Close)

	c, err := ejudge.New(srv.URL, "tok", 0)
	if err != nil {
		t.Fatal(err)
	}
	info, err := c.ContestStatus(context.Background(), 50601)
	if err != nil {
		t.Fatal(err)
	}
	if info.ID != 50601 || info.Name != "День 01" {
		t.Fatalf("contest: got %+v", info)
	}
	if len(info.Problems) != 2 {
		t.Fatalf("problems len: got %d", len(info.Problems))
	}
	if info.Problems[0].ProblemKey() != "find-cycle" || info.Problems[0].DisplayName() != "A" {
		t.Fatalf("problem[0]: %+v", info.Problems[0])
	}
	if info.Problems[1].ProblemKey() != "B" || info.Problems[1].DisplayName() != "B" {
		t.Fatalf("problem[1]: %+v", info.Problems[1])
	}
}

func TestProblemBriefProblemKeyFallback(t *testing.T) {
	tests := []struct {
		name string
		p    ejudge.ProblemBrief
		want string
	}{
		{"internal", ejudge.ProblemBrief{ID: 1, InternalName: "slug", ShortName: "A"}, "slug"},
		{"short", ejudge.ProblemBrief{ID: 1, ShortName: "A"}, "A"},
		{"id", ejudge.ProblemBrief{ID: 42}, "42"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.p.ProblemKey(); got != tc.want {
				t.Fatalf("ProblemKey: got %q want %q", got, tc.want)
			}
		})
	}
}
