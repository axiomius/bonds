package services

import (
	"fmt"
	"time"

	lunarCalendar "github.com/6tail/lunar-go/calendar"
	calendarPkg "github.com/naiba/bonds/internal/calendar"
	"github.com/naiba/bonds/internal/models"
)

// formatReminderDate produces the human date string embedded in reminder
// notifications. Shape depends on calendar type and user preference:
//
//   - reminder is Gregorian, or user has alt-calendar OFF:
//     "2026-09-25" (the Gregorian fire date)
//   - reminder is Lunar AND user has alt-calendar ON:
//     "农历八月十五 (2026-09-25)" — the lunar date the user originally
//     entered, with the Gregorian equivalent in parens so the email is
//     still useful to anyone reading it who runs a Gregorian planner
//
// The Gregorian-in-parens convention matches what CalendarDatePicker shows
// in the UI when alt calendar is enabled, so user expectation is preserved
// between the form they filled in and the notification they get back.
func formatReminderDate(reminder *models.ContactReminder, fireAt time.Time, enableAltCalendar bool) string {
	gregStr := fireAt.Format("2006-01-02")

	ct := calendarPkg.CalendarType(reminder.CalendarType)
	if !enableAltCalendar || ct == "" || ct == calendarPkg.Gregorian {
		return gregStr
	}

	if ct == calendarPkg.Lunar {
		solar := lunarCalendar.NewSolarFromYmd(fireAt.Year(), int(fireAt.Month()), fireAt.Day())
		lunar := solar.GetLunar()
		return fmt.Sprintf("农历%s月%s (%s)",
			lunar.GetMonthInChinese(),
			lunar.GetDayInChinese(),
			gregStr,
		)
	}

	return gregStr
}
