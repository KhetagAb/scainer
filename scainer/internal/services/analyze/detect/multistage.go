package detect

import (
	"context"

	"scainer/internal/domain"
	"scainer/internal/services/analyze/selectors"
)

func MultiStage(stages ...Stage) Stage {
	return multiStage(stages)
}

type multiStage []Stage

func (m multiStage) Run(ctx context.Context, store selectors.Store) ([]domain.Signal, error) {
	var out []domain.Signal
	for _, st := range m {
		if st == nil {
			continue
		}
		sigs, err := st.Run(ctx, store)
		if err != nil {
			return nil, err
		}
		out = append(out, sigs...)
	}
	return out, nil
}
