package ejudge

import (
	"testing"
	"time"

	ejgen "github.com/lksh/scainer/generated/ejudge"
	"github.com/lksh/scainer/internal/domain"
	ejudgeapi "github.com/lksh/scainer/pkg/ejudge"
)

func TestMapToSubmission(t *testing.T) {
	runID, status, score := 7, 5, 0
	lang, statusStr := "g++", "WA"
	user, prob, uuid := "alice", "find-cycle", "abc-uuid"
	runTimeUs := int(time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC).UnixMicro())

	run := ejgen.Run{
		RunId:            &runID,
		UserLogin:        &user,
		ProbInternalName: &prob,
		LangName:         &lang,
		Status:           &status,
		StatusStr:        &statusStr,
		Score:            &score,
		RunTimeUs:        &runTimeUs,
		RunUuid:          &uuid,
	}
	env := &ejudgeapi.Env{LangMap: map[string]domain.Lang{"g++": domain.LangCPP}}

	probName := "A"
	run.ProbName = &probName
	sub, err := mapToSubmission(50501, run, []byte("int main(){}"), env, "День 01")
	if err != nil {
		t.Fatal(err)
	}
	if sub.ID != "ejudge:50501:7" {
		t.Fatalf("ID = %q", sub.ID)
	}
	if sub.Participant != "alice" || sub.Problem != "find-cycle" || sub.Contest != "50501" {
		t.Fatalf("ids = %+v", sub)
	}
	if sub.Lang != domain.LangCPP || sub.Verdict != domain.VerdictWA || string(sub.Source) != "int main(){}" {
		t.Fatalf("fields = lang=%q verdict=%q src=%q", sub.Lang, sub.Verdict, sub.Source)
	}
	if !sub.SubmittedAt.Equal(time.UnixMicro(int64(runTimeUs))) {
		t.Fatalf("SubmittedAt = %v", sub.SubmittedAt)
	}
	if sub.Meta["run_uuid"] != "abc-uuid" || sub.Meta["status"] != 5 {
		t.Fatalf("Meta = %#v", sub.Meta)
	}
	if sub.Meta["contest_name"] != "День 01" || sub.Meta["problem_name"] != "A" {
		t.Fatalf("display Meta = %#v", sub.Meta)
	}
}

func TestMapParticipant_FallbackUserName(t *testing.T) {
	runID := 1
	name := "[5] Alice"
	run := ejgen.Run{RunId: &runID, UserName: &name, ProbId: intPtr(3)}
	sub, err := mapToSubmission(1, run, nil, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if sub.Participant != "[5] Alice" {
		t.Fatalf("Participant = %q", sub.Participant)
	}
}

func TestMapProblem_FallbackShortThenID(t *testing.T) {
	runID, probID := 1, 42
	short := "A"
	run := ejgen.Run{RunId: &runID, UserLogin: strPtr("u"), ProbShortName: &short}
	sub, err := mapToSubmission(1, run, nil, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if sub.Problem != "A" {
		t.Fatalf("Problem = %q", sub.Problem)
	}

	run = ejgen.Run{RunId: &runID, UserLogin: strPtr("u"), ProbId: &probID}
	sub, err = mapToSubmission(1, run, nil, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if sub.Problem != "42" {
		t.Fatalf("Problem = %q", sub.Problem)
	}
}

func TestMapToSubmission_MissingParticipant(t *testing.T) {
	runID := 1
	_, err := mapToSubmission(1, ejgen.Run{RunId: &runID, ProbId: intPtr(1)}, nil, nil, "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func strPtr(s string) *string { return &s }
func intPtr(n int) *int       { return &n }
