package importer

import (
	"context"

	"gopkg.in/yaml.v3"

	"scainer/internal/domain"
)

type Store interface {
	ByProblem(ctx context.Context, contest domain.ContestID) (map[domain.ProblemID][]domain.Submission, error)
	GetCursor(ctx context.Context, key string) (value string, ok bool, err error)
	SetCursor(ctx context.Context, key string, value string) error
}

type Result struct {
	ContestName string
	Submissions []domain.Submission
}

type Importer interface {
	Name() string
	Import(ctx context.Context, store Store) (Result, error)
}

type Factory func(cfg *yaml.Node) (Importer, error)

var registry = map[string]Factory{}

func Register(kind string, f Factory) { registry[kind] = f }

func Build(kind string, cfg *yaml.Node) (Importer, error) {
	f, ok := registry[kind]
	if !ok {
		return nil, &UnknownSourceError{Kind: kind}
	}
	return f(cfg)
}

type UnknownSourceError struct{ Kind string }

func (e *UnknownSourceError) Error() string { return "importer: unknown source type: " + e.Kind }
