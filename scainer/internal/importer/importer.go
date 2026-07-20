package importer

import (
	"context"

	"gopkg.in/yaml.v3"

	"github.com/lksh/scainer/internal/domain"
	"github.com/lksh/scainer/internal/store"
)

type Importer interface {
	Name() string
	Import(ctx context.Context, s store.Store) ([]domain.Submission, error)
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
