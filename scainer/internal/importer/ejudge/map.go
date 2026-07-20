package ejudge

import (
	"fmt"
	"strconv"
	"time"

	ejgen "github.com/lksh/scainer/generated/ejudge"
	"github.com/lksh/scainer/internal/domain"
	ejudgeapi "github.com/lksh/scainer/pkg/ejudge"
)

func mapToSubmission(contestID int, run ejgen.Run, src []byte, env *ejudgeapi.Env, contestName string) (domain.Submission, error) {
	if run.RunId == nil {
		return domain.Submission{}, fmt.Errorf("ejudge: run без run_id")
	}
	runID := *run.RunId

	participant, err := mapParticipant(run)
	if err != nil {
		return domain.Submission{}, err
	}
	problem, err := mapProblem(run)
	if err != nil {
		return domain.Submission{}, err
	}

	langName := ""
	if run.LangName != nil {
		langName = *run.LangName
	}
	var lang domain.Lang
	if env != nil {
		lang = env.NormalizeLang(langName)
	} else {
		lang = domain.Lang(langName)
	}

	verdict := mapVerdict(run)

	var submittedAt time.Time
	if run.RunTimeUs != nil && *run.RunTimeUs > 0 {
		submittedAt = time.UnixMicro(int64(*run.RunTimeUs))
	}

	meta := map[string]any{
		"run_id":     runID,
		"contest_id": contestID,
	}
	if contestName != "" {
		meta["contest_name"] = contestName
	}
	if run.ProbName != nil && *run.ProbName != "" {
		meta["problem_name"] = *run.ProbName
	}
	if langName != "" {
		meta["lang_name"] = langName
	}
	if run.StatusStr != nil && *run.StatusStr != "" {
		meta["status_str"] = *run.StatusStr
	}
	if run.Status != nil {
		meta["status"] = *run.Status
	}
	if run.Score != nil {
		meta["score"] = *run.Score
	}
	if run.RunUuid != nil && *run.RunUuid != "" {
		meta["run_uuid"] = *run.RunUuid
	}

	contest := strconv.Itoa(contestID)
	return domain.Submission{
		ID:          domain.SubmissionID(fmt.Sprintf("ejudge:%s:%d", contest, runID)),
		Participant: participant,
		Problem:     problem,
		Contest:     domain.ContestID(contest),
		Lang:        lang,
		Source:      src,
		SubmittedAt: submittedAt,
		Verdict:     verdict,
		Meta:        meta,
	}, nil
}

func mapVerdict(run ejgen.Run) domain.Verdict {
	if run.StatusStr != nil && *run.StatusStr != "" {
		return domain.ParseVerdict(*run.StatusStr)
	}
	if run.Status != nil {
		return domain.VerdictFromEjudgeStatus(*run.Status)
	}
	return domain.VerdictUnknown
}

func mapParticipant(run ejgen.Run) (domain.ParticipantID, error) {
	if run.UserLogin != nil && *run.UserLogin != "" {
		return domain.ParticipantID(*run.UserLogin), nil
	}
	if run.UserName != nil && *run.UserName != "" {
		return domain.ParticipantID(*run.UserName), nil
	}
	return "", fmt.Errorf("ejudge: run_id=%v без user_login/user_name", derefInt(run.RunId))
}

func mapProblem(run ejgen.Run) (domain.ProblemID, error) {
	if run.ProbInternalName != nil && *run.ProbInternalName != "" {
		return domain.ProblemID(*run.ProbInternalName), nil
	}
	if run.ProbShortName != nil && *run.ProbShortName != "" {
		return domain.ProblemID(*run.ProbShortName), nil
	}
	if run.ProbId != nil {
		return domain.ProblemID(strconv.Itoa(*run.ProbId)), nil
	}
	return "", fmt.Errorf("ejudge: run_id=%v без ключа задачи", derefInt(run.RunId))
}

func derefInt(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}
