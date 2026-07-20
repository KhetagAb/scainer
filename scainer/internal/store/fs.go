package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"scainer/internal/domain"
	pkgfs "scainer/pkg/fs"
)

const (
	metaFileName    = "meta.json"
	sourceBaseName  = "source"
	defaultSrcExt   = ".src"
	cursorsFileName = "cursors.json"
)

type FS struct {
	root    string
	mu      sync.RWMutex
	byID    map[domain.SubmissionID]domain.Submission
	order   []domain.SubmissionID
	cursors map[string]string
}

type submissionMeta struct {
	ID          domain.SubmissionID
	Participant domain.ParticipantID
	Problem     domain.ProblemID
	Contest     domain.ContestID
	Lang        domain.Lang
	SubmittedAt time.Time
	Verdict     domain.Verdict
	Meta        map[string]any
}

func NewFS(root string) (*FS, error) {
	inst := &FS{
		root:    root,
		byID:    make(map[domain.SubmissionID]domain.Submission),
		cursors: make(map[string]string),
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("fs store: mkdir root: %w", err)
	}
	if err := inst.hydrate(); err != nil {
		return nil, fmt.Errorf("fs store: hydrate: %w", err)
	}
	return inst, nil
}

var _ Store = (*FS)(nil)
var _ SourcePather = (*FS)(nil)

func (f *FS) Root() string { return f.root }

func (f *FS) SourcePath(s domain.Submission) (string, bool) {
	path := f.sourceFilePath(s)
	if _, err := os.Stat(path); err != nil {
		return "", false
	}
	return path, true
}

func (f *FS) submissionDir(s domain.Submission) string {
	return filepath.Join(f.root,
		pkgfs.SanitizeFileName(string(s.Contest)),
		pkgfs.SanitizeFileName(string(s.Problem)),
		pkgfs.SanitizeFileName(string(s.Participant)),
		pkgfs.SanitizeFileName(string(s.ID)),
	)
}

func (f *FS) sourceFilePath(s domain.Submission) string {
	ext := domain.FileExt[s.Lang]
	if ext == "" {
		ext = defaultSrcExt
	}
	return filepath.Join(f.submissionDir(s), sourceBaseName+ext)
}

func (f *FS) hydrate() error {
	metaPaths, err := findMetaFiles(f.root)
	if err != nil {
		return err
	}
	for _, p := range metaPaths {
		sub, err := readSubmission(p)
		if err != nil {
			return fmt.Errorf("read %s: %w", p, err)
		}
		if _, ok := f.byID[sub.ID]; !ok {
			f.order = append(f.order, sub.ID)
		}
		f.byID[sub.ID] = sub
	}

	cursorsPath := filepath.Join(f.root, cursorsFileName)
	raw, err := os.ReadFile(cursorsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(raw, &f.cursors)
}

func findMetaFiles(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == metaFileName {
			out = append(out, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func readSubmission(metaPath string) (domain.Submission, error) {
	raw, err := os.ReadFile(metaPath)
	if err != nil {
		return domain.Submission{}, err
	}
	var m submissionMeta
	if err := json.Unmarshal(raw, &m); err != nil {
		return domain.Submission{}, err
	}

	dir := filepath.Dir(metaPath)
	ext := domain.FileExt[m.Lang]
	if ext == "" {
		ext = defaultSrcExt
	}
	src, err := os.ReadFile(filepath.Join(dir, sourceBaseName+ext))
	if err != nil {
		return domain.Submission{}, err
	}

	return domain.Submission{
		ID:          m.ID,
		Participant: m.Participant,
		Problem:     m.Problem,
		Contest:     m.Contest,
		Lang:        m.Lang,
		Source:      src,
		SubmittedAt: m.SubmittedAt,
		Verdict:     m.Verdict,
		Meta:        m.Meta,
	}, nil
}

func (f *FS) Put(ctx context.Context, subs []domain.Submission) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, s := range subs {
		if err := f.writeSubmission(s); err != nil {
			return fmt.Errorf("fs store: write %s: %w", s.ID, err)
		}
		if _, ok := f.byID[s.ID]; !ok {
			f.order = append(f.order, s.ID)
		}
		f.byID[s.ID] = s
	}
	return nil
}

func (f *FS) writeSubmission(s domain.Submission) error {
	ext := domain.FileExt[s.Lang]
	if ext == "" {
		ext = defaultSrcExt
	}

	dir := f.submissionDir(s)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	metaRaw, err := json.MarshalIndent(submissionMeta{
		ID:          s.ID,
		Participant: s.Participant,
		Problem:     s.Problem,
		Contest:     s.Contest,
		Lang:        s.Lang,
		SubmittedAt: s.SubmittedAt,
		Verdict:     s.Verdict,
		Meta:        s.Meta,
	}, "", "  ")
	if err != nil {
		return err
	}

	if err := pkgfs.WriteFileAtomic(filepath.Join(dir, metaFileName), metaRaw, 0o644); err != nil {
		return err
	}
	return pkgfs.WriteFileAtomic(filepath.Join(dir, sourceBaseName+ext), s.Source, 0o644)
}

func (f *FS) Get(ctx context.Context, ids []domain.SubmissionID) ([]domain.Submission, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make([]domain.Submission, 0, len(ids))
	for _, id := range ids {
		if s, ok := f.byID[id]; ok {
			out = append(out, s)
		}
	}
	return out, nil
}

func (f *FS) ByProblem(ctx context.Context, contest domain.ContestID) (map[domain.ProblemID][]domain.Submission, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make(map[domain.ProblemID][]domain.Submission)
	for _, id := range f.order {
		s := f.byID[id]
		if s.Contest != contest {
			continue
		}
		out[s.Problem] = append(out[s.Problem], s)
	}
	return out, nil
}

func (f *FS) ByParticipant(ctx context.Context) (map[domain.ParticipantID][]domain.Submission, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make(map[domain.ParticipantID][]domain.Submission)
	for _, id := range f.order {
		s := f.byID[id]
		out[s.Participant] = append(out[s.Participant], s)
	}
	return out, nil
}

func (f *FS) GetCursor(ctx context.Context, key string) (string, bool, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	v, ok := f.cursors[key]
	return v, ok, nil
}

func (f *FS) SetCursor(ctx context.Context, key string, value string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	next := make(map[string]string, len(f.cursors)+1)
	for k, v := range f.cursors {
		next[k] = v
	}
	next[key] = value

	raw, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return err
	}
	if err := pkgfs.WriteFileAtomic(filepath.Join(f.root, cursorsFileName), raw, 0o644); err != nil {
		return fmt.Errorf("fs store: write cursors: %w", err)
	}
	f.cursors = next
	return nil
}
