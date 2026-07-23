package jplag

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"scainer/internal/services/detect"
	"scainer/internal/domain"
	"scainer/pkg/store"
	pkgfs "scainer/pkg/store/fs"
)

type Detector struct {
	JarPath  string
	LangMap  map[domain.Lang]string
	Sources  store.SourcePather // nil только в тестах prepareSubmissions
	WorkRoot string             // пусто → MkdirTemp
}

var defaultLangMap = map[domain.Lang]string{
	domain.LangCPP:        "cpp",
	domain.LangPython:     "python3",
	domain.LangJava:       "java",
	domain.LangGo:         "go",
	domain.LangJavaScript: "javascript",
}

const (
	defaultJarPath  = "bin/jplag.jar"
	jplagWorkSubdir = ".jplag"
)

func New(jarPath string, fs *store.FS) (*Detector, error) {
	if fs == nil {
		return nil, fmt.Errorf("jplag: store.FS обязателен")
	}
	if jarPath == "" {
		jarPath = defaultJarPath
	}
	m := make(map[domain.Lang]string, len(defaultLangMap))
	for k, v := range defaultLangMap {
		m[k] = v
	}
	return &Detector{
		JarPath:  jarPath,
		LangMap:  m,
		Sources:  fs,
		WorkRoot: filepath.Join(fs.Root(), jplagWorkSubdir),
	}, nil
}

func (d *Detector) CheckRuntime() error {
	if d == nil {
		return fmt.Errorf("jplag: detector nil")
	}
	if _, err := exec.LookPath("java"); err != nil {
		return fmt.Errorf("jplag: java не найден в PATH (нужен JRE 17+): %w", err)
	}
	if _, err := os.Stat(d.JarPath); err != nil {
		return fmt.Errorf("jplag: jar %q не найден (make setup / jplag.jar_path): %w", d.JarPath, err)
	}
	return nil
}

var _ detect.Detector[domain.ProblemUnit] = (*Detector)(nil)

func (d *Detector) Name() string { return "jplag" }

func (d *Detector) AI() bool { return false }

func (d *Detector) Analyze(ctx context.Context, u domain.ProblemUnit) ([]domain.Signal, error) {
	subs := detect.LatestOkPerParticipant(u.Subs)
	if len(subs) < 2 {
		return nil, nil
	}

	jplagLang, ok := d.LangMap[u.Lang]
	if !ok {
		return nil, fmt.Errorf("jplag: неподдерживаемый язык %q", u.Lang)
	}
	ext, ok := domain.FileExt[u.Lang]
	if !ok {
		return nil, fmt.Errorf("jplag: нет расширения файла для языка %q", u.Lang)
	}

	workDir, cleanup, err := d.prepareWorkDir(u, subs[0].Contest)
	if err != nil {
		return nil, err
	}
	if cleanup != nil {
		defer cleanup()
	}

	byDir, err := prepareSubmissions(workDir, subs, ext, d.Sources)
	if err != nil {
		return nil, err
	}

	jarAbs, err := filepath.Abs(d.JarPath)
	if err != nil {
		return nil, err
	}
	subsAbs, err := filepath.Abs(filepath.Join(workDir, "submissions"))
	if err != nil {
		return nil, err
	}

	args := buildArgs(jarAbs, subsAbs, jplagLang)
	cmd := exec.CommandContext(ctx, "java", args...)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("jplag: java: %w\n%s", err, truncateBytes(out, 800))
	}

	resultPath := filepath.Join(workDir, "result.zip")
	if _, err := os.Stat(resultPath); err != nil {
		return nil, fmt.Errorf("jplag: нет result.zip после запуска: %w", err)
	}

	overview, comps, err := parseResult(resultPath)
	if err != nil {
		return nil, err
	}
	return signalsFromResult(u.Problem, overview, comps, byDir), nil
}

