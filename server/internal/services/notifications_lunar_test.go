package services

import (
	"testing"
	"time"

	calendarPkg "github.com/naiba/bonds/internal/calendar"
	"github.com/naiba/bonds/internal/dto"
	"github.com/naiba/bonds/internal/models"
	"github.com/naiba/bonds/internal/testutil"
)

// TestScheduleAllContactRemindersUsesCalendarConverterForLunar pins the bug
// in notifications.go's ScheduleAllContactReminders: when activating or
// verifying a channel, the function used the cached r.Month/r.Day (which are
// the Gregorian projection captured the year the lunar reminder was first
// created) and naively assembled time.Date(now.Year(), cachedMonth, cachedDay,
// ...). Lunar dates drift year over year, so reusing the old projection
// silently scheduled the reminder on the wrong day every subsequent year.
// The reminder.go calcInitialSchedule path handles this correctly; this test
// guards the notifications.go path against drifting apart again.
func TestScheduleAllContactRemindersUsesCalendarConverterForLunar(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.TestJWTConfig()
	auth := NewAuthService(db, cfg)
	resp, err := auth.Register(dto.RegisterRequest{
		FirstName: "Lunar",
		LastName:  "Owner",
		Email:     "lunar-owner@example.com",
		Password:  "password123",
	}, "zh")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	vault := models.Vault{Name: "Vault", AccountID: resp.User.AccountID}
	if err := db.Create(&vault).Error; err != nil {
		t.Fatalf("create vault: %v", err)
	}
	if err := db.Create(&models.UserVault{
		UserID: resp.User.ID, VaultID: vault.ID, Permission: 100,
	}).Error; err != nil {
		t.Fatalf("create user-vault: %v", err)
	}

	first := "Lunar"
	last := "Anchored"
	contact := models.Contact{VaultID: vault.ID, FirstName: &first, LastName: &last}
	if err := db.Create(&contact).Error; err != nil {
		t.Fatalf("create contact: %v", err)
	}

	// Lunar 8/15 (Mid-Autumn). Cache stale Gregorian projection from years
	// ago — picking a day-of-year unlikely to coincide with this year's
	// actual Mid-Autumn so the test fails clearly when the bug is back.
	staleMonth, staleDay := 9, 1
	origMonth, origDay := 8, 15
	reminder := models.ContactReminder{
		ContactID:     contact.ID,
		Label:         "Mid-Autumn",
		Type:          "recurring_year",
		CalendarType:  "lunar",
		OriginalMonth: &origMonth,
		OriginalDay:   &origDay,
		Month:         &staleMonth,
		Day:           &staleDay,
	}
	if err := db.Create(&reminder).Error; err != nil {
		t.Fatalf("create reminder: %v", err)
	}

	now := time.Now()
	channel := models.UserNotificationChannel{
		UserID:     &resp.User.ID,
		Type:       "email",
		Content:    "lunar-owner@example.com",
		Active:     true,
		VerifiedAt: &now,
	}
	if err := db.Create(&channel).Error; err != nil {
		t.Fatalf("create channel: %v", err)
	}

	svc := NewNotificationService(db)
	if err := svc.ScheduleAllContactReminders(channel.ID, resp.User.ID); err != nil {
		t.Fatalf("ScheduleAllContactReminders: %v", err)
	}

	var scheduled models.ContactReminderScheduled
	if err := db.Where("user_notification_channel_id = ? AND contact_reminder_id = ?", channel.ID, reminder.ID).
		First(&scheduled).Error; err != nil {
		t.Fatalf("find scheduled row: %v", err)
	}

	converter, ok := calendarPkg.Get(calendarPkg.Lunar)
	if !ok {
		t.Fatal("lunar converter not registered")
	}
	expectedG, err := converter.NextOccurrence(calendarPkg.DateInfo{Day: origDay, Month: origMonth}, now.AddDate(0, 0, -1))
	if err != nil {
		t.Fatalf("converter.NextOccurrence: %v", err)
	}
	sLocal := scheduled.ScheduledAt
	if sLocal.Day() != expectedG.Day || int(sLocal.Month()) != expectedG.Month || sLocal.Year() != expectedG.Year {
		t.Errorf("scheduled_at = %s; expected lunar-derived Gregorian %d-%02d-%02d",
			sLocal.Format(time.RFC3339), expectedG.Year, expectedG.Month, expectedG.Day)
	}
}
