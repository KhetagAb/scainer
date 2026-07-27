package store

import (
	"context"
	"sync"

	"scainer/internal/domain"
)

type SourcePather interface {
	SourcePath(s domain.Submission) (path string, ok bool)
}

type Mem struct {
	mu      sync.RWMutex
	byID    map[domain.SubmissionID]domain.Submission
	order   []domain.SubmissionID
	cursors map[string]string
}

func NewMem() *Mem {
	return &Mem{
		byID:    make(map[domain.SubmissionID]domain.Submission),
		cursors: make(map[string]string),
	}
}

func (m *Mem) Put(ctx context.Context, submissions []domain.Submission) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, submission := range submissions {
		if _, ok := m.byID[submission.ID]; !ok {
			m.order = append(m.order, submission.ID)
		}
		m.byID[submission.ID] = submission
	}
	return nil
}

func (m *Mem) GetByID(ctx context.Context, id domain.SubmissionID) (domain.Submission, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.byID[id]
	if !ok {
		return domain.Submission{}, domain.ErrSubmissionNotFound
	}
	return s, nil
}

func (m *Mem) GetByIDs(ctx context.Context, ids []domain.SubmissionID) ([]domain.Submission, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.Submission, 0, len(ids))
	for _, id := range ids {
		if s, ok := m.byID[id]; ok {
			out = append(out, s)
		}
	}
	return out, nil
}

func (m *Mem) ByProblem(ctx context.Context, contest domain.ContestID) (map[domain.ProblemID][]domain.Submission, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[domain.ProblemID][]domain.Submission)
	for _, id := range m.order {
		s := m.byID[id]
		if s.Contest != contest {
			continue
		}
		out[s.Problem] = append(out[s.Problem], s)
	}
	return out, nil
}

func (m *Mem) ByParticipant(ctx context.Context) (map[domain.ParticipantID][]domain.Submission, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[domain.ParticipantID][]domain.Submission)
	for _, id := range m.order {
		s := m.byID[id]
		out[s.Participant] = append(out[s.Participant], s)
	}
	return out, nil
}

func (m *Mem) GetCursor(ctx context.Context, key string) (string, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.cursors[key]
	return v, ok, nil
}

func (m *Mem) SetCursor(ctx context.Context, key string, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cursors[key] = value
	return nil
}

func (m *Mem) DeleteByContest(ctx context.Context, contest domain.ContestID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var nextOrder []domain.SubmissionID
	for _, id := range m.order {
		s, ok := m.byID[id]
		if !ok || s.Contest != contest {
			nextOrder = append(nextOrder, id)
			continue
		}
		delete(m.byID, id)
	}
	m.order = nextOrder
	return nil
}

func (m *Mem) DeleteCursor(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.cursors, key)
	return nil
}
