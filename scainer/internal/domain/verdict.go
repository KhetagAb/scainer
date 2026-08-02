package domain

type Verdict string

const (
	VerdictOK      Verdict = "OK"
	VerdictTL      Verdict = "TL"
	VerdictML      Verdict = "ML"
	VerdictWA      Verdict = "WA"
	VerdictPR      Verdict = "PR"
	VerdictCF      Verdict = "CF"
	VerdictCE      Verdict = "CE"
	VerdictDQ      Verdict = "DQ"
	VerdictRJ      Verdict = "RJ"
	VerdictRT      Verdict = "RT"
	VerdictIG      Verdict = "IG"
	VerdictPE      Verdict = "PE"
	VerdictRU      Verdict = "RU"
	VerdictUnknown Verdict = "UNKNOWN"
)

func NeedsStatusRefresh(v Verdict) bool {
	return v == VerdictPR || v == VerdictRU || v == VerdictCF
}
