package reminder

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/usememos/memos/plugin/notification"
	"github.com/usememos/memos/store"
)

// Runner processes due reminders and sends notifications.
type Runner struct {
	store               *store.Store
	notificationService *notification.Service
}

// NewRunner creates a new reminder runner.
func NewRunner(store *store.Store, notificationService *notification.Service) *Runner {
	return &Runner{
		store:               store,
		notificationService: notificationService,
	}
}

// RunOnce processes all due reminders.
// This is idempotent - it only processes PENDING reminders where remind_at <= now.
func (r *Runner) RunOnce(ctx context.Context) error {
	now := time.Now().Unix()

	// Find all pending reminders that are due
	pendingStatus := store.ReminderStatusPending
	reminders, err := r.store.ListReminders(ctx, &store.FindReminder{
		Status:         &pendingStatus,
		RemindAtBefore: &now,
	})
	if err != nil {
		return fmt.Errorf("failed to list due reminders: %w", err)
	}

	if len(reminders) == 0 {
		return nil
	}

	slog.Info("processing due reminders", "count", len(reminders))

	processed := 0
	for _, reminder := range reminders {
		if err := r.processReminder(ctx, reminder); err != nil {
			slog.Error("failed to process reminder",
				"reminderID", reminder.ID,
				"error", err,
			)
			continue
		}
		processed++
	}

	slog.Info("finished processing reminders",
		"total", len(reminders),
		"processed", processed,
	)

	return nil
}

// processReminder handles a single reminder: sends notification and marks as triggered.
func (r *Runner) processReminder(ctx context.Context, reminder *store.Reminder) error {
	// Get the memo to include in the notification
	memo, err := r.store.GetMemo(ctx, &store.FindMemo{ID: &reminder.MemoID})
	if err != nil {
		return fmt.Errorf("failed to get memo: %w", err)
	}
	if memo == nil {
		// Memo was deleted, mark reminder as triggered anyway
		slog.Warn("memo not found for reminder, marking as triggered",
			"reminderID", reminder.ID,
			"memoID", reminder.MemoID,
		)
		return r.markAsTriggered(ctx, reminder)
	}

	// Get the user
	user, err := r.store.GetUser(ctx, &store.FindUser{ID: &reminder.CreatorID})
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		// User was deleted, mark reminder as triggered
		slog.Warn("user not found for reminder, marking as triggered",
			"reminderID", reminder.ID,
			"userID", reminder.CreatorID,
		)
		return r.markAsTriggered(ctx, reminder)
	}

	// Build notification
	notif := &notification.Notification{
		UserID: reminder.CreatorID,
		Title:  "Reminder",
		Body:   truncateContent(memo.Content, 100),
		URL:    fmt.Sprintf("/memos/%s", memo.UID),
		Data: map[string]any{
			"reminder_uid": reminder.UID,
			"memo_id":      memo.ID,
			"memo_uid":     memo.UID,
		},
	}

	// Send notification through all enabled providers
	sentCount := r.notificationService.Notify(ctx, notif)
	slog.Debug("sent reminder notification",
		"reminderID", reminder.ID,
		"userID", reminder.CreatorID,
		"providersNotified", sentCount,
	)

	// Mark as triggered
	return r.markAsTriggered(ctx, reminder)
}

// markAsTriggered updates the reminder status to TRIGGERED.
func (r *Runner) markAsTriggered(ctx context.Context, reminder *store.Reminder) error {
	triggeredStatus := store.ReminderStatusTriggered
	return r.store.UpdateReminder(ctx, &store.UpdateReminder{
		ID:     reminder.ID,
		Status: &triggeredStatus,
	})
}

// Run starts a continuous loop that processes reminders at the specified interval.
func (r *Runner) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run once immediately
	if err := r.RunOnce(ctx); err != nil {
		slog.Error("reminder runner error", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("reminder runner stopping")
			return
		case <-ticker.C:
			if err := r.RunOnce(ctx); err != nil {
				slog.Error("reminder runner error", "error", err)
			}
		}
	}
}

// truncateContent truncates content to maxLen characters, adding "..." if truncated.
func truncateContent(content string, maxLen int) string {
	runes := []rune(content)
	if len(runes) <= maxLen {
		return content
	}
	return string(runes[:maxLen-3]) + "..."
}
