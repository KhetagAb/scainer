package ejudge

import (
	"context"
	"fmt"
	"path"
	"strings"

	"scainer/internal/services/statements"
	"scainer/pkg/lksh"
)

type LkshStatementProvider struct {
	client *lksh.Client
}

func NewLkshStatementProvider(client *lksh.Client) *LkshStatementProvider {
	return &LkshStatementProvider{client: client}
}

func (p *LkshStatementProvider) SourceType() string {
	return statements.SourceLksh
}

func (p *LkshStatementProvider) Fetch(ctx context.Context, c statements.Contest) (statements.Document, error) {
	if p == nil || p.client == nil {
		return statements.Document{}, fmt.Errorf("ejudge statement: lksh provider not configured")
	}
	parallelID := strings.TrimSpace(c.ParallelID)
	if parallelID == "" {
		return statements.Document{}, statements.ErrNotAvailable
	}
	pageHTML, err := p.client.FetchParallelPage(ctx, parallelID)
	if err != nil {
		return statements.Document{}, fmt.Errorf("ejudge statement: lksh parallel page: %w", err)
	}
	pdfURL, err := lksh.StatementPDFURLFromPage(p.client.PortalURL(), parallelID, string(c.ID), pageHTML)
	if err != nil {
		return statements.Document{}, fmt.Errorf("ejudge statement: lksh resolve pdf: %w", err)
	}
	body, contentType, err := p.client.FetchPDF(ctx, pdfURL)
	if err != nil {
		return statements.Document{}, fmt.Errorf("ejudge statement: lksh fetch pdf: %w", err)
	}
	filename := path.Base(pdfURL)
	if filename == "" || filename == "." || filename == "/" {
		filename = string(c.ID) + ".pdf"
	}
	return statements.Document{
		Body:        body,
		ContentType: contentType,
		Filename:    filename,
	}, nil
}

var _ statements.Provider = (*LkshStatementProvider)(nil)
