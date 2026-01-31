package notification

import (
	"context"

	storepb "github.com/usememos/memos/proto/gen/store"
	"github.com/usememos/memos/store"
)

const InboxProviderName = "inbox"

// InboxProvider sends notifications via the in-app inbox system.
type InboxProvider struct {
	store *store.Store
}

// NewInboxProvider creates a new inbox notification provider.
func NewInboxProvider(store *store.Store) *InboxProvider {
	return &InboxProvider{
		store: store,
	}
}

// Name returns the provider name.
func (p *InboxProvider) Name() string {
	return InboxProviderName
}

// IsEnabled returns true since inbox notifications are always enabled.
func (p *InboxProvider) IsEnabled(_ context.Context, _ int32) bool {
	// Inbox notifications are always enabled for all users
	return true
}

// Send creates an inbox notification for the user.
func (p *InboxProvider) Send(ctx context.Context, notification *Notification) error {
	// Extract reminder UID from notification data
	var reminderUID *string
	if notification.Data != nil {
		if uid, ok := notification.Data["reminder_uid"].(string); ok {
			reminderUID = &uid
		}
	}

	// The sender is the system (user ID 0 or the same as receiver for self-reminders)
	senderID := notification.UserID

	inbox := &store.Inbox{
		SenderID:   senderID,
		ReceiverID: notification.UserID,
		Status:     store.UNREAD,
		Message: &storepb.InboxMessage{
			Type:        storepb.InboxMessage_REMINDER,
			ReminderUid: reminderUID,
		},
	}

	_, err := p.store.CreateInbox(ctx, inbox)
	return err
}
