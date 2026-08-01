package domain

import "testing"

func TestNeedsStatusRefresh(t *testing.T) {
	if !NeedsStatusRefresh(VerdictPR) || !NeedsStatusRefresh(VerdictRU) {
		t.Fatal("PR/RU")
	}
	if NeedsStatusRefresh(VerdictOK) {
		t.Fatal("OK final")
	}
}
