package notification

import (
	"context"
	"time"
)

// reminderHourUTC is the wall-clock hour (UTC) at which the daily reminder
// sweep runs. UTC is used so the schedule never shifts across daylight-saving
// transitions.
const reminderHourUTC = 8

// startupDelay lets the HTTP server settle before the first catch-up run.
const startupDelay = 10 * time.Second

// StartReminderScheduler launches the daily certification reminder sweep in a
// background goroutine. It runs once shortly after startup (so the engine is
// not silent until the first scheduled hour) and then every day at
// reminderHourUTC. It returns immediately and stops when ctx is cancelled.
func StartReminderScheduler(ctx context.Context, svc *Service) {
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(startupDelay):
		}
		svc.RunDailyReminders(ctx)

		for {
			timer := time.NewTimer(time.Until(nextRunAt(time.Now().UTC(), reminderHourUTC)))
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				svc.RunDailyReminders(ctx)
			}
		}
	}()
}

// nextRunAt returns the next instant at the given hour (in now's location)
// strictly after now. Recomputing this each iteration avoids the cumulative
// drift a fixed 24h ticker would suffer if a run is slow or the process pauses.
func nextRunAt(now time.Time, hour int) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}
