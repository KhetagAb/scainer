package aiusage

import (
	"context"
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
	prompt := buildTaskPrompt(u)
	return d.Analyzer.Ask(ctx, DetectorTaskName, subject, prompt)
}

func buildTaskPrompt(u domain.ProblemParticipantUnit) string {
	var b strings.Builder
	b.WriteString(taskPromptPreamble)
	b.WriteString("\n\n## Посылки (хронология, последняя — OK)\n\n")
	b.WriteString(FormatSubs(u.Subs))
	b.WriteString("\n\n## Ответ\nВерни только JSON по контракту выше.\n")
	return b.String()
}

const taskPromptPreamble = `Ты помощник преподавателя олимпиадного программирования (ЛКШ).
Оцени, похоже ли, что финальное OK-решение по задаче получено с помощью ИИ / чужого кода, а не самостоятельным прогрессом школьника.

Гипотезы (ищи любое):
1. Поверхность кода — учебниковые/chat-комментарии, шаблонные имена, markdown-артефакты.
2. Несоответствие зрелости — идиомы выше уровня тура, идеальный first-try без черновиков.
3. История по задаче — серия WA/TL-черновиков, затем внезапный «чужой» OK (другой стиль/подход).

Важно: при сомнении ставь низкий score. Лучше пропустить, чем ложно обвинить.

Шкала score:
- 0.0–0.3: нормальный прогресс / недостаточно признаков
- 0.3–0.6: подозрительные намёки, нужна ручная проверка
- 0.6–1.0: сильные признаки (внезапный стиль, tutorial-комменты, maturity mismatch)

JSON-контракт ответа (только JSON, без markdown вокруг):
{
  "score": 0.0,
  "summary": "1–3 предложения по-русски",
  "evidence": [
    {
      "description": "...",
      "submission_id": "...",
      "start_line": 1,
      "end_line": 10
    }
  ]
}

Few-shot (иллюстрация, не копируй score):
История: два коротких WA с printf-отладкой → OK с развёрнутыми docstring на английском и «Here's the approach».
→ score≈0.75, evidence по OK-посылке.`
