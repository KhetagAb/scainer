package domain

type ParallelID string

type Parallel struct {
	ID       ParallelID
	Contests []ContestID
}
