package domain

type ProblemExample struct {
	Input  string `json:"input" yaml:"input"`
	Output string `json:"output" yaml:"output"`
}

type ProblemStatement struct {
	Contest      ContestID        `json:"contest" yaml:"contest"`
	Problem      ProblemID        `json:"problem" yaml:"problem"`
	Title        string           `json:"title" yaml:"title"`
	RawStatement string           `json:"rawStatement" yaml:"statement"`
	Examples     []ProblemExample `json:"examples,omitempty" yaml:"examples,omitempty"`
}

type FormalizationSource string

const (
	FormalizationSourceAI     FormalizationSource = "ai"
	FormalizationSourceManual FormalizationSource = "manual"
)

type ProblemStatementFormalization struct {
	Contest   ContestID           `json:"contest" yaml:"contest"`
	Problem   ProblemID           `json:"problem" yaml:"problem"`
	Title     string              `json:"title" yaml:"title"`
	Statement string              `json:"statement" yaml:"statement"`
	Source    FormalizationSource `json:"source" yaml:"source"`
	Model     string              `json:"model,omitempty" yaml:"model,omitempty"`
	Prompt    string              `json:"prompt,omitempty" yaml:"prompt,omitempty"`
	UpdatedAt string              `json:"updated_at" yaml:"updated_at"`
}

type ProblemStatementExplain struct {
	Problem   ProblemID           `json:"problem"`
	Title     string              `json:"title"`
	Statement string              `json:"statement"`
	Source    FormalizationSource `json:"source"`
	Model     string              `json:"model,omitempty"`
	Prompt    string              `json:"prompt,omitempty"`
}
