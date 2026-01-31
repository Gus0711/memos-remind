package v1

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/usememos/memos/internal/util"
	v1pb "github.com/usememos/memos/proto/gen/api/v1"
	"github.com/usememos/memos/store"
)

func (s *APIV1Service) CreateReminder(ctx context.Context, request *v1pb.CreateReminderRequest) (*v1pb.Reminder, error) {
	user, err := s.fetchCurrentUser(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current user: %v", err)
	}
	if user == nil {
		return nil, status.Errorf(codes.Unauthenticated, "authentication required")
	}

	// Extract memo UID from the memo name
	memoUID, err := ExtractMemoUIDFromName(request.Memo)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid memo name: %v", err)
	}

	// Verify the memo exists
	memo, err := s.Store.GetMemo(ctx, &store.FindMemo{UID: &memoUID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get memo: %v", err)
	}
	if memo == nil {
		return nil, status.Errorf(codes.NotFound, "memo not found")
	}

	// Validate remind_at is in the future
	if request.RemindAt == nil {
		return nil, status.Errorf(codes.InvalidArgument, "remind_at is required")
	}
	remindAt := request.RemindAt.AsTime()
	if remindAt.Before(time.Now()) {
		return nil, status.Errorf(codes.InvalidArgument, "remind_at must be in the future")
	}

	// Create the reminder
	reminder, err := s.Store.CreateReminder(ctx, &store.Reminder{
		UID:       util.GenUUID(),
		MemoID:    memo.ID,
		CreatorID: user.ID,
		RemindAt:  remindAt.Unix(),
		Status:    store.ReminderStatusPending,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create reminder: %v", err)
	}

	return convertReminderFromStore(reminder, memoUID, user.ID), nil
}

func (s *APIV1Service) ListReminders(ctx context.Context, request *v1pb.ListRemindersRequest) (*v1pb.ListRemindersResponse, error) {
	user, err := s.fetchCurrentUser(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current user: %v", err)
	}
	if user == nil {
		return nil, status.Errorf(codes.Unauthenticated, "authentication required")
	}

	find := &store.FindReminder{
		CreatorID: &user.ID,
	}

	// Filter by memo if provided
	if request.Memo != "" {
		memoUID, err := ExtractMemoUIDFromName(request.Memo)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid memo name: %v", err)
		}
		memo, err := s.Store.GetMemo(ctx, &store.FindMemo{UID: &memoUID})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get memo: %v", err)
		}
		if memo != nil {
			find.MemoID = &memo.ID
		}
	}

	// Filter by status if provided
	if request.Status != v1pb.Reminder_STATUS_UNSPECIFIED {
		storeStatus := convertReminderStatusToStore(request.Status)
		find.Status = &storeStatus
	}

	// Pagination
	limit := int(request.PageSize)
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	find.Limit = &limit

	reminders, err := s.Store.ListReminders(ctx, find)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list reminders: %v", err)
	}

	// Build memo UID lookup map for efficiency
	memoUIDs := make(map[int32]string)
	for _, reminder := range reminders {
		if _, ok := memoUIDs[reminder.MemoID]; !ok {
			memo, err := s.Store.GetMemo(ctx, &store.FindMemo{ID: &reminder.MemoID})
			if err == nil && memo != nil {
				memoUIDs[reminder.MemoID] = memo.UID
			}
		}
	}

	// Convert to proto
	pbReminders := make([]*v1pb.Reminder, 0, len(reminders))
	for _, reminder := range reminders {
		memoUID := memoUIDs[reminder.MemoID]
		if memoUID == "" {
			continue // Skip reminders with deleted memos
		}
		pbReminders = append(pbReminders, convertReminderFromStore(reminder, memoUID, user.ID))
	}

	return &v1pb.ListRemindersResponse{
		Reminders: pbReminders,
	}, nil
}

