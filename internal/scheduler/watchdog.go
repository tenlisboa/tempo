package scheduler

import (
	"log"
	"time"

	"github.com/tenlisboa/tempo/internal/store"
)

const (
	watchdogInterval = 5 * time.Second
	sleepThreshold   = 30 * time.Second
)

func (s *Scheduler) StartWatchdog(st store.Store) {
	go func() {
		ticker := time.NewTicker(watchdogInterval)
		defer ticker.Stop()
		lastTick := time.Now()

		for {
			select {
			case <-s.ctx.Done():
				return
			case now := <-ticker.C:
				elapsed := now.Sub(lastTick)
				if elapsed > sleepThreshold {
					log.Printf("watchdog: detected time jump of %s (sleep/suspend), recovering missed jobs", elapsed.Round(time.Second))
					tasks := s.recoverMissed(st, lastTick, now)
					if tasks != nil {
						log.Printf("watchdog: resetting gocron timers for %d tasks", len(tasks))
						s.resetTimers(tasks)
					}
				}
				s.checkOverdue(st, now)
				lastTick = now
			}
		}
	}()
}
