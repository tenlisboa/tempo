package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/tenlisboa/tempo/internal/store"
)

type Scheduler struct {
	mu     sync.Mutex
	sched  gocron.Scheduler
	runner *Runner
	jobs   map[string]gocron.Job
}

func New(runner *Runner) (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}
	return &Scheduler{
		sched:  s,
		runner: runner,
		jobs:   make(map[string]gocron.Job),
	}, nil
}

func (s *Scheduler) Start() {
	s.sched.Start()
}

func (s *Scheduler) Stop() error {
	return s.sched.Shutdown()
}

func (s *Scheduler) LoadAll(tasks []*store.Task) {
	for _, t := range tasks {
		if t.Enabled {
			_ = s.Add(t)
		}
	}
}

func (s *Scheduler) Add(task *store.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	def, err := jobDef(task)
	if err != nil {
		return err
	}

	job, err := s.sched.NewJob(def, gocron.NewTask(func(t *store.Task) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		_, _ = s.runner.Run(ctx, t, "scheduled")
	}, task))
	if err != nil {
		return err
	}
	s.jobs[task.ID] = job
	return nil
}

func (s *Scheduler) Remove(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[taskID]
	if !ok {
		return nil
	}
	if err := s.sched.RemoveJob(job.ID()); err != nil {
		return err
	}
	delete(s.jobs, taskID)
	return nil
}

func (s *Scheduler) RunNow(ctx context.Context, task *store.Task) (*store.RunLog, error) {
	return s.runner.Run(ctx, task, "manual")
}

func jobDef(task *store.Task) (gocron.JobDefinition, error) {
	switch task.ScheduleType {
	case store.ScheduleOnce:
		t, err := time.Parse(time.RFC3339, task.ScheduleExpr)
		if err != nil {
			return nil, fmt.Errorf("invalid once time %q: %w", task.ScheduleExpr, err)
		}
		return gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(t)), nil

	case store.ScheduleInterval:
		d, err := time.ParseDuration(task.ScheduleExpr)
		if err != nil {
			return nil, fmt.Errorf("invalid interval %q: %w", task.ScheduleExpr, err)
		}
		return gocron.DurationJob(d), nil

	case store.ScheduleDaily:
		h, m, err := parseHHMM(task.ScheduleExpr)
		if err != nil {
			return nil, err
		}
		return gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(uint(h), uint(m), 0))), nil

	case store.ScheduleWeekly:
		parts := strings.SplitN(task.ScheduleExpr, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("weekly expr must be 'day:HH:MM', got %q", task.ScheduleExpr)
		}
		wd, err := parseWeekday(parts[0])
		if err != nil {
			return nil, err
		}
		h, m, err := parseHHMM(parts[1])
		if err != nil {
			return nil, err
		}
		return gocron.WeeklyJob(1, gocron.NewWeekdays(wd), gocron.NewAtTimes(gocron.NewAtTime(uint(h), uint(m), 0))), nil

	case store.ScheduleCron:
		return gocron.CronJob(task.ScheduleExpr, false), nil

	default:
		return nil, fmt.Errorf("unknown schedule type %q", task.ScheduleType)
	}
}

func parseHHMM(s string) (int, int, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected HH:MM, got %q", s)
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return h, m, nil
}

func parseWeekday(s string) (time.Weekday, error) {
	days := map[string]time.Weekday{
		"sun": time.Sunday, "mon": time.Monday, "tue": time.Tuesday,
		"wed": time.Wednesday, "thu": time.Thursday, "fri": time.Friday, "sat": time.Saturday,
	}
	if d, ok := days[strings.ToLower(s)]; ok {
		return d, nil
	}
	return 0, fmt.Errorf("unknown weekday %q", s)
}
