package timeutil

import (
	"os"
	"testing"
	"time"
)

func TestClinicLocation_Default(t *testing.T) {
	os.Unsetenv("CLINIC_TZ")
	loc := ClinicLocation()
	if loc == nil {
		t.Fatal("expected non-nil location")
	}
}

func TestStartOfClinicDay(t *testing.T) {
	os.Setenv("CLINIC_TZ", "America/Guatemala")
	ts := time.Date(2025, 11, 13, 10, 30, 0, 0, time.FixedZone("X", 0))
	sod := StartOfClinicDay(ts)
	if sod.Hour() != 0 || sod.Minute() != 0 {
		t.Fatalf("expected midnight, got %v", sod)
	}
}

func TestFormatClinicRFC3339(t *testing.T) {
	os.Setenv("CLINIC_TZ", "America/Guatemala")
	ts := time.Date(2025, 11, 13, 9, 0, 0, 0, time.UTC)
	s := FormatClinicRFC3339(ts)
	if len(s) == 0 {
		t.Fatal("expected non-empty rfc3339")
	}
}
