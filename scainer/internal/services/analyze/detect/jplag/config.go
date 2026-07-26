package jplag

import (
	"fmt"

	"scainer/internal/configs"
	"scainer/internal/domain"
	"scainer/internal/services/analyze/detect"
	"scainer/pkg/store"
)

func NewFromConfig(cfg configs.JPlagConfig, st *store.FS) (detect.Detector[domain.ProblemUnit], error) {
	d, err := New(cfg.JarPath, st)
	if err != nil {
		return nil, fmt.Errorf("jplag: %w", err)
	}
	if err := d.CheckRuntime(); err != nil {
		return nil, err
	}
	return d, nil
}
