package teachers

import (
	"context"
	"crypto/subtle"
	"errors"
)

var (
	ErrNotFound            = errors.New("teacher not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
)

type repository interface {
	Get(ctx context.Context, login string) (Record, bool, error)
}

type Service struct {
	repo repository
}

func NewService(repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Login(ctx context.Context, login, password string) error {
	rec, ok, err := s.repo.Get(ctx, login)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFound
	}
	if subtle.ConstantTimeCompare([]byte(rec.Password), []byte(password)) != 1 {
		return ErrInvalidCredentials
	}
	return nil
}
