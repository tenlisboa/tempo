package scheduler

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/tenlisboa/tempo/internal/store"
)

func (s *Scheduler) recoverMissed(st store.Store, sleepStart, now time.Time) []*store.Task {
	tasks, err := st.ListTasks()
	if err != nil {
		log.Printf("watchdog: failed to list tasks: %v", err)
		return nil
	}

	for _, task := range tasks {
		if !task.Enabled {
			continue
		}
		if task.Running {
			continue
		}
		if shouldHaveRun(task, sleepStart, now) {
			log.Printf("watchdog: running missed task %q (%s)", task.Name, task.ID)
			ctx, cancel := context.WithTimeout(s.ctx, 10*time.Minute)
			s.runner.Run(ctx, task, "scheduled")
			cancel()
		}
	}
	return tasks
}

func shouldHaveRun(task *store.Task, sleepStart, now time.Time) bool {
	switch task.ScheduleType {
	case store.ScheduleOnce:
		return shouldHaveRunOnce(task, sleepStart, now)
	case store.ScheduleInterval:
		return shouldHaveRunInterval(task, sleepStart, now)
	case store.ScheduleDaily:
		return shouldHaveRunDaily(task, sleepStart, now)
	case store.ScheduleWeekly:
		return shouldHaveRunWeekly(task, sleepStart, now)
	case store.ScheduleCron:
		return shouldHaveRunCron(task, sleepStart, now)
	default:
		return false
	}
}

func shouldHaveRunOnce(task *store.Task, sleepStart, now time.Time) bool {
	t, err := time.Parse(time.RFC3339, task.ScheduleExpr)
	if err != nil {
		return false
	}
	if task.LastRunAt != nil {
		return false
	}
	return t.After(sleepStart) && t.Before(now)
}

func shouldHaveRunInterval(task *store.Task, sleepStart, now time.Time) bool {
	d, err := time.ParseDuration(task.ScheduleExpr)
	if err != nil || d <= 0 {
		return false
	}
	if task.LastRunAt == nil {
		return true
	}
	nextDue := task.LastRunAt.Add(d)
	return nextDue.After(sleepStart) && nextDue.Before(now)
}

func shouldHaveRunDaily(task *store.Task, sleepStart, now time.Time) bool {
	parts := splitHHMM(task.ScheduleExpr)
	if parts == nil {
		return false
	}
	hour, min := parts[0], parts[1]

	day := sleepStart.Truncate(24 * time.Hour)
	end := now.Add(24 * time.Hour)
	for day.Before(end) {
		candidate := time.Date(day.Year(), day.Month(), day.Day(), hour, min, 0, 0, now.Location())
		if candidate.After(sleepStart) && candidate.Before(now) {
			return true
		}
		day = day.Add(24 * time.Hour)
	}
	return false
}

func shouldHaveRunWeekly(task *store.Task, sleepStart, now time.Time) bool {
	parts := strings.SplitN(task.ScheduleExpr, ":", 2)
	if len(parts) != 2 {
		return false
	}
	weekday, err := parseWeekday(parts[0])
	if err != nil {
		return false
	}
	hhmm := splitHHMM(parts[1])
	if hhmm == nil {
		return false
	}
	hour, min := hhmm[0], hhmm[1]

	day := sleepStart.Truncate(24 * time.Hour)
	end := now.Add(7 * 24 * time.Hour)
	for day.Before(end) {
		if day.Weekday() == weekday {
			candidate := time.Date(day.Year(), day.Month(), day.Day(), hour, min, 0, 0, now.Location())
			if candidate.After(sleepStart) && candidate.Before(now) {
				return true
			}
		}
		day = day.Add(24 * time.Hour)
	}
	return false
}

func shouldHaveRunCron(task *store.Task, sleepStart, now time.Time) bool {
	sched, err := cron.ParseStandard(task.ScheduleExpr)
	if err != nil {
		return false
	}
	next := sched.Next(sleepStart)
	return next.Before(now)
}

func splitHHMM(s string) []int {
	h, m, err := parseHHMM(s)
	if err != nil {
		return nil
	}
	return []int{h, m}
}
