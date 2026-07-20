package ejudge_test

import (
	"testing"

	"github.com/lksh/scainer/internal/domain"
	"github.com/lksh/scainer/pkg/ejudge"
)

func TestLoadEnv_OK(t *testing.T) {
	t.Setenv(ejudge.EnvBaseURL, "https://ejudge.lksh.ru")
	t.Setenv(ejudge.EnvAPIKey, "tok")
	t.Setenv(ejudge.EnvTimeout, "5s")

	env, err := ejudge.LoadEnv()
	if err != nil {
		t.Fatal(err)
	}
	if env.Client.Timeout().String() != "5s" {
		t.Fatalf("timeout = %v", env.Client.Timeout())
	}
	if env.NormalizeLang("g++") != domain.LangCPP {
		t.Fatalf("g++ -> %q", env.NormalizeLang("g++"))
	}
	if env.NormalizeLang("java") != domain.LangJava {
		t.Fatalf("java -> %q", env.NormalizeLang("java"))
	}
	if env.NormalizeLang("clang++-32") != domain.LangCPP {
		t.Fatalf("clang++-32 -> %q", env.NormalizeLang("clang++-32"))
	}
	if env.NormalizeLang("g++14") != domain.LangCPP {
		t.Fatalf("g++14 -> %q", env.NormalizeLang("g++14"))
	}
	if env.NormalizeLang("python3.11") != domain.LangPython {
		t.Fatalf("python3.11 -> %q", env.NormalizeLang("python3.11"))
	}
	if env.NormalizeLang("fortran") != domain.Lang("fortran") {
		t.Fatalf("fortran -> %q", env.NormalizeLang("fortran"))
	}
}

func TestLoadEnv_MissingKey(t *testing.T) {
	t.Setenv(ejudge.EnvBaseURL, "https://ejudge.lksh.ru")
	t.Setenv(ejudge.EnvAPIKey, "")
	_, err := ejudge.LoadEnv()
	if err == nil {
		t.Fatal("expected error")
	}
}
