package services

import (
	"strings"
	"testing"
	"time"

	"github.com/naiba/bonds/internal/models"
)

func TestFormatReminderDateGregorian(t *testing.T) {
	r := &models.ContactReminder{CalendarType: "gregorian"}
	fireAt := time.Date(2026, 6, 14, 9, 0, 0, 0, time.UTC)

	if got := formatReminderDate(r, fireAt, false); got != "2026-06-14" {
		t.Errorf("Gregorian + alt off: got %q want 2026-06-14", got)
	}
	if got := formatReminderDate(r, fireAt, true); got != "2026-06-14" {
		t.Errorf("Gregorian + alt on (no lunar conversion expected): got %q want 2026-06-14", got)
	}
}

func TestFormatReminderDateLunarAltCalendarOff(t *testing.T) {
	r := &models.ContactReminder{CalendarType: "lunar"}
	fireAt := time.Date(2026, 9, 25, 9, 0, 0, 0, time.UTC)

	got := formatReminderDate(r, fireAt, false)
	if got != "2026-09-25" {
		t.Errorf("Lunar + alt off should fall back to Gregorian-only: got %q want 2026-09-25", got)
	}
	if strings.Contains(got, "农历") {
		t.Errorf("alt-calendar OFF must not produce lunar formatting: %q", got)
	}
}

func TestFormatReminderDateLunarAltCalendarOn(t *testing.T) {
	r := &models.ContactReminder{CalendarType: "lunar"}
	fireAt := time.Date(2026, 9, 25, 9, 0, 0, 0, time.UTC) // Mid-Autumn 2026

	got := formatReminderDate(r, fireAt, true)
	if !strings.HasPrefix(got, "农历") {
		t.Errorf("Lunar + alt on must lead with 农历: %q", got)
	}
	if !strings.Contains(got, "(2026-09-25)") {
		t.Errorf("Lunar + alt on must include Gregorian in parens: %q", got)
	}
}

func TestFormatReminderDateEmptyCalendarTypeTreatedAsGregorian(t *testing.T) {
	r := &models.ContactReminder{CalendarType: ""}
	fireAt := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	if got := formatReminderDate(r, fireAt, true); got != "2026-01-01" {
		t.Errorf("empty CalendarType must be treated as Gregorian: got %q", got)
	}
}
