package explain

import (
	"context"

	"scainer/internal/domain"
)

type StatementReader interface {
	EnsureContestStatements(ctx context.Context, contestID domain.ContestID) error
}

type ProblemStore interface {
	GetStatement(contestID domain.ContestID, problemID domain.ProblemID) (domain.ProblemStatement, error)
	GetFormalization(contestID domain.ContestID, problemID domain.ProblemID) (domain.ProblemStatementFormalization, error)
	SaveFormalization(f domain.ProblemStatementFormalization) error
	DeleteFormalization(contestID domain.ContestID, problemID domain.ProblemID) error
}
