package statements

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"scainer/internal/domain"
	storefs "scainer/pkg/store/fs"
)

var (
	ErrProblemStatementNotFound = errors.New("problem statement not found")
	ErrFormalizationNotFound    = errors.New("problem statement formalization not found")
)

const problemStatementsDir = "problem-statements"

type ProblemStore struct {
	root string
}

func NewProblemStore(storeDir string) *ProblemStore {
	if storeDir == "" {
		return nil
	}
	return &ProblemStore{root: filepath.Join(storeDir, problemStatementsDir)}
}

func (s *ProblemStore) contestDir(contestID domain.ContestID) string {
	return filepath.Join(s.root, storefs.SanitizeFileName(string(contestID)))
}

func (s *ProblemStore) yamlPath(contestID domain.ContestID, problemID domain.ProblemID) string {
	return filepath.Join(s.contestDir(contestID), storefs.SanitizeFileName(string(problemID))+".yaml")
}

func (s *ProblemStore) formalizationPath(contestID domain.ContestID, problemID domain.ProblemID) string {
	return filepath.Join(s.contestDir(contestID), storefs.SanitizeFileName(string(problemID))+"-explain.yaml")
}

func (s *ProblemStore) pdfPath(contestID domain.ContestID, problemID domain.ProblemID) string {
	return filepath.Join(s.contestDir(contestID), storefs.SanitizeFileName(string(problemID))+".pdf")
}

func (s *ProblemStore) sourcePDFPath(contestID domain.ContestID) string {
	return filepath.Join(s.contestDir(contestID), "source.pdf")
}

func (s *ProblemStore) GetStatement(contestID domain.ContestID, problemID domain.ProblemID) (domain.ProblemStatement, error) {
	if s == nil {
		return domain.ProblemStatement{}, ErrProblemStatementNotFound
	}
	data, err := os.ReadFile(s.yamlPath(contestID, problemID))
	if err != nil {
		if os.IsNotExist(err) {
			return domain.ProblemStatement{}, ErrProblemStatementNotFound
		}
		return domain.ProblemStatement{}, err
	}
	var ps domain.ProblemStatement
	if err := yaml.Unmarshal(data, &ps); err != nil {
		return domain.ProblemStatement{}, fmt.Errorf("problem store: yaml: %w", err)
	}
	return ps, nil
}

func (s *ProblemStore) OpenStatementPDF(contestID domain.ContestID, problemID domain.ProblemID) (io.ReadCloser, error) {
	if s == nil {
		return nil, ErrProblemStatementNotFound
	}
	f, err := os.Open(s.pdfPath(contestID, problemID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrProblemStatementNotFound
		}
		return nil, err
	}
	return f, nil
}

func (s *ProblemStore) HasSourcePDF(contestID domain.ContestID) bool {
	if s == nil {
		return false
	}
	_, err := os.Stat(s.sourcePDFPath(contestID))
	return err == nil
}

func (s *ProblemStore) OpenSourcePDF(contestID domain.ContestID) (io.ReadCloser, error) {
	if s == nil {
		return nil, ErrProblemStatementNotFound
	}
	f, err := os.Open(s.sourcePDFPath(contestID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrProblemStatementNotFound
		}
		return nil, err
	}
	return f, nil
}

func (s *ProblemStore) InvalidateContest(contestID domain.ContestID) error {
	if s == nil {
		return nil
	}
	dir := s.contestDir(contestID)
	if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("problem store: invalidate %s: %w", contestID, err)
	}
	return nil
}

func (s *ProblemStore) WriteSourcePDF(contestID domain.ContestID, data []byte) error {
	if s == nil {
		return fmt.Errorf("problem store: not configured")
	}
	path := s.sourcePDFPath(contestID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return storefs.WriteFileAtomic(path, data, 0o644)
}

func (s *ProblemStore) WriteStatement(contestID domain.ContestID, ps domain.ProblemStatement) error {
	if s == nil {
		return fmt.Errorf("problem store: not configured")
	}
	ps.Contest = contestID
	data, err := yaml.Marshal(ps)
	if err != nil {
		return err
	}
	path := s.yamlPath(contestID, ps.Problem)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return storefs.WriteFileAtomic(path, data, 0o644)
}

func (s *ProblemStore) SourcePDFPath(contestID domain.ContestID) string {
	if s == nil {
		return ""
	}
	return s.sourcePDFPath(contestID)
}

func (s *ProblemStore) ProblemPDFPath(contestID domain.ContestID, problemID domain.ProblemID) string {
	if s == nil {
		return ""
	}
	return s.pdfPath(contestID, problemID)
}

func (s *ProblemStore) GetFormalization(contestID domain.ContestID, problemID domain.ProblemID) (domain.ProblemStatementFormalization, error) {
	if s == nil {
		return domain.ProblemStatementFormalization{}, ErrFormalizationNotFound
	}
	data, err := os.ReadFile(s.formalizationPath(contestID, problemID))
	if err != nil {
		if os.IsNotExist(err) {
			return domain.ProblemStatementFormalization{}, ErrFormalizationNotFound
		}
		return domain.ProblemStatementFormalization{}, err
	}
	var f domain.ProblemStatementFormalization
	if err := yaml.Unmarshal(data, &f); err != nil {
		return domain.ProblemStatementFormalization{}, fmt.Errorf("problem store: formalization yaml: %w", err)
	}
	return f, nil
}

func (s *ProblemStore) SaveFormalization(f domain.ProblemStatementFormalization) error {
	if s == nil {
		return fmt.Errorf("problem store: not configured")
	}
	data, err := yaml.Marshal(f)
	if err != nil {
		return err
	}
	path := s.formalizationPath(f.Contest, f.Problem)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return storefs.WriteFileAtomic(path, data, 0o644)
}

func (s *ProblemStore) DeleteFormalization(contestID domain.ContestID, problemID domain.ProblemID) error {
	if s == nil {
		return fmt.Errorf("problem store: not configured")
	}
	err := os.Remove(s.formalizationPath(contestID, problemID))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
