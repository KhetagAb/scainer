package domain

type Unit interface{ unit() }

type StandaloneUnit struct{ Sub Submission }

type PairUnit struct{ A, B Submission }

type ProblemUnit struct {
	Problem ProblemID
	Lang    Lang
	Subs    []Submission
}

type ProblemParticipantUnit struct {
	Problem     ProblemID
	Participant ParticipantID
	Subs        []Submission // упорядочены по SubmittedAt
}

type ParticipantUnit struct {
	Participant ParticipantID
	Subs        []Submission
}

func (StandaloneUnit) unit()         {}
func (PairUnit) unit()               {}
func (ProblemUnit) unit()            {}
func (ProblemParticipantUnit) unit() {}
func (ParticipantUnit) unit()        {}

type PairKey struct{ A, B ParticipantID }

func NewPairKey(x, y ParticipantID) PairKey {
	if x <= y {
		return PairKey{A: x, B: y}
	}
	return PairKey{A: y, B: x}
}
