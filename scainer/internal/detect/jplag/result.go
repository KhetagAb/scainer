package jplag

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"path"
)

func parseResult(zipPath string) (*jplagOverview, map[string]jplagComparison, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, nil, fmt.Errorf("jplag: open result: %w", err)
	}
	defer zr.Close()

	byName := make(map[string]*zip.File, len(zr.File))
	for _, f := range zr.File {
		byName[path.Clean(f.Name)] = f
	}

	ovFile, ok := byName["overview.json"]
	if !ok {
		return nil, nil, fmt.Errorf("jplag: overview.json не найден в архиве")
	}
	ovData, err := readZipFile(ovFile)
	if err != nil {
		return nil, nil, err
	}
	var overview jplagOverview
	if err := json.Unmarshal(ovData, &overview); err != nil {
		return nil, nil, fmt.Errorf("jplag: overview.json: %w", err)
	}

	comps := make(map[string]jplagComparison)
	seen := make(map[string]struct{})
	for _, row := range overview.TopComparisons {
		name := comparisonFileName(overview, row.FirstSubmission, row.SecondSubmission)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		f, ok := byName[path.Clean(name)]
		if !ok {
			return nil, nil, fmt.Errorf("jplag: файл сравнения %q не найден", name)
		}
		raw, err := readZipFile(f)
		if err != nil {
			return nil, nil, err
		}
		var c jplagComparison
		if err := json.Unmarshal(raw, &c); err != nil {
			return nil, nil, fmt.Errorf("jplag: %s: %w", name, err)
		}
		comps[name] = c
	}
	return &overview, comps, nil
}

func comparisonFileName(ov jplagOverview, a, b string) string {
	if ov.SubmissionIDsToComparisonFileName == nil {
		return ""
	}
	if m, ok := ov.SubmissionIDsToComparisonFileName[a]; ok {
		if name, ok := m[b]; ok {
			return name
		}
	}
	if m, ok := ov.SubmissionIDsToComparisonFileName[b]; ok {
		if name, ok := m[a]; ok {
			return name
		}
	}
	return ""
}

func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

type jplagOverview struct {
	TopComparisons                    []jplagTopComparison         `json:"top_comparisons"`
	SubmissionIDsToComparisonFileName map[string]map[string]string `json:"submission_ids_to_comparison_file_name"`
	TotalComparisons                  int                          `json:"total_comparisons"`
}

type jplagTopComparison struct {
	FirstSubmission  string             `json:"first_submission"`
	SecondSubmission string             `json:"second_submission"`
	Similarities     map[string]float64 `json:"similarities"`
}

func (t jplagTopComparison) avgSimilarity() float64 {
	if t.Similarities == nil {
		return 0
	}
	if v, ok := t.Similarities["AVG"]; ok {
		return v
	}
	if v, ok := t.Similarities["MAX"]; ok {
		return v
	}
	return 0
}

type jplagComparison struct {
	ID1              string             `json:"id1"`
	ID2              string             `json:"id2"`
	Similarities     map[string]float64 `json:"similarities"`
	FirstSimilarity  float64            `json:"first_similarity"`
	SecondSimilarity float64            `json:"second_similarity"`
	Matches          []jplagMatch       `json:"matches"`
}

type jplagMatch struct {
	File1  string `json:"file1"`
	File2  string `json:"file2"`
	Start1 int    `json:"start1"`
	End1   int    `json:"end1"`
	Start2 int    `json:"start2"`
	End2   int    `json:"end2"`
	Tokens int    `json:"tokens"`
}
