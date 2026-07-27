package lksh

type Lesson struct {
	ContestURL    string `json:"contest_url"`
	StatementsURL string `json:"statements_url"`
}

func CollectLessons(layout any) []Lesson {
	var out []Lesson
	collectLessons(layout, &out)
	return out
}

func collectLessons(node any, out *[]Lesson) {
	switch v := node.(type) {
	case map[string]any:
		if cfg, ok := v["config"].(map[string]any); ok {
			if lessons, ok := cfg["lessons"].([]any); ok {
				for _, item := range lessons {
					lm, ok := item.(map[string]any)
					if !ok {
						continue
					}
					lesson := Lesson{
						ContestURL:    stringField(lm, "contest_url"),
						StatementsURL: stringField(lm, "statements_url"),
					}
					if lesson.StatementsURL != "" {
						*out = append(*out, lesson)
					}
				}
			}
		}
		for _, child := range v {
			collectLessons(child, out)
		}
	case []any:
		for _, child := range v {
			collectLessons(child, out)
		}
	}
}

func stringField(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}
