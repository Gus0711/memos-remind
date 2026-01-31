package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/usememos/memos/store"
)

func (d *DB) CreatePushSubscription(ctx context.Context, create *store.PushSubscription) (*store.PushSubscription, error) {
	fields := []string{"user_id", "endpoint", "p256dh", "auth"}
	args := []any{create.UserID, create.Endpoint, create.P256dh, create.Auth}

	if create.UserAgent != nil {
		fields = append(fields, "user_agent")
		args = append(args, *create.UserAgent)
	}

	stmt := "INSERT INTO push_subscription (" + strings.Join(fields, ", ") + ") VALUES (" + placeholders(len(args)) + ") RETURNING id, created_ts"
	if err := d.db.QueryRowContext(ctx, stmt, args...).Scan(
		&create.ID,
		&create.CreatedTs,
	); err != nil {
		return nil, errors.Wrap(err, "failed to create push subscription")
	}

	return create, nil
}

func (d *DB) ListPushSubscriptions(ctx context.Context, find *store.FindPushSubscription) ([]*store.PushSubscription, error) {
	where, args := []string{"1 = 1"}, []any{}

	if find.ID != nil {
		where, args = append(where, "id = "+placeholder(len(args)+1)), append(args, *find.ID)
	}
	if find.UserID != nil {
		where, args = append(where, "user_id = "+placeholder(len(args)+1)), append(args, *find.UserID)
	}
	if find.Endpoint != nil {
		where, args = append(where, "endpoint = "+placeholder(len(args)+1)), append(args, *find.Endpoint)
	}

	query := "SELECT id, user_id, endpoint, p256dh, auth, user_agent, created_ts FROM push_subscription WHERE " + strings.Join(where, " AND ") + " ORDER BY created_ts DESC"

	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list push subscriptions")
	}
	defer rows.Close()

	list := []*store.PushSubscription{}
	for rows.Next() {
		sub := &store.PushSubscription{}
		if err := rows.Scan(
			&sub.ID,
			&sub.UserID,
			&sub.Endpoint,
			&sub.P256dh,
			&sub.Auth,
			&sub.UserAgent,
			&sub.CreatedTs,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan push subscription")
		}
		list = append(list, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate push subscriptions")
	}

	return list, nil
}

func (d *DB) GetPushSubscription(ctx context.Context, find *store.FindPushSubscription) (*store.PushSubscription, error) {
	list, err := d.ListPushSubscriptions(ctx, find)
	if err != nil {
		return nil, err
	}
	if len(list) != 1 {
		return nil, errors.Errorf("unexpected push subscription count: %d", len(list))
	}
	return list[0], nil
}

func (d *DB) DeletePushSubscription(ctx context.Context, delete *store.DeletePushSubscription) error {
	where, args := []string{}, []any{}
	argIndex := 1

	if delete.ID != nil {
		where, args = append(where, fmt.Sprintf("id = $%d", argIndex)), append(args, *delete.ID)
		argIndex++
	}
	if delete.UserID != nil {
		where, args = append(where, fmt.Sprintf("user_id = $%d", argIndex)), append(args, *delete.UserID)
		argIndex++
	}
	if delete.Endpoint != nil {
		where, args = append(where, fmt.Sprintf("endpoint = $%d", argIndex)), append(args, *delete.Endpoint)
		argIndex++
	}

	if len(where) == 0 {
		return nil
	}

	query := "DELETE FROM push_subscription WHERE " + strings.Join(where, " AND ")
	_, err := d.db.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "failed to delete push subscription")
	}
	return nil
}
