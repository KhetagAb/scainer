package aiusage

import (
	"fmt"
	"strings"
	"time"

	"scainer/internal/domain"
)

const defaultMaxSourceRunes = 3500

func FormatSubs(subs []domain.Submission) string {
	return FormatSubsLimited(subs, defaultMaxSourceRunes)
}

func FormatSubsLimited(subs []domain.Submission, maxSourceRunes int) string {
	if maxSourceRunes <= 0 {
		maxSourceRunes = defaultMaxSourceRunes
	}
	var b strings.Builder
	for i, sub := range subs {
		if i > 0 {
			b.WriteString("\n\n")
		}
		fmt.Fprintf(&b, "### submission_id=%s problem=%s participant=%s lang=%s verdict=%s submitted_at=%s\n",
			sub.ID, sub.Problem, sub.Participant, sub.Lang, sub.Verdict, sub.SubmittedAt.UTC().Format(time.RFC3339))
		b.WriteString("```\n")
		b.WriteString(truncateSource(string(sub.Source), maxSourceRunes))
		b.WriteString("\n```")
	}
	return b.String()
}

func truncateSource(src string, maxRunes int) string {
	r := []rune(src)
	if len(r) <= maxRunes {
		return src
	}
	return string(r[:maxRunes]) + "\n…[truncated]"
}
