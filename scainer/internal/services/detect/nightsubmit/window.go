package nightsubmit

import (
	"fmt"
	"time"
)

const (
	DetectorName = "night-submit"

	windowStartHour = 22
	windowStartMin  = 30
	windowEndHour   = 6
	windowEndMin    = 30
)

var msk = time.FixedZone("MSK", 3*60*60)

func WindowLabel() string {
	return fmt.Sprintf("%02d:%02d–%02d:%02d", windowStartHour, windowStartMin, windowEndHour, windowEndMin)
}

func windowStartMinutes() int { return windowStartHour*60 + windowStartMin }

func windowEndMinutes() int { return windowEndHour*60 + windowEndMin }

func InNightWindow(t time.Time) bool {
	if t.IsZero() {
		return false
	}
	local := t.In(msk)
	mins := local.Hour()*60 + local.Minute()
	return mins >= windowStartMinutes() || mins <= windowEndMinutes()
}

func EvidenceDescription(submittedAt time.Time) string {
	return fmt.Sprintf(
		"Отправлено в %s МСК · ночной интервал %s",
		formatMSK(submittedAt),
		WindowLabel(),
	)
}

func formatMSK(t time.Time) string {
	return t.In(msk).Format("15:04")
}