func (s *APIV1Service) GetReminder(ctx context.Context, request *v1pb.GetReminderRequest) (*v1pb.Reminder, error) {
	user, err := s.fetchCurrentUser(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current user: %v", err)
	}
	if user == nil {
		return nil, status.Errorf(codes.Unauthenticated, "authentication required")
	}

	reminderUID, err := ExtractReminderUIDFromName(request.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid reminder name: %v", err)
	}

	reminder, err := s.Store.GetReminder(ctx, &store.FindReminder{UID: &reminderUID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get reminder: %v", err)
	}
	if reminder == nil {
		return nil, status.Errorf(codes.NotFound, "reminder not found")
	}

	// Check permission
	if reminder.CreatorID != user.ID {
		return nil, status.Errorf(codes.PermissionDenied, "permission denied")
	}

	// Get memo UID
	memo, err := s.Store.GetMemo(ctx, &store.FindMemo{ID: &reminder.MemoID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get memo: %v", err)
	}
	memoUID := ""
	if memo != nil {
		memoUID = memo.UID
	}

	return convertReminderFromStore(reminder, memoUID, user.ID), nil
}

func (s *APIV1Service) UpdateReminder(ctx context.Context, request *v1pb.UpdateReminderRequest) (*v1pb.Reminder, error) {
	user, err := s.fetchCurrentUser(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current user: %v", err)
	}
	if user == nil {
		return nil, status.Errorf(codes.Unauthenticated, "authentication required")
	}

	if request.Reminder == nil {
		return nil, status.Errorf(codes.InvalidArgument, "reminder is required")
	}
	if request.UpdateMask == nil || len(request.UpdateMask.Paths) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "update mask is required")
	}

	reminderUID, err := ExtractReminderUIDFromName(request.Reminder.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid reminder name: %v", err)
	}

	reminder, err := s.Store.GetReminder(ctx, &store.FindReminder{UID: &reminderUID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get reminder: %v", err)
	}
	if reminder == nil {
		return nil, status.Errorf(codes.NotFound, "reminder not found")
	}

	// Check permission
	if reminder.CreatorID != user.ID {
		return nil, status.Errorf(codes.PermissionDenied, "permission denied")
	}

	update := &store.UpdateReminder{
		ID: reminder.ID,
	}

	for _, path := range request.UpdateMask.Paths {
		switch path {
		case "remind_at":
			if request.Reminder.RemindAt == nil {
				return nil, status.Errorf(codes.InvalidArgument, "remind_at is required")
			}
			remindAt := request.Reminder.RemindAt.AsTime().Unix()
			if remindAt < time.Now().Unix() {
				return nil, status.Errorf(codes.InvalidArgument, "remind_at must be in the future")
			}
			update.RemindAt = &remindAt
		}
	}

	if err := s.Store.UpdateReminder(ctx, update); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update reminder: %v", err)
	}

	// Re-fetch the updated reminder
	reminder, err = s.Store.GetReminder(ctx, &store.FindReminder{ID: &reminder.ID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get updated reminder: %v", err)
	}

	// Get memo UID
	memo, err := s.Store.GetMemo(ctx, &store.FindMemo{ID: &reminder.MemoID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get memo: %v", err)
	}
	memoUID := ""
	if memo != nil {
		memoUID = memo.UID
	}

	return convertReminderFromStore(reminder, memoUID, user.ID), nil
}

func (s *APIV1Service) DeleteReminder(ctx context.Context, request *v1pb.DeleteReminderRequest) (*emptypb.Empty, error) {
	user, err := s.fetchCurrentUser(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current user: %v", err)
	}
	if user == nil {
		return nil, status.Errorf(codes.Unauthenticated, "authentication required")
	}

	reminderUID, err := ExtractReminderUIDFromName(request.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid reminder name: %v", err)
	}

	reminder, err := s.Store.GetReminder(ctx, &store.FindReminder{UID: &reminderUID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get reminder: %v", err)
	}
	if reminder == nil {
		return nil, status.Errorf(codes.NotFound, "reminder not found")
	}

	// Check permission
	if reminder.CreatorID != user.ID {
		return nil, status.Errorf(codes.PermissionDenied, "permission denied")
	}

	if err := s.Store.DeleteReminder(ctx, &store.DeleteReminder{ID: reminder.ID}); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete reminder: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *APIV1Service) DismissReminder(ctx context.Context, request *v1pb.DismissReminderRequest) (*v1pb.Reminder, error) {
	user, err := s.fetchCurrentUser(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current user: %v", err)
	}
	if user == nil {
		return nil, status.Errorf(codes.Unauthenticated, "authentication required")
	}

	reminderUID, err := ExtractReminderUIDFromName(request.Name)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid reminder name: %v", err)
	}

	reminder, err := s.Store.GetReminder(ctx, &store.FindReminder{UID: &reminderUID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get reminder: %v", err)
	}
	if reminder == nil {
		return nil, status.Errorf(codes.NotFound, "reminder not found")
	}

	// Check permission
	if reminder.CreatorID != user.ID {
		return nil, status.Errorf(codes.PermissionDenied, "permission denied")
	}

	// Update status to dismissed
	dismissedStatus := store.ReminderStatusDismissed
	if err := s.Store.UpdateReminder(ctx, &store.UpdateReminder{
		ID:     reminder.ID,
		Status: &dismissedStatus,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to dismiss reminder: %v", err)
	}

	// Re-fetch the updated reminder
	reminder, err = s.Store.GetReminder(ctx, &store.FindReminder{ID: &reminder.ID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get updated reminder: %v", err)
	}

	// Get memo UID
	memo, err := s.Store.GetMemo(ctx, &store.FindMemo{ID: &reminder.MemoID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get memo: %v", err)
	}
	memoUID := ""
	if memo != nil {
		memoUID = memo.UID
	}

	return convertReminderFromStore(reminder, memoUID, user.ID), nil
}

// Helper functions

func convertReminderFromStore(reminder *store.Reminder, memoUID string, creatorID int32) *v1pb.Reminder {
	return &v1pb.Reminder{
		Name:       fmt.Sprintf("%s%s", ReminderNamePrefix, reminder.UID),
		Memo:       fmt.Sprintf("%s%s", MemoNamePrefix, memoUID),
		Creator:    fmt.Sprintf("%s%d", UserNamePrefix, creatorID),
		RemindAt:   timestamppb.New(time.Unix(reminder.RemindAt, 0)),
		Status:     convertReminderStatusFromStore(reminder.Status),
		CreateTime: timestamppb.New(time.Unix(reminder.CreatedTs, 0)),
		UpdateTime: timestamppb.New(time.Unix(reminder.UpdatedTs, 0)),
	}
}

func convertReminderStatusFromStore(status store.ReminderStatus) v1pb.Reminder_Status {
	switch status {
	case store.ReminderStatusPending:
		return v1pb.Reminder_PENDING
	case store.ReminderStatusTriggered:
		return v1pb.Reminder_TRIGGERED
	case store.ReminderStatusDismissed:
		return v1pb.Reminder_DISMISSED
	default:
		return v1pb.Reminder_STATUS_UNSPECIFIED
	}
}

func convertReminderStatusToStore(status v1pb.Reminder_Status) store.ReminderStatus {
	switch status {
	case v1pb.Reminder_PENDING:
		return store.ReminderStatusPending
	case v1pb.Reminder_TRIGGERED:
		return store.ReminderStatusTriggered
	case v1pb.Reminder_DISMISSED:
		return store.ReminderStatusDismissed
	default:
		return store.ReminderStatusPending
	}
}
