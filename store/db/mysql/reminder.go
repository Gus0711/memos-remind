package mysql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/usememos/memos/store"
)

func (d *DB) CreateReminder(ctx context.Context, create *store.Reminder) (*store.Reminder, error) {
	fields := []string{"`uid`", "`memo_id`", "`creator_id`", "`remind_at`", "`status`"}
	placeholder := []string{"?", "?", "?", "?", "?"}
	args := []any{create.UID, create.MemoID, create.CreatorID, create.RemindAt, create.Status}

	stmt := "INSERT INTO `reminder` (" + strings.Join(fields, ", ") + ") VALUES (" + strings.Join(placeholder, ", ") + ")"
	result, err := d.db.ExecContext(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	id32 := int32(id)
	reminder, err := d.GetReminder(ctx, &store.FindReminder{ID: &id32})
	if err != nil {
		return nil, err
	}
	return reminder, nil
}

func (d *DB) ListReminders(ctx context.Context, find *store.FindReminder) ([]*store.Reminder, error) {
	where, args := []string{"1 = 1"}, []any{}

	if find.ID != nil {
		where, args = append(where, "`id` = ?"), append(args, *find.ID)
	}
	if find.UID != nil {
		where, args = append(where, "`uid` = ?"), append(args, *find.UID)
	}
	if find.MemoID != nil {
		where, args = append(where, "`memo_id` = ?"), append(args, *find.MemoID)
	}
	if find.CreatorID != nil {
		where, args = append(where, "`creator_id` = ?"), append(args, *find.CreatorID)
	}
	if find.Status != nil {
		where, args = append(where, "`status` = ?"), append(args, *find.Status)
	}
	if find.RemindAtBefore != nil {
		where, args = append(where, "`remind_at` <= ?"), append(args, *find.RemindAtBefore)
	}

	query := "SELECT `id`, `uid`, `memo_id`, `creator_id`, `remind_at`, `status`, `created_ts`, `updated_ts` FROM `reminder` WHERE " + strings.Join(where, " AND ") + " ORDER BY `remind_at` ASC"
	if find.Limit != nil {
		query = fmt.Sprintf("%s LIMIT %d", query, *find.Limit)
		if find.Offset != nil {
			query = fmt.Sprintf("%s OFFSET %d", query, *find.Offset)
		}
	}

	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*store.Reminder{}
	for rows.Next() {
		reminder := &store.Reminder{}
		if err := rows.Scan(
			&reminder.ID,
			&reminder.UID,
			&reminder.MemoID,
			&reminder.CreatorID,
			&reminder.RemindAt,
			&reminder.Status,
			&reminder.CreatedTs,
			&reminder.UpdatedTs,
		); err != nil {
			return nil, err
		}
		list = append(list, reminder)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (d *DB) GetReminder(ctx context.Context, find *store.FindReminder) (*store.Reminder, error) {
	list, err := d.ListReminders(ctx, find)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get reminder")
	}
	if len(list) != 1 {
		return nil, errors.Errorf("unexpected reminder count: %d", len(list))
	}
	return list[0], nil
}

func (d *DB) UpdateReminder(ctx context.Context, update *store.UpdateReminder) error {
	set, args := []string{}, []any{}

	if update.RemindAt != nil {
		set, args = append(set, "`remind_at` = ?"), append(args, *update.RemindAt)
	}
	if update.Status != nil {
		set, args = append(set, "`status` = ?"), append(args, update.Status.String())
	}

	// Always update updated_ts
	set, args = append(set, "`updated_ts` = ?"), append(args, time.Now().Unix())

	if len(set) == 1 {
		// Only updated_ts, nothing else to update
		return nil
	}

	args = append(args, update.ID)
	query := "UPDATE `reminder` SET " + strings.Join(set, ", ") + " WHERE `id` = ?"

	_, err := d.db.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "failed to update reminder")
	}
	return nil
}

func (d *DB) DeleteReminder(ctx context.Context, delete *store.DeleteReminder) error {
	result, err := d.db.ExecContext(ctx, "DELETE FROM `reminder` WHERE `id` = ?", delete.ID)
	if err != nil {
		return errors.Wrap(err, "failed to delete reminder")
	}
	if _, err := result.RowsAffected(); err != nil {
		return err
	}
	return nil
}
