package jplag

import (
	"slices"
	"testing"
)

func TestBuildArgs_AlwaysIncludesReportAndClusterSkip(t *testing.T) {
	args := buildArgs("/jar.jar", "/subs", "python3")
	want := []string{
		"-jar", "/jar.jar", "/subs",
		"-l", "python3",
		"-r", "result",
		"-M", "RUN",
		"-n", "-1",
		"--cluster-skip",
	}
	if !slices.Equal(args, want) {
		t.Fatalf("got %#v\nwant %#v", args, want)
	}
}

func TestBuildArgs_NormalizeOnlyForCPPAndJava(t *testing.T) {
	for _, lang := range []string{"cpp", "java"} {
		args := buildArgs("/j.jar", "/s", lang)
		if !slices.Contains(args, "--normalize") {
			t.Fatalf("%s: expected --normalize in %#v", lang, args)
		}
	}
	for _, lang := range []string{"python3", "go", "javascript"} {
		args := buildArgs("/j.jar", "/s", lang)
		if slices.Contains(args, "--normalize") {
			t.Fatalf("%s: unexpected --normalize in %#v", lang, args)
		}
	}
}

func TestNormalizeSupported(t *testing.T) {
	if !normalizeSupported("cpp") || !normalizeSupported("java") {
		t.Fatal("cpp/java должны поддерживать normalize")
	}
	if normalizeSupported("python3") || normalizeSupported("go") {
		t.Fatal("python3/go не поддерживают normalize")
	}
}
