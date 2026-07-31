package domain

type ProblemExample struct {
	Input  string `json:"input" yaml:"input"`
	Output string `json:"output" yaml:"output"`
}

type ProblemStatement struct {
	Contest   ContestID        `json:"contest" yaml:"contest"`
	Problem   ProblemID        `json:"problem" yaml:"problem"`
	Title     string           `json:"title" yaml:"title"`
	Statement string           `json:"statement" yaml:"statement"`
	Examples  []ProblemExample `json:"examples,omitempty" yaml:"examples,omitempty"`
}

type ProblemStatementExplain struct {
	Problem   ProblemID `json:"problem"`
	Title     string    `json:"title"`
	Statement string    `json:"statement"`
}
