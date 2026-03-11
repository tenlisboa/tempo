package views

import (
	"fmt"
	"strings"
	"time"

	lnqcron "github.com/lnquy/cron"
	"github.com/tenlisboa/tempo/internal/store"
)

var cronDesc *lnqcron.ExpressionDescriptor

func init() {
	cronDesc, _ = lnqcron.NewDescriptor(lnqcron.Use24HourTimeFormat(true))
}

func formatScheduleHuman(schedType store.ScheduleType, expr string) string {
	switch schedType {
	case store.ScheduleInterval:
		d, err := time.ParseDuration(expr)
		if err != nil {
			return expr
		}
		return "every " + formatDuration(d)

	case store.ScheduleDaily:
		return "daily at " + expr

	case store.ScheduleWeekly:
		parts := strings.SplitN(expr, ":", 2)
		if len(parts) != 2 {
			return expr
		}
		day := strings.ToUpper(parts[0][:1]) + strings.ToLower(parts[0][1:])
		return day + " at " + parts[1]

	case store.ScheduleOnce:
		t, err := time.Parse(time.RFC3339, expr)
		if err != nil {
			return "once: " + expr
		}
		return "once at " + t.Format("Jan 2 15:04")

	case store.ScheduleCron:
		if cronDesc == nil {
			return "cron: " + expr
		}
		desc, err := cronDesc.ToDescription(expr, lnqcron.Locale_en)
		if err != nil {
			return "cron: " + expr
		}
		return desc
	}
	return expr
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	switch {
	case h > 0 && m > 0:
		return fmt.Sprintf("%dh %dm", h, m)
	case h == 1:
		return "1 hour"
	case h > 1:
		return fmt.Sprintf("%d hours", h)
	case m == 1:
		return "1 min"
	default:
		return fmt.Sprintf("%d min", m)
	}
}
