package store

import (
	"context"
)

// ReminderStatus represents the status of a reminder.
type ReminderStatus string

const (
	// ReminderStatusPending indicates the reminder has not been triggered yet.
	ReminderStatusPending ReminderStatus = "PENDING"
	// ReminderStatusTriggered indicates the reminder has been triggered.
	ReminderStatusTriggered ReminderStatus = "TRIGGERED"
	// ReminderStatusDismissed indicates the reminder has been dismissed by the user.
	ReminderStatusDismissed ReminderStatus = "DISMISSED"
)

func (s ReminderStatus) String() string {
	return string(s)
}

// Reminder represents a reminder for a memo.
type Reminder struct {
	ID        int32
	UID       string
	MemoID    int32
	CreatorID int32
	RemindAt  int64 // Unix timestamp when the reminder should trigger
	Status    ReminderStatus
	CreatedTs int64
	UpdatedTs int64
}

// FindReminder specifies filter criteria for querying reminders.
type FindReminder struct {
	ID             *int32
	UID            *string
	MemoID         *int32
	CreatorID      *int32
	Status         *ReminderStatus
	RemindAtBefore *int64 // Find reminders due before this time

	// Pagination
	Limit  *int
	Offset *int
}

// UpdateReminder contains fields that can be updated for a reminder.
type UpdateReminder struct {
	ID        int32
	RemindAt  *int64
	Status    *ReminderStatus
	UpdatedTs *int64
}

// DeleteReminder specifies which reminder to delete.
type DeleteReminder struct {
	ID int32
}

// CreateReminder creates a new reminder.
func (s *Store) CreateReminder(ctx context.Context, create *Reminder) (*Reminder, error) {
	return s.driver.CreateReminder(ctx, create)
}

// ListReminders retrieves reminders matching the filter criteria.
func (s *Store) ListReminders(ctx context.Context, find *FindReminder) ([]*Reminder, error) {
	return s.driver.ListReminders(ctx, find)
}

// GetReminder retrieves a single reminder matching the filter criteria.
func (s *Store) GetReminder(ctx context.Context, find *FindReminder) (*Reminder, error) {
	reminders, err := s.ListReminders(ctx, find)
	if err != nil {
		return nil, err
	}
	if len(reminders) == 0 {
		return nil, nil
	}
	return reminders[0], nil
}

// UpdateReminder updates an existing reminder.
func (s *Store) UpdateReminder(ctx context.Context, update *UpdateReminder) error {
	return s.driver.UpdateReminder(ctx, update)
}

// DeleteReminder permanently removes a reminder.
func (s *Store) DeleteReminder(ctx context.Context, delete *DeleteReminder) error {
	return s.driver.DeleteReminder(ctx, delete)
}
