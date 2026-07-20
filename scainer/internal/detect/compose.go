package detect

import "scainer/internal/domain"

// StageFactory builds a Stage for one contest.
// TODO: someday may want to load/run multiple contests in one stage.
type StageFactory func(contestID domain.ContestID) Stage

type Pipeline []StageFactory

func Compose(factories ...StageFactory) Pipeline {
	return Pipeline(factories)
}

func (p Pipeline) ForContest(id domain.ContestID) []Stage {
	out := make([]Stage, 0, len(p))
	for _, f := range p {
		if f == nil {
			continue
		}
		out = append(out, f(id))
	}
	return out
}
