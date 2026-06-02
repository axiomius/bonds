# Reminders

Bonds has a built-in reminder system that notifies you about important dates and events via email or Shoutrrr-backed channels such as Telegram.

## Types

| Type | Behavior |
|------|----------|
| **One-time** | Triggers once at the scheduled date/time, then done |
| **Weekly** | Repeats every week |
| **Monthly** | Repeats every month |
| **Yearly** | Repeats every year (ideal for birthdays) |

## How It Works

1. **Create a reminder** on a contact — choose the date, time, type, and a label
2. **The cron scheduler** runs every minute, scanning for due reminders
3. **Notifications are sent** through your configured notification channels at each channel's preferred send time
4. **For recurring reminders**, the next occurrence is automatically scheduled based on the previous scheduled time (not current time, to prevent drift)

## Notification Channels

Bonds supports email plus Shoutrrr-compatible notification channels:

### Email

Configure SMTP settings in the admin panel. Email notifications are enabled by default when a user registers — a notification channel is automatically created with the user's email address.

### Shoutrrr / Telegram

Add a Shoutrrr URL in user notification settings, such as a Telegram URL (`telegram://token@telegram?channels=123456`). Shoutrrr channels are active immediately after creation and can use any supported Shoutrrr service.

Each channel has a **preferred send time**. New reminders, existing-reminder backfills, and recurring reminder reschedules use that local time. Empty or invalid values fall back to `09:00`.

See [Shoutrrr / Telegram Notifications](/features/more#telegram-notifications) for setup details.

## Channel Reliability

Each notification channel tracks a failure counter. If a channel fails **10 consecutive times**, it is automatically disabled to prevent spam. You can re-enable it manually from user settings after fixing the underlying issue.

## Notification History

Every notification attempt is recorded in `UserNotificationSent`, including:
- Delivery status (success/failure)
- Error message (if failed)
- Timestamp
