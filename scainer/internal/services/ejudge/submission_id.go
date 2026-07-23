package ejudge

import (
	"fmt"
	"strconv"
	"strings"

	"scainer/internal/domain"
)

const submissionIDPrefix = "ejudge"

type SubmissionKey struct {
	ContestID int
	RunID     int
}

func (k SubmissionKey) ID() string {
	return fmt.Sprintf("%s:%d:%d", submissionIDPrefix, k.ContestID, k.RunID)
}

func (k SubmissionKey) Contest() string {
	return strconv.Itoa(k.ContestID)
}

func ParseSubmissionID(id string) (SubmissionKey, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 || parts[0] != submissionIDPrefix {
		return SubmissionKey{}, fmt.Errorf("ejudge: неверный submission id %q", id)
	}
	contestID, err := strconv.Atoi(parts[1])
	if err != nil {
		return SubmissionKey{}, fmt.Errorf("ejudge: contest_id в %q: %w", id, err)
	}
	runID, err := strconv.Atoi(parts[2])
	if err != nil {
		return SubmissionKey{}, fmt.Errorf("ejudge: run_id в %q: %w", id, err)
	}
	return SubmissionKey{ContestID: contestID, RunID: runID}, nil
}

func parseSubmissionKey(id domain.SubmissionID) (SubmissionKey, error) {
	return ParseSubmissionID(string(id))
}
