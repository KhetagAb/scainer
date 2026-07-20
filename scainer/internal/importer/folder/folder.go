package folder

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"scainer/internal/domain"
	"scainer/internal/importer"
)

func init() { importer.Register("folder", newFromConfig) }

type Config struct {
	Root    string `yaml:"root"`
	Contest string `yaml:"contest"`
}

type Importer struct{ cfg Config }

var _ importer.Importer = (*Importer)(nil)

func (i *Importer) Name() string { return "folder" }

func newFromConfig(node *yaml.Node) (importer.Importer, error) {
	var cfg Config
	if node != nil && node.Kind != 0 {
		if err := node.Decode(&cfg); err != nil {
			return nil, err
		}
	}
	if cfg.Root == "" {
		return nil, errors.New("folder importer: поле root обязательно")
	}
	return &Importer{cfg: cfg}, nil
}

var langByExt = map[string]domain.Lang{
	".cpp": domain.LangCPP, ".cc": domain.LangCPP, ".cxx": domain.LangCPP, ".hpp": domain.LangCPP,
	".py": domain.LangPython, ".java": domain.LangJava, ".go": domain.LangGo, ".js": domain.LangJavaScript,
}

func (i *Importer) Import(ctx context.Context, _ importer.Store) ([]domain.Submission, error) {
	problems, err := os.ReadDir(i.cfg.Root)
	if err != nil {
		return nil, err
	}

	var subs []domain.Submission
	for _, pd := range problems {
		if !pd.IsDir() {
			continue
		}
		problem := pd.Name()
		dir := filepath.Join(i.cfg.Root, problem)
		files, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			name := f.Name()
			ext := strings.ToLower(filepath.Ext(name))
			participant := strings.TrimSuffix(name, filepath.Ext(name))
			src, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				return nil, err
			}
			id := i.cfg.Contest + "/" + problem + "/" + participant
			subs = append(subs, domain.Submission{
				ID:          domain.SubmissionID(id),
				Participant: domain.ParticipantID(participant),
				Problem:     domain.ProblemID(problem),
				Contest:     domain.ContestID(i.cfg.Contest),
				Lang:        langByExt[ext],
				Source:      src,
				Verdict:     domain.VerdictOK,
			})
		}
	}
	return subs, nil
}
