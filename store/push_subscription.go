package store

import (
	"context"
)

// PushSubscription represents a web push subscription for a user/device.
type PushSubscription struct {
	ID        int32
	UserID    int32
	Endpoint  string
	P256dh    string
	Auth      string
	UserAgent *string
	CreatedTs int64
}

// FindPushSubscription specifies filter criteria for querying push subscriptions.
type FindPushSubscription struct {
	ID       *int32
	UserID   *int32
	Endpoint *string
}

// DeletePushSubscription specifies which push subscription to delete.
type DeletePushSubscription struct {
	ID       *int32
	UserID   *int32
	Endpoint *string
}

// CreatePushSubscription creates a new push subscription.
func (s *Store) CreatePushSubscription(ctx context.Context, create *PushSubscription) (*PushSubscription, error) {
	return s.driver.CreatePushSubscription(ctx, create)
}

// ListPushSubscriptions retrieves push subscriptions matching the filter criteria.
func (s *Store) ListPushSubscriptions(ctx context.Context, find *FindPushSubscription) ([]*PushSubscription, error) {
	return s.driver.ListPushSubscriptions(ctx, find)
}

// GetPushSubscription retrieves a single push subscription matching the filter criteria.
func (s *Store) GetPushSubscription(ctx context.Context, find *FindPushSubscription) (*PushSubscription, error) {
	subscriptions, err := s.ListPushSubscriptions(ctx, find)
	if err != nil {
		return nil, err
	}
	if len(subscriptions) == 0 {
		return nil, nil
	}
	return subscriptions[0], nil
}

// DeletePushSubscription permanently removes a push subscription.
func (s *Store) DeletePushSubscription(ctx context.Context, delete *DeletePushSubscription) error {
	return s.driver.DeletePushSubscription(ctx, delete)
}
