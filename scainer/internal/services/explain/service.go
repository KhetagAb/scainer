package explain

import (
	"context"
	"errors"
	"strings"
	"time"

	"scainer/internal/domain"
	"scainer/internal/services/statements"
	"scainer/pkg/llm"

	"golang.org/x/sync/singleflight"
)

type explainPrep struct {
	Letter domain.ProblemID
	Parsed domain.ProblemStatement
	Cached *domain.ProblemStatementExplain
}

type Service struct {
	statements StatementReader
	problems   ProblemStore
	labels     LabelResolver
	llm        llm.Model
	flight     singleflight.Group
}

func New(statements StatementReader, problems ProblemStore, labels LabelResolver, model llm.Model) *Service {
	return &Service{
		statements: statements,
		problems:   problems,
		labels:     labels,
		llm:        model,
	}
}

func (s *Service) Explain(
	ctx context.Context,
	contestID domain.ContestID,
	problemID domain.ProblemID,
) (domain.ProblemStatementExplain, error) {
	prep, err := s.prepareExplain(ctx, contestID, problemID)
	if err != nil {
		return domain.ProblemStatementExplain{}, err
	}
	if prep.Cached != nil {
		return *prep.Cached, nil
	}

	key := formalizationFlightKey(contestID, prep.Letter)
	v, err, _ := s.flight.Do(key, func() (any, error) {
		return s.generateFormalization(ctx, contestID, prep)
	})
	if err != nil {
		return domain.ProblemStatementExplain{}, err
	}
	return v.(domain.ProblemStatementExplain), nil
}

func formalizationFlightKey(contestID domain.ContestID, letter domain.ProblemID) string {
	return string(contestID) + "/" + string(letter)
}

func (s *Service) generateFormalization(
	ctx context.Context,
	contestID domain.ContestID,
	prep explainPrep,
) (domain.ProblemStatementExplain, error) {
	if formal, loadErr := s.problems.GetFormalization(contestID, prep.Letter); loadErr == nil {
		return explainView(formal), nil
	} else if !errors.Is(loadErr, statements.ErrFormalizationNotFound) {
		return domain.ProblemStatementExplain{}, loadErr
	}

	prompt := buildPrompt(prep.Parsed)
	out, err := s.llm.Prompt(ctx, prompt)
	if err != nil {
		if errors.Is(err, llm.ErrNotConfigured) {
			return domain.ProblemStatementExplain{}, ErrNotConfigured
		}
		return domain.ProblemStatementExplain{}, ErrLLM
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return domain.ProblemStatementExplain{}, ErrLLM
	}
	return s.saveAIFormalization(contestID, prep.Letter, out, prompt)
}

func (s *Service) prepareExplain(
	ctx context.Context,
	contestID domain.ContestID,
	problemID domain.ProblemID,
) (explainPrep, error) {
	if err := s.statements.EnsureContestStatements(ctx, contestID); err != nil {
		return explainPrep{}, err
	}

	letter, ps, err := s.resolveAndLoad(ctx, contestID, problemID)
	if err != nil {
		return explainPrep{}, err
	}

	if formal, loadErr := s.problems.GetFormalization(contestID, letter); loadErr == nil {
		ex := explainView(formal)
		return explainPrep{Letter: letter, Cached: &ex}, nil
	} else if !errors.Is(loadErr, statements.ErrFormalizationNotFound) {
		return explainPrep{}, loadErr
	}

	return explainPrep{Letter: letter, Parsed: ps}, nil
}

func (s *Service) Update(
	ctx context.Context,
	contestID domain.ContestID,
	problemID domain.ProblemID,
	title, statement string,
) (domain.ProblemStatementExplain, error) {
	if err := s.statements.EnsureContestStatements(ctx, contestID); err != nil {
		return domain.ProblemStatementExplain{}, err
	}
	letter, _, err := s.resolveAndLoad(ctx, contestID, problemID)
	if err != nil {
		return domain.ProblemStatementExplain{}, err
	}

	title = strings.TrimSpace(title)
	statement = strings.TrimSpace(statement)
	if title == "" || statement == "" {
		return domain.ProblemStatementExplain{}, ErrNotFound
	}

	return s.saveFormalization(domain.ProblemStatementFormalization{
		Contest:   contestID,
		Problem:   letter,
		Title:     title,
		Statement: statement,
		Source:    domain.FormalizationSourceManual,
		Model:     "",
		Prompt:    "",
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Service) Delete(
	ctx context.Context,
	contestID domain.ContestID,
	problemID domain.ProblemID,
) error {
	if err := s.statements.EnsureContestStatements(ctx, contestID); err != nil {
		return err
	}
	letter, _, err := s.resolveAndLoad(ctx, contestID, problemID)
	if err != nil {
		return err
	}
	return s.problems.DeleteFormalization(contestID, letter)
}

func (s *Service) resolveAndLoad(
	ctx context.Context,
	contestID domain.ContestID,
	problemID domain.ProblemID,
) (domain.ProblemID, domain.ProblemStatement, error) {
	letter, ok, err := s.labels.ProblemLabel(ctx, contestID, problemID)
	if err != nil {
		return "", domain.ProblemStatement{}, err
	}
	if !ok || letter == "" {
		return "", domain.ProblemStatement{}, ErrNotFound
	}
	problemLetter := domain.ProblemID(letter)
	ps, err := s.problems.GetStatement(contestID, problemLetter)
	if err != nil {
		return "", domain.ProblemStatement{}, ErrNotFound
	}
	return problemLetter, ps, nil
}

func (s *Service) saveAIFormalization(
	contestID domain.ContestID,
	letter domain.ProblemID,
	statement, prompt string,
) (domain.ProblemStatementExplain, error) {
	return s.saveFormalization(domain.ProblemStatementFormalization{
		Contest:   contestID,
		Problem:   letter,
		Title:     string(letter),
		Statement: statement,
		Source:    domain.FormalizationSourceAI,
		Model:     strings.TrimSpace(s.llm.ModelName()),
		Prompt:    prompt,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Service) saveFormalization(f domain.ProblemStatementFormalization) (domain.ProblemStatementExplain, error) {
	if err := s.problems.SaveFormalization(f); err != nil {
		return domain.ProblemStatementExplain{}, err
	}
	return explainView(f), nil
}

func explainView(f domain.ProblemStatementFormalization) domain.ProblemStatementExplain {
	return domain.ProblemStatementExplain{
		Problem:   f.Problem,
		Title:     f.Title,
		Statement: f.Statement,
		Source:    f.Source,
		Model:     f.Model,
		Prompt:    f.Prompt,
	}
}
