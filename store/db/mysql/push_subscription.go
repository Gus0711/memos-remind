package mysql

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/usememos/memos/store"
)

func (d *DB) CreatePushSubscription(ctx context.Context, create *store.PushSubscription) (*store.PushSubscription, error) {
	fields := []string{"`user_id`", "`endpoint`", "`p256dh`", "`auth`"}
	placeholder := []string{"?", "?", "?", "?"}
	args := []any{create.UserID, create.Endpoint, create.P256dh, create.Auth}

	if create.UserAgent != nil {
		fields = append(fields, "`user_agent`")
		placeholder = append(placeholder, "?")
		args = append(args, *create.UserAgent)
	}

	stmt := "INSERT INTO `push_subscription` (" + strings.Join(fields, ", ") + ") VALUES (" + strings.Join(placeholder, ", ") + ")"
	result, err := d.db.ExecContext(ctx, stmt, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create push subscription")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get last insert id")
	}

	id32 := int32(id)
	sub, err := d.GetPushSubscription(ctx, &store.FindPushSubscription{ID: &id32})
	if err != nil {
		return nil, err
	}
	return sub, nil
}

func (d *DB) ListPushSubscriptions(ctx context.Context, find *store.FindPushSubscription) ([]*store.PushSubscription, error) {
	where, args := []string{"1 = 1"}, []any{}

	if find.ID != nil {
		where, args = append(where, "`id` = ?"), append(args, *find.ID)
	}
	if find.UserID != nil {
		where, args = append(where, "`user_id` = ?"), append(args, *find.UserID)
	}
	if find.Endpoint != nil {
		where, args = append(where, "`endpoint` = ?"), append(args, *find.Endpoint)
	}

	query := "SELECT `id`, `user_id`, `endpoint`, `p256dh`, `auth`, `user_agent`, `created_ts` FROM `push_subscription` WHERE " + strings.Join(where, " AND ") + " ORDER BY `created_ts` DESC"

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

	if delete.ID != nil {
		where, args = append(where, "`id` = ?"), append(args, *delete.ID)
	}
	if delete.UserID != nil {
		where, args = append(where, "`user_id` = ?"), append(args, *delete.UserID)
	}
	if delete.Endpoint != nil {
		where, args = append(where, "`endpoint` = ?"), append(args, *delete.Endpoint)
	}

	if len(where) == 0 {
		return nil
	}

	query := "DELETE FROM `push_subscription` WHERE " + strings.Join(where, " AND ")
	_, err := d.db.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "failed to delete push subscription")
	}
	return nil
}
