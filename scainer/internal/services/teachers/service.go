package teachers

import (
	"context"
	"errors"
)

var (
	ErrNotFound           = errors.New("teacher not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type repository interface {
	Get(ctx context.Context, login string) (Record, bool, error)
	Upsert(ctx context.Context, login, password string) error
}

type ejudgeAuthenticator interface {
	Validate(ctx context.Context, login, password string) error
}

type Service struct {
	repo repository
	ej   ejudgeAuthenticator
}

func NewService(repo repository, ej ejudgeAuthenticator) *Service {
	return &Service{repo: repo, ej: ej}
}

func (s *Service) Login(ctx context.Context, login, password string) error {
	_, ok, err := s.repo.Get(ctx, login)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFound
	}
	if s.ej == nil {
		return ErrInvalidCredentials
	}
	if err := s.ej.Validate(ctx, login, password); err != nil {
		return ErrInvalidCredentials
	}
	// TODO: plaintext password, replace with hashing
	return s.repo.Upsert(ctx, login, password)
}
