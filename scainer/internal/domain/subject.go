package domain

import (
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

type SubjectKind string

const (
	SubjectSubmission         SubjectKind = "submission"
	SubjectPair               SubjectKind = "pair"
	SubjectParticipantProblem SubjectKind = "participant_problem"
	SubjectParticipant        SubjectKind = "participant"
)

type Subject interface {
	subject()
	Kind() SubjectKind
	Key() string
}

type SubmissionSubject struct {
	Contest    ContestID    `json:"contest" bson:"contest"`
	Submission SubmissionID `json:"submission" bson:"submission"`
}

type PairSubject struct {
	Contest ContestID     `json:"contest" bson:"contest"`
	Problem ProblemID     `json:"problem" bson:"problem"`
	A       ParticipantID `json:"a" bson:"a"`
	B       ParticipantID `json:"b" bson:"b"`
}

type ParticipantProblemSubject struct {
	Contest     ContestID     `json:"contest" bson:"contest"`
	Problem     ProblemID     `json:"problem" bson:"problem"`
	Participant ParticipantID `json:"participant" bson:"participant"`
}

type ParticipantSubject struct {
	Participant ParticipantID `json:"participant" bson:"participant"`
}

func (SubmissionSubject) subject()         {}
func (PairSubject) subject()               {}
func (ParticipantProblemSubject) subject() {}
func (ParticipantSubject) subject()        {}

func (SubmissionSubject) Kind() SubjectKind         { return SubjectSubmission }
func (PairSubject) Kind() SubjectKind               { return SubjectPair }
func (ParticipantProblemSubject) Kind() SubjectKind { return SubjectParticipantProblem }
func (ParticipantSubject) Kind() SubjectKind        { return SubjectParticipant }

func (s SubmissionSubject) Key() string {
	return fmt.Sprintf("%s|%s|%s", s.Kind(), s.Contest, s.Submission)
}

func (s PairSubject) Key() string {
	a, b := s.A, s.B
	if b < a {
		a, b = b, a
	}
	return fmt.Sprintf("%s|%s|%s|%s|%s", s.Kind(), s.Contest, s.Problem, a, b)
}

func (s ParticipantProblemSubject) Key() string {
	return fmt.Sprintf("%s|%s|%s|%s", s.Kind(), s.Contest, s.Problem, s.Participant)
}

func (s ParticipantSubject) Key() string {
	return fmt.Sprintf("%s|%s", s.Kind(), s.Participant)
}

func NewSubmissionSubject(contest ContestID, submission SubmissionID) SubmissionSubject {
	return SubmissionSubject{Contest: contest, Submission: submission}
}

func NewPairSubject(contest ContestID, problem ProblemID, a, b ParticipantID) PairSubject {
	return PairSubject{Contest: contest, Problem: problem, A: a, B: b}
}

func NewParticipantProblemSubject(contest ContestID, problem ProblemID, p ParticipantID) ParticipantProblemSubject {
	return ParticipantProblemSubject{Contest: contest, Problem: problem, Participant: p}
}

func NewParticipantSubject(p ParticipantID) ParticipantSubject {
	return ParticipantSubject{Participant: p}
}

// encoding/json не пишет тип интерфейса — добавляем дискриминатор kind.
func (s SubmissionSubject) MarshalJSON() ([]byte, error) {
	type out SubmissionSubject
	return json.Marshal(struct {
		Kind SubjectKind `json:"kind"`
		out
	}{Kind: s.Kind(), out: out(s)})
}

func (s PairSubject) MarshalJSON() ([]byte, error) {
	type out PairSubject
	return json.Marshal(struct {
		Kind SubjectKind `json:"kind"`
		out
	}{Kind: s.Kind(), out: out(s)})
}

func (s ParticipantProblemSubject) MarshalJSON() ([]byte, error) {
	type out ParticipantProblemSubject
	return json.Marshal(struct {
		Kind SubjectKind `json:"kind"`
		out
	}{Kind: s.Kind(), out: out(s)})
}

func (s ParticipantSubject) MarshalJSON() ([]byte, error) {
	type out ParticipantSubject
	return json.Marshal(struct {
		Kind SubjectKind `json:"kind"`
		out
	}{Kind: s.Kind(), out: out(s)})
}

func SubjectKey(s Subject) string { return s.Key() }

// Через map, не bson:",inline": драйвер не инлайнит анонимно встроенный неэкспортируемый тип.
func marshalSubjectBSON(kind SubjectKind, fields any) ([]byte, error) {
	raw, err := bson.Marshal(fields)
	if err != nil {
		return nil, err
	}
	var m bson.M
	if err := bson.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	m["kind"] = kind
	return bson.Marshal(m)
}

func (s SubmissionSubject) MarshalBSON() ([]byte, error) {
	type out SubmissionSubject
	return marshalSubjectBSON(s.Kind(), out(s))
}

func (s PairSubject) MarshalBSON() ([]byte, error) {
	type out PairSubject
	return marshalSubjectBSON(s.Kind(), out(s))
}

func (s ParticipantProblemSubject) MarshalBSON() ([]byte, error) {
	type out ParticipantProblemSubject
	return marshalSubjectBSON(s.Kind(), out(s))
}

func (s ParticipantSubject) MarshalBSON() ([]byte, error) {
	type out ParticipantSubject
	return marshalSubjectBSON(s.Kind(), out(s))
}

// Subject — интерфейс; BSON-драйвер не конструирует его сам.
func unmarshalSubjectBSON(raw bson.Raw) (Subject, error) {
	var disc struct {
		Kind SubjectKind `bson:"kind"`
	}
	if err := bson.Unmarshal(raw, &disc); err != nil {
		return nil, err
	}
	switch disc.Kind {
	case SubjectSubmission:
		var v SubmissionSubject
		if err := bson.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return v, nil
	case SubjectPair:
		var v PairSubject
		if err := bson.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return v, nil
	case SubjectParticipantProblem:
		var v ParticipantProblemSubject
		if err := bson.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return v, nil
	case SubjectParticipant:
		var v ParticipantSubject
		if err := bson.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return v, nil
	default:
		return nil, fmt.Errorf("domain: неизвестный kind субъекта %q", disc.Kind)
	}
}
