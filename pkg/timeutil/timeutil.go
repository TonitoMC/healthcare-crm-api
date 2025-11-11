package timeutil

import (
	"os"
	"sync"
	"time"
)

var (
	clinicLoc     *time.Location
	clinicLocOnce sync.Once
)

const defaultClinicTZ = "America/Guatemala"

// ClinicLocation returns the singleton time.Location for the clinic timezone.
// It reads CLINIC_TZ from environment (fallback to America/Guatemala).
func ClinicLocation() *time.Location {
	clinicLocOnce.Do(func() {
		tz := os.Getenv("CLINIC_TZ")
		if tz == "" {
			tz = defaultClinicTZ
		}
		loc, err := time.LoadLocation(tz)
		if err != nil {
			// Fallback robusto: UTC (último recurso)
			clinicLoc = time.UTC
			return
		}
		clinicLoc = loc
	})
	return clinicLoc
}

// NormalizeToClinic returns the same instant expressed in the clinic timezone.
func NormalizeToClinic(t time.Time) time.Time {
	return t.In(ClinicLocation())
}

// StartOfClinicDay returns the midnight at clinic timezone for the provided time.
func StartOfClinicDay(t time.Time) time.Time {
	loc := ClinicLocation()
	tt := t.In(loc)
	return time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, loc)
}

// ParseYMDToClinic parses YYYY-MM-DD into a time at midnight in clinic timezone.
func ParseYMDToClinic(ymd string) (time.Time, error) {
	d, err := time.Parse("2006-01-02", ymd)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, ClinicLocation()), nil
}

// FormatClinicRFC3339 returns RFC3339 string with clinic timezone offset.
func FormatClinicRFC3339(t time.Time) string {
	return t.In(ClinicLocation()).Format(time.RFC3339)
}

// TimeOfDayMinutes returns minutes since midnight for given time in its location.
func TimeOfDayMinutes(t time.Time) int {
	tt := t
	return tt.Hour()*60 + tt.Minute()
}
