package statements

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"scainer/internal/domain"
	"scainer/internal/services/contests"
	"scainer/internal/services/statements/pdfparse"
)

func (s *Service) EnsureContestStatements(ctx context.Context, contestID domain.ContestID) error {
	if s == nil || s.problems == nil {
		return ErrProblemStatementNotFound
	}
	if s.problems.HasSourcePDF(contestID) {
		return nil
	}
	return s.RefreshContestStatements(ctx, contestID)
}

func (s *Service) RefreshContestStatements(ctx context.Context, contestID domain.ContestID) error {
	if s == nil || s.problems == nil {
		return fmt.Errorf("statements: problem store not configured")
	}
	c, err := s.contest(ctx, contestID)
	if err != nil {
		return err
	}
	data, err := s.fetchPDF(ctx, c)
	if err != nil {
		return err
	}
	return s.buildProblemStatementsFromPDF(contestID, data)
}

func (s *Service) InvalidateContestStatements(contestID domain.ContestID) error {
	if s == nil || s.problems == nil {
		return nil
	}
	return s.problems.InvalidateContest(contestID)
}

func (s *Service) OpenContestStatementPDF(ctx context.Context, contestID domain.ContestID) (io.ReadCloser, error) {
	if s == nil || s.problems == nil {
		return nil, ErrProblemStatementNotFound
	}
	if err := s.EnsureContestStatements(ctx, contestID); err != nil {
		return nil, err
	}
	return s.problems.OpenSourcePDF(contestID)
}

func (s *Service) GetProblemStatement(ctx context.Context, contestID domain.ContestID, problemID domain.ProblemID) (domain.ProblemStatement, error) {
	if s == nil || s.problems == nil {
		return domain.ProblemStatement{}, ErrProblemStatementNotFound
	}
	if err := s.EnsureContestStatements(ctx, contestID); err != nil {
		return domain.ProblemStatement{}, err
	}
	return s.problems.GetStatement(contestID, problemID)
}

func (s *Service) OpenProblemStatementPDF(ctx context.Context, contestID domain.ContestID, problemID domain.ProblemID) (io.ReadCloser, error) {
	if s == nil || s.problems == nil {
		return nil, ErrProblemStatementNotFound
	}
	if err := s.EnsureContestStatements(ctx, contestID); err != nil {
		return nil, err
	}
	return s.problems.OpenStatementPDF(contestID, problemID)
}


func (s *Service) IngestContestPDF(_ context.Context, contestID domain.ContestID, r io.Reader) error {
	if s == nil || s.problems == nil {
		return fmt.Errorf("statements: problem store not configured")
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return s.buildProblemStatementsFromPDF(contestID, data)
}

func (s *Service) contest(ctx context.Context, contestID domain.ContestID) (Contest, error) {
	if s == nil || s.registry == nil {
		return Contest{}, fmt.Errorf("statements: service not configured")
	}
	record, ok, err := s.registry.Get(ctx, contestID)
	if err != nil {
		return Contest{}, err
	}
	if !ok {
		return Contest{}, contests.ErrContestNotFound
	}
	return Contest{
		ID:         record.Contest.ID,
		ParallelID: record.Contest.ParallelID,
		Source:     record.Source,
	}, nil
}

func (s *Service) fetchPDF(ctx context.Context, c Contest) ([]byte, error) {
	key := resolveStatementSource(c)
	if key == "" {
		return nil, ErrNotAvailable
	}
	p, ok := s.providers[key]
	if !ok {
		return nil, ErrNotAvailable
	}
	doc, err := p.Fetch(ctx, c)
	if err != nil {
		return nil, err
	}
	defer doc.Body.Close()
	return io.ReadAll(doc.Body)
}

func (s *Service) buildProblemStatementsFromPDF(contestID domain.ContestID, data []byte) error {
	if err := s.problems.WriteSourcePDF(contestID, data); err != nil {
		return err
	}
	parsed, err := pdfparse.ParseContestPDF(bytes.NewReader(data))
	if err != nil {
		return err
	}
	src := s.problems.SourcePDFPath(contestID)
	for pid, ps := range parsed.Statements {
		if err := s.problems.WriteStatement(contestID, ps); err != nil {
			return err
		}
		pr, ok := parsed.PageRanges[pid]
		if !ok {
			continue
		}
		dst := s.problems.ProblemPDFPath(contestID, pid)
		if err := pdfparse.CropPages(src, dst, pr); err != nil {
			return fmt.Errorf("statements: crop %s: %w", pid, err)
		}
	}
	return nil
}
