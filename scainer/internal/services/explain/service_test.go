package explain_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"scainer/internal/domain"
	"scainer/internal/services/explain"
	"scainer/internal/services/statements"
	"scainer/pkg/llm"
)

type noopStatementReader struct{}

func (noopStatementReader) EnsureContestStatements(context.Context, domain.ContestID) error {
	return nil
}

type fakeLabelResolver struct {
	labels map[domain.ProblemID]string
}

func (r fakeLabelResolver) ProblemLabel(_ context.Context, _ domain.ContestID, problemID domain.ProblemID) (string, bool, error) {
	label, ok := r.labels[problemID]
	return label, ok, nil
}

type memProblemStore struct {
	mu             sync.Mutex
	statements     map[string]domain.ProblemStatement
	formalizations map[string]domain.ProblemStatementFormalization
}

func newMemProblemStore() *memProblemStore {
	return &memProblemStore{
		statements:     make(map[string]domain.ProblemStatement),
		formalizations: make(map[string]domain.ProblemStatementFormalization),
	}
}

func (s *memProblemStore) key(contest domain.ContestID, problem domain.ProblemID) string {
	return string(contest) + "/" + string(problem)
}

func (s *memProblemStore) GetStatement(contestID domain.ContestID, problemID domain.ProblemID) (domain.ProblemStatement, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ps, ok := s.statements[s.key(contestID, problemID)]
	if !ok {
		return domain.ProblemStatement{}, statements.ErrProblemStatementNotFound
	}
	return ps, nil
}

func (s *memProblemStore) GetFormalization(contestID domain.ContestID, problemID domain.ProblemID) (domain.ProblemStatementFormalization, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, ok := s.formalizations[s.key(contestID, problemID)]
	if !ok {
		return domain.ProblemStatementFormalization{}, statements.ErrFormalizationNotFound
	}
	return f, nil
}

func (s *memProblemStore) SaveFormalization(f domain.ProblemStatementFormalization) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.formalizations[s.key(f.Contest, f.Problem)] = f
	return nil
}

func (s *memProblemStore) DeleteFormalization(contestID domain.ContestID, problemID domain.ProblemID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.formalizations, s.key(contestID, problemID))
	return nil
}

type stubModel struct {
	out string
	err error
}

func (m stubModel) Prompt(context.Context, string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.out, nil
}

func (m stubModel) PromptStream(_ context.Context, _ string) (<-chan llm.StreamChunk, error) {
	if m.err != nil {
		return nil, m.err
	}
	ch := make(chan llm.StreamChunk, 2)
	go func() {
		defer close(ch)
		ch <- llm.StreamChunk{Text: m.out}
		ch <- llm.StreamChunk{Done: true}
	}()
	return ch, nil
}

func (m stubModel) ModelName() string { return "stub-model" }

type blockingCountingModel struct {
	mu    sync.Mutex
	calls int
	block chan struct{}
	out   string
}

func (m *blockingCountingModel) Prompt(context.Context, string) (string, error) {
	m.mu.Lock()
	m.calls++
	m.mu.Unlock()
	<-m.block
	return m.out, nil
}

func (m *blockingCountingModel) PromptStream(context.Context, string) (<-chan llm.StreamChunk, error) {
	return nil, errors.New("not implemented")
}

func (m *blockingCountingModel) ModelName() string { return "blocking-model" }

func (m *blockingCountingModel) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