func (d *Detector) prepareWorkDir(u domain.ProblemUnit, contest domain.ContestID) (workDir string, cleanup func(), err error) {
	if d.WorkRoot == "" {
		tempDir, err := os.MkdirTemp("", "scainer-jplag-*")
		if err != nil {
			return "", nil, err
		}
		return tempDir, func() { _ = os.RemoveAll(tempDir) }, nil
	}

	workDir = filepath.Join(
		d.WorkRoot,
		pkgfs.SanitizeFileName(string(contest)),
		pkgfs.SanitizeFileName(string(u.Problem)),
		pkgfs.SanitizeFileName(string(u.Lang)),
	)
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return "", nil, err
	}
	// Параллельные Analyze / повторный импорт не должны делить result.zip / submissions.
	runDir, err := os.MkdirTemp(workDir, "run-*")
	if err != nil {
		return "", nil, err
	}
	return runDir, func() { _ = os.RemoveAll(runDir) }, nil
}

func prepareSubmissions(
	workDir string,
	subs []domain.Submission,
	ext string,
	sources store.SourcePather,
) (map[string]domain.Submission, error) {
	submissionsDir := filepath.Join(workDir, "submissions")
	if err := os.RemoveAll(submissionsDir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(submissionsDir, 0o755); err != nil {
		return nil, err
	}

	byDir := make(map[string]domain.Submission, len(subs))
	for _, sub := range subs {
		dirName := pkgfs.SanitizeFileName(string(sub.ID))
		dir := filepath.Join(submissionsDir, dirName)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		dest := filepath.Join(dir, "main"+ext)
		if err := placeSource(dest, sub, sources); err != nil {
			return nil, err
		}
		byDir[dirName] = sub
	}
	return byDir, nil
}

func placeSource(dest string, sub domain.Submission, sources store.SourcePather) error {
	if sources != nil {
		if src, ok := sources.SourcePath(sub); ok {
			abs, err := filepath.Abs(src)
			if err != nil {
				return err
			}
			if err := os.Symlink(abs, dest); err == nil {
				return nil
			}
			// Symlink недоступен (напр. Windows без привилегий) — пишем байты.
		}
	}
	return os.WriteFile(dest, sub.Source, 0o644)
}

func signalsFromResult(
	problem domain.ProblemID,
	ov *jplagOverview,
	comps map[string]jplagComparison,
	byDir map[string]domain.Submission,
) []domain.Signal {
	var out []domain.Signal
	for _, top := range ov.TopComparisons {
		subA, okA := byDir[top.FirstSubmission]
		subB, okB := byDir[top.SecondSubmission]
		if !okA || !okB {
			continue
		}
		score := top.avgSimilarity()
		if score <= 0 {
			continue // AVG=0 — шум JPlag topComparisons, не сигнал
		}
		ev := domain.Evidence{
			Kind:        "jplag_match",
			Description: fmt.Sprintf("JPlag similarity AVG=%.3f (%s vs %s)", score, subA.Participant, subB.Participant),
		}
		name := comparisonFileName(*ov, top.FirstSubmission, top.SecondSubmission)
		if c, ok := comps[name]; ok {
			for _, m := range c.Matches {
				ev.Spans = append(ev.Spans,
					domain.Span{Submission: subA.ID, StartLine: m.Start1, EndLine: m.End1},
					domain.Span{Submission: subB.ID, StartLine: m.Start2, EndLine: m.End2},
				)
			}
		}
		out = append(out, domain.Signal{
			Detector: "jplag",
			Subject:  domain.NewPairSubject(subA.Contest, problem, subA.Participant, subB.Participant),
			Score:    score,
			Evidence: []domain.Evidence{ev},
		})
	}
	return out
}

// buildArgs — аргументы java -jar … для одного Analyze.
// --normalize только для языков, где JPlag это поддерживает (cpp/java).
// -n -1: не резать topComparisons (дефолт 500 тесен на больших параллелях).
// --cluster-skip: кластеры в scainer не используются.
func buildArgs(jarAbs, subsAbs, jplagLang string) []string {
	args := []string{
		"-jar", jarAbs, subsAbs,
		"-l", jplagLang,
		"-r", "result",
		"-M", "RUN",
		"-n", "-1",
		"--cluster-skip",
	}
	if normalizeSupported(jplagLang) {
		args = append(args, "--normalize")
	}
	return args
}

func normalizeSupported(jplagLang string) bool {
	return jplagLang == "cpp" || jplagLang == "java"
}

func truncateBytes(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "…"
}
