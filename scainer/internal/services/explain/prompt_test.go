package explain

import "testing"

func TestTrimStatementForExplain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "empty",
			in:   "",
			want: "",
		},
		{
			name: "story only",
			in:   "  Ася любит числа.  ",
			want: "Ася любит числа.",
		},
		{
			name: "keeps input format",
			in:   "Сюжет.\n\nФормат входных данных\nПервая строка: n",
			want: "Сюжет.\n\nФормат входных данных\nПервая строка: n",
		},
		{
			name: "cuts at output format",
			in:   "Сюжет.\n\nФормат входных данных\nn\n\nФормат выходных данных\nответ",
			want: "Сюжет.\n\nФормат входных данных\nn",
		},
		{
			name: "collapsed pdf spacing",
			in:   "Сюжет. Форматвходныхданных n. Форматвыходныхданных k",
			want: "Сюжет. Форматвходныхданных n.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := trimStatementForExplain(tt.in); got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}
