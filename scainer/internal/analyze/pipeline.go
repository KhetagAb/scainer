package analyze

import (
	"scainer/internal/detect"
	"scainer/internal/domain"
)

// Pipeline builds detect stages for a contest.
// TODO: someday may want to load/run multiple contests in one stage.
type Pipeline interface {
	ForContest(id domain.ContestID) []detect.Stage
}