func TestExplain_usesLLMAndCaches(t *testing.T) {
	store := newMemProblemStore()
	store.statements["50506/G"] = domain.ProblemStatement{
		Contest: "50506",
		Problem: "G",
		Title:   "Gray code",
		RawStatement: "Сюжет.\n\nФормат входных данных\nn\n\nФормат выходных данных\nk",
	}
	svc := explain.New(noopStatementReader{}, store, fakeLabelResolver{labels: map[domain.ProblemID]string{"gray-code": "G"}}, stubModel{out: "## Формализация\n\nТекст"})

	ex, err := svc.Explain(context.Background(), "50506", "gray-code")
	if err != nil {
		t.Fatal(err)
	}
	if ex.Problem != "G" || ex.Source != domain.FormalizationSourceAI {
		t.Fatalf("ex=%+v", ex)
	}
	if ex.Model != "stub-model" {
		t.Fatalf("model=%q", ex.Model)
	}
	if !strings.Contains(ex.Prompt, "Сюжет.") {
		t.Fatalf("prompt=%q", ex.Prompt)
	}
	if !strings.Contains(ex.Prompt, "Формат входных данных") {
		t.Fatalf("prompt should include input format block: %q", ex.Prompt)
	}
	if strings.Contains(ex.Prompt, "Формат выходных данных") {
		t.Fatalf("prompt should not include output format block: %q", ex.Prompt)
	}

	cached, err := svc.Explain(context.Background(), "50506", "gray-code")
	if err != nil {
		t.Fatal(err)
	}
	if cached.Statement != ex.Statement {
		t.Fatalf("cached=%q", cached.Statement)
	}
}

func TestExplain_deduplicatesConcurrentGeneration(t *testing.T) {
	store := newMemProblemStore()
	store.statements["50506/G"] = domain.ProblemStatement{
		Contest: "50506",
		Problem: "G",
		Title:   "Gray code",
		RawStatement: "Сюжет.\n\nФормат входных данных\nn\n\nФормат выходных данных\nk",
	}
	block := make(chan struct{})
	model := &blockingCountingModel{block: block, out: "formalized"}
	svc := explain.New(noopStatementReader{}, store, fakeLabelResolver{labels: map[domain.ProblemID]string{"gray-code": "G"}}, model)
	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(2)
	for range 2 {
		go func() {
			defer wg.Done()
			ex, err := svc.Explain(ctx, "50506", "gray-code")
			if err != nil {
				t.Error(err)
				return
			}
			if ex.Statement != "formalized" {
				t.Errorf("statement=%q", ex.Statement)
			}
		}()
	}

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if model.callCount() == 1 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if got := model.callCount(); got != 1 {
		t.Fatalf("llm calls=%d want 1", got)
	}

	close(block)
	wg.Wait()
}

func TestDelete_allowsRegeneration(t *testing.T) {
	store := newMemProblemStore()
	store.statements["50506/G"] = domain.ProblemStatement{
		Contest: "50506",
		Problem: "G",
		Title:   "Gray code",
		RawStatement: "Сюжет.\n\nФормат входных данных\nn\n\nФормат выходных данных\nk",
	}
	svc := explain.New(noopStatementReader{}, store, fakeLabelResolver{labels: map[domain.ProblemID]string{"gray-code": "G"}}, stubModel{out: "first"})
	ctx := context.Background()

	if _, err := svc.Explain(ctx, "50506", "gray-code"); err != nil {
		t.Fatal(err)
	}
	if err := svc.Delete(ctx, "50506", "gray-code"); err != nil {
		t.Fatal(err)
	}

	svc2 := explain.New(noopStatementReader{}, store, fakeLabelResolver{labels: map[domain.ProblemID]string{"gray-code": "G"}}, stubModel{out: "second"})
	ex, err := svc2.Explain(ctx, "50506", "gray-code")
	if err != nil {
		t.Fatal(err)
	}
	if ex.Statement != "second" {
		t.Fatalf("statement=%q", ex.Statement)
	}
}

func TestExplain_unconfiguredOnCacheMiss(t *testing.T) {
	store := newMemProblemStore()
	store.statements["50506/G"] = domain.ProblemStatement{Contest: "50506", Problem: "G", Title: "t", RawStatement: "raw"}
	svc := explain.New(noopStatementReader{}, store, fakeLabelResolver{labels: map[domain.ProblemID]string{"gray-code": "G"}}, llm.Unconfigured{})

	_, err := svc.Explain(context.Background(), "50506", "gray-code")
	if !errors.Is(err, explain.ErrNotConfigured) {
		t.Fatalf("err=%v", err)
	}
}
