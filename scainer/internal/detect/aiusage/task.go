package aiusage

import (
	"context"
	"fmt"
	"strings"

	"scainer/internal/detect"
	"scainer/internal/domain"
)

const DetectorTaskName = "aiusage-task"

// TaskDetector — гипотеза по истории попыток участника на одной задаче.
type TaskDetector struct {
	Analyzer *Analyzer
}

func NewTaskDetector(a *Analyzer) *TaskDetector {
	return &TaskDetector{Analyzer: a}
}

var _ detect.Detector[domain.ProblemParticipantUnit] = (*TaskDetector)(nil)

func (d *TaskDetector) Name() string { return DetectorTaskName }

func (d *TaskDetector) AI() bool { return true }

func (d *TaskDetector) Analyze(ctx context.Context, u domain.ProblemParticipantUnit) ([]domain.Signal, error) {
	if len(u.Subs) == 0 {
		return nil, nil
	}
	contest := u.Subs[0].Contest
	subject := domain.NewParticipantProblemSubject(contest, u.Problem, u.Participant)
	current := u.Subs[len(u.Subs)-1]
	prompt := buildTaskPrompt(u)
	return d.Analyzer.Ask(ctx, DetectorTaskName, subject, prompt, current.ID)
}

func buildTaskPrompt(u domain.ProblemParticipantUnit) string {
	var b strings.Builder
	b.WriteString(taskPromptPreamble)
	b.WriteString("\n\n")
	b.WriteString(formatTaskSubs(u.Subs))
	b.WriteString(taskPromptOutro)
	return b.String()
}

// formatTaskSubs: current = last (OK), previous = preceding submission if any.
func formatTaskSubs(subs []domain.Submission) string {
	if len(subs) == 0 {
		return ""
	}
	current := subs[len(subs)-1]
	var b strings.Builder
	fmt.Fprintf(&b, "## Current submission\nsubmission_id=%s problem=%s lang=%s verdict=%s\n```\n%s\n```\n",
		current.ID, current.Problem, current.Lang, current.Verdict,
		truncateSource(string(current.Source), defaultMaxSourceRunes))
	if len(subs) >= 2 {
		prev := subs[len(subs)-2]
		fmt.Fprintf(&b, "\n## Previous submission\nsubmission_id=%s problem=%s lang=%s verdict=%s\n```\n%s\n```\n",
			prev.ID, prev.Problem, prev.Lang, prev.Verdict,
			truncateSource(string(prev.Source), defaultMaxSourceRunes))
	}
	return b.String()
}

const taskPromptPreamble = `You are a highly qualified AI agent for programming tutor. Your task is to assess whether a given solution to an algorithmic problem was written without the use of AI assistance.
Examine the source code of the current submission, as well as the previous submission (for a different task).
If you determine that the code appears to be written by a human, output only one word: ok.
Otherwise, if you suspect AI involvement, output: fail. Then, on the following lines, specify the suspicious code fragments in the format: from:to — your mark or comment regarding the suspected lines.
Please be careful with format, your output will be parsed using programm.`

const taskPromptOutro = `

## Output format (mandatory)
- First line: exactly ok or fail (lowercase).
- If fail: next lines only in the form N:M — comment (1-based line numbers of the current submission).
- No markdown fences, no JSON, no extra prose.
`

/*
Старый JSON-вариант промпта (закомментирован):

func buildTaskPromptJSON(u domain.ProblemParticipantUnit) string {
	var b strings.Builder
	b.WriteString(taskPromptPreambleJSON)
	b.WriteString("\n\n## Посылки (хронология, последняя — OK)\n\n")
	b.WriteString(FormatSubs(u.Subs))
	b.WriteString(taskPromptOutroJSON)
	return b.String()
}

const taskPromptPreambleJSON = `Ты помощник преподавателя олимпиадного программирования (ЛКШ).
Оцени, похоже ли, что финальное OK-решение по задаче получено с помощью ИИ / чужого кода, а не самостоятельным прогрессом школьника.
...
JSON-контракт ответа (только JSON, без markdown вокруг):
{"score": 0.0, "summary": "...", "evidence": [...]}
`

const taskPromptOutroJSON = `
## Ответ — обязательный формат
1. Ответь ОДНИМ JSON-объектом (не массивом []).
2. Поля строго: "score", "summary", "evidence".
`
*/
