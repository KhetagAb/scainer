package cron

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"scainer/internal/services/contests"
	"scainer/pkg/auth"
)

const DefaultParallelsSchedule = "0 9 * * *"

var parallelsSchedulerParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

type ParallelsScheduler struct {
	contests *contests.Service
	list     contests.EjudgeContestLister
	login    string
}

func NewParallelsScheduler(
	contestsSvc *contests.Service,
	list contests.EjudgeContestLister,
	login string,
) *ParallelsScheduler {
	return &ParallelsScheduler{
		contests: contestsSvc,
		list:     list,
		login:    login,
	}
}

func ValidateParallelsSchedule(schedule string) error {
	schedule = strings.TrimSpace(schedule)
	if schedule == "" {
		schedule = DefaultParallelsSchedule
	}
	_, err := parallelsSchedulerParser.Parse(schedule)
	if err != nil {
		return fmt.Errorf("invalid cron schedule %q: %w", schedule, err)
	}
	return nil
}

func (s *ParallelsScheduler) tick(ctx context.Context) {
	ctx = auth.WithLogin(ctx, s.login)

	result, err := s.contests.ImportFromEjudge(ctx, s.list)
	if err != nil {
		log.Printf("parallels_scheduler: import err=%v login=%s", err, s.login)
		return
	}
	log.Printf(
		"parallels_scheduler: added=%d duplicates=%d skipped=%d login=%s",
		len(result.Added),
		len(result.Duplicates),
		len(result.Skipped),
		s.login,
	)
}

func StartParallelsScheduler(
	ctx context.Context,
	contestsSvc *contests.Service,
	list contests.EjudgeContestLister,
	schedule string,
	login string,
) (func(), error) {
	if login == "" {
		return nil, fmt.Errorf("parallels_scheduler: import_cron.masterlogin обязателен")
	}
	if list == nil {
		return nil, fmt.Errorf("parallels_scheduler: ejudge lister is not configured")
	}

	schedule = strings.TrimSpace(schedule)
	if schedule == "" {
		schedule = DefaultParallelsSchedule
	}
	if err := ValidateParallelsSchedule(schedule); err != nil {
		return nil, err
	}

	sched := NewParallelsScheduler(contestsSvc, list, login)
	runCtx, cancel := context.WithCancel(ctx)

	c := cron.New(cron.WithLocation(parallelsSchedulerLocation()))
	if _, err := c.AddFunc(schedule, func() { sched.tick(runCtx) }); err != nil {
		cancel()
		return nil, fmt.Errorf("parallels_scheduler schedule: %w", err)
	}
	c.Start()
	log.Printf("parallels_scheduler: schedule=%q location=%s login=%s", schedule, parallelsSchedulerLocation(), login)

	return func() {
		stopCtx := c.Stop()
		<-stopCtx.Done()
		cancel()
	}, nil
}

func parallelsSchedulerLocation() *time.Location {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.FixedZone("MSK", 3*3600)
	}
	return loc
}
