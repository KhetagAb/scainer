package teachers

import (
	"context"
	"crypto/subtle"
	"errors"
	"log"
)

var (
	ErrNotFound           = errors.New("teacher not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type repository interface {
	Get(ctx context.Context, login string) (Record, bool, error)
}

type loginAudit interface {
	LogLoginFailed(ctx context.Context, login, password string) error
}

type Service struct {
	repo  repository
	audit loginAudit
}

func NewService(repo repository, audit loginAudit) *Service {
	return &Service{repo: repo, audit: audit}
}

func (s *Service) Login(ctx context.Context, login, password string) error {
	rec, ok, err := s.repo.Get(ctx, login)
	if err != nil {
		return err
	}
	if !ok {
		s.recordFailedLogin(ctx, login, password)
		return ErrNotFound
	}
	if subtle.ConstantTimeCompare([]byte(rec.Password), []byte(password)) != 1 {
		s.recordFailedLogin(ctx, login, password)
		return ErrInvalidCredentials
	}
	return nil
}

func (s *Service) recordFailedLogin(ctx context.Context, login, password string) {
	if s.audit == nil {
		return
	}
	if err := s.audit.LogLoginFailed(ctx, login, password); err != nil {
		log.Printf("login audit: %v", err)
	}
}
