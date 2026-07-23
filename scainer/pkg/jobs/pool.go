// Package jobs — пул фоновых задач с прогрессом (progress) и подпиской на статус.
// Домен (import/detect) сюда не входит: только Func + ограниченная параллельность.
// Состояние только в памяти процесса — рестарт теряет историю job'ов.
package jobs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"scainer/pkg/progress"
)

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
)

type State struct {
	ID         string
	Status     Status
	Progress   progress.Event
	Err        string
	StartedAt  time.Time
	FinishedAt time.Time
}

// Func — работа job'а; progress.Report(ctx, ...) пишется в State через Reporter, вшитый Pool'ом.
type Func func(ctx context.Context) error

const maxHistory = 200 // завершённых; queued/running не вытесняются

type entry struct {
	state     State
	subs      map[int]chan State
	nextSubID int
}

type Pool struct {
	sem chan struct{}

	mu    sync.Mutex
	jobs  map[string]*entry
	order []string
}

func NewPool(maxConcurrent int) *Pool {
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}
	return &Pool{
		sem:  make(chan struct{}, maxConcurrent),
		jobs: make(map[string]*entry),
	}
}

// Submit сразу возвращает ID. Выполнение не наследует отмену ctx (HTTP-запрос умрёт после
// ответа); значения контекста — через WithoutCancel.
func (p *Pool) Submit(ctx context.Context, fn Func) string {
	id := newID()
	e := &entry{
		state: State{ID: id, Status: StatusQueued},
		subs:  make(map[int]chan State),
	}

	p.mu.Lock()
	p.jobs[id] = e
	p.order = append(p.order, id)
	p.evictLocked()
	p.mu.Unlock()

	runCtx := context.WithoutCancel(ctx)
	go p.run(runCtx, id, fn)

	return id
}

func (p *Pool) run(ctx context.Context, id string, fn Func) {
	p.sem <- struct{}{}
	defer func() { <-p.sem }()

	p.update(id, func(s *State) {
		s.Status = StatusRunning
		s.StartedAt = time.Now()
	})

	reportCtx := progress.With(ctx, func(e progress.Event) {
		p.update(id, func(s *State) { s.Progress = e })
	})

	err := fn(reportCtx)

	p.update(id, func(s *State) {
		s.FinishedAt = time.Now()
		if err != nil {
			s.Status = StatusFailed
			s.Err = err.Error()
		} else {
			s.Status = StatusSucceeded
		}
	})
}

func (p *Pool) Get(id string) (State, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	e, ok := p.jobs[id]
	if !ok {
		return State{}, false
	}
	return e.state, true
}

// Subscribe: сразу текущий снимок, дальше обновления. cancel отписывает.
func (p *Pool) Subscribe(id string) (events <-chan State, cancel func(), ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	e, exists := p.jobs[id]
	if !exists {
		return nil, func() {}, false
	}

	ch := make(chan State, 1)
	ch <- e.state
	subID := e.nextSubID
	e.nextSubID++
	e.subs[subID] = ch

	cancelFn := func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		if e2, ok := p.jobs[id]; ok {
			delete(e2.subs, subID)
		}
	}
	return ch, cancelFn, true
}

func (p *Pool) update(id string, mutate func(*State)) {
	p.mu.Lock()
	defer p.mu.Unlock()

	e, ok := p.jobs[id]
	if !ok {
		return
	}
	mutate(&e.state)
	snapshot := e.state
	for _, ch := range e.subs {
		select {
		case ch <- snapshot:
		default:
			select {
			case <-ch:
			default:
			}
			select {
			case ch <- snapshot:
			default:
			}
		}
	}
}

func (p *Pool) evictLocked() {
	for len(p.order) > maxHistory {
		oldest := p.order[0]
		e, ok := p.jobs[oldest]
		if ok && (e.state.Status == StatusQueued || e.state.Status == StatusRunning) {
			break
		}
		p.order = p.order[1:]
		delete(p.jobs, oldest)
	}
}

func newID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("job_%s", hex.EncodeToString(b[:]))
}
