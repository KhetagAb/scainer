package domain

import "testing"

func TestNeedsStatusRefresh(t *testing.T) {
	if !NeedsStatusRefresh(VerdictPR) || !NeedsStatusRefresh(VerdictRU) || !NeedsStatusRefresh(VerdictCF) {
		t.Fatal("PR/RU/CF")
	}
	if NeedsStatusRefresh(VerdictOK) {
		t.Fatal("OK final")
	}
}
