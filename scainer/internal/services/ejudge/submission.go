package ejudge

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	ejgen "scainer/generated/ejudge"
	"scainer/internal/domain"
	ejudgeapi "scainer/pkg/ejudge"
)

var defaultLangMap = map[string]domain.Lang{
	"gcc": domain.LangCPP, "g++": domain.LangCPP, "clang++": domain.LangCPP,
	"python3": domain.LangPython, "python": domain.LangPython,
	"java": domain.LangJava,
	"go":   domain.LangGo,
}

func submissionFromRun(contestID int, run ejgen.Run, src []byte, contestName string) (domain.Submission, error) {
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
	lang := normalizeLang(langName)

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

	key := SubmissionKey{ContestID: contestID, RunID: runID}
	return domain.Submission{
		ID:          domain.SubmissionID(key.ID()),
		Participant: participant,
		Problem:     problem,
		Contest:     domain.ContestID(key.Contest()),
		Lang:        lang,
		Source:      src,
		SubmittedAt: submittedAt,
		Verdict:     verdict,
		Meta:        meta,
	}, nil
}

func normalizeLang(langName string) domain.Lang {
	if langName == "" {
		return ""
	}
	if v, ok := defaultLangMap[langName]; ok {
		return v
	}
	return guessLang(langName)
}

func guessLang(langName string) domain.Lang {
	lower := strings.ToLower(langName)
	switch {
	case strings.Contains(lower, "clang++"),
		strings.HasPrefix(lower, "g++"),
		strings.HasPrefix(lower, "gcc"),
		strings.Contains(lower, "c++"):
		return domain.LangCPP
	case strings.HasPrefix(lower, "python"):
		return domain.LangPython
	case strings.HasPrefix(lower, "java"):
		return domain.LangJava
	case lower == "go" || strings.HasPrefix(lower, "golang"):
		return domain.LangGo
	case strings.Contains(lower, "javascript"), strings.HasPrefix(lower, "node"):
		return domain.LangJavaScript
	default:
		return domain.Lang(langName)
	}
}

func mapVerdict(run ejgen.Run) domain.Verdict {
	if run.StatusStr != nil && *run.StatusStr != "" {
		return toDomainVerdict(ejudgeapi.ParseVerdict(*run.StatusStr))
	}
	if run.Status != nil {
		return toDomainVerdict(ejudgeapi.VerdictFromStatus(*run.Status))
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

func hasProblemKey(run ejgen.Run) bool {
	_, err := mapProblem(run)
	return err == nil
}

func derefInt(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}
