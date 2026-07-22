package ejudge

import (
	"testing"

	"scainer/internal/domain"
)

func TestNormalizeLang(t *testing.T) {
	cases := []struct {
		in   string
		want domain.Lang
	}{
		{"g++", domain.LangCPP},
		{"java", domain.LangJava},
		{"clang++-32", domain.LangCPP},
		{"g++14", domain.LangCPP},
		{"python3.11", domain.LangPython},
		{"fortran", domain.Lang("fortran")},
	}
	for _, tc := range cases {
		if got := NormalizeLang(tc.in); got != tc.want {
			t.Fatalf("%q -> %q, want %q", tc.in, got, tc.want)
		}
	}
}
