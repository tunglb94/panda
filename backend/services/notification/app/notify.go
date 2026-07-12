package app

import (
	"context"
	"time"

	"github.com/fairride/notification/domain/entity"
	"github.com/fairride/notification/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
	"github.com/google/uuid"
)

// CreateNotificationUseCase writes one Notification to a user's in-app feed
// and best-effort forwards it to PushSender. Used both as a direct side
// effect (chat's new-message notification, call's incoming-call
// notification) and as the target of Gateway's additive trip-lifecycle hook
// (Part 3 — accept/arrive/start/finish/cancel/pickup/deliver events).
type CreateNotificationUseCase struct {
	notifications repository.NotificationRepository
	push          PushSender
}

func NewCreateNotificationUseCase(notifications repository.NotificationRepository, push PushSender) *CreateNotificationUseCase {
	return &CreateNotificationUseCase{notifications: notifications, push: push}
}

type CreateNotificationInput struct {
	UserID         string
	Category       entity.Category
	Title          string
	Body           string
	TripID         string
	ConversationID string
}

func (uc *CreateNotificationUseCase) Execute(ctx context.Context, in CreateNotificationInput) (*entity.Notification, error) {
	n, err := entity.NewNotification(uuid.NewString(), in.UserID, in.Category, in.Title, in.Body, in.TripID, in.ConversationID, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if err := uc.notifications.Save(ctx, n); err != nil {
		return nil, err
	}
	// Best-effort: a push-delivery failure must never fail the notification
	// write itself — the in-app feed is the source of truth.
	if uc.push != nil {
		_ = uc.push.Send(ctx, in.UserID, in.Title, in.Body)
	}
	return n, nil
}

// ListNotificationsUseCase returns a user's own notification feed.
type ListNotificationsUseCase struct {
	notifications repository.NotificationRepository
}

func NewListNotificationsUseCase(notifications repository.NotificationRepository) *ListNotificationsUseCase {
	return &ListNotificationsUseCase{notifications: notifications}
}

type NotificationList struct {
	Items       []*entity.Notification
	UnreadCount int
}

func (uc *ListNotificationsUseCase) Execute(ctx context.Context, userID string, limit int) (NotificationList, error) {
	if userID == "" {
		return NotificationList{}, domainerrors.InvalidArgument("user_id is required")
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	items, err := uc.notifications.ListByUser(ctx, userID, limit)
	if err != nil {
		return NotificationList{}, err
	}
	unread, err := uc.notifications.CountUnread(ctx, userID)
	if err != nil {
		return NotificationList{}, err
	}
	return NotificationList{Items: items, UnreadCount: unread}, nil
}

// MarkNotificationReadUseCase marks a single notification read. Scoped to
// the requester's own userID — nothing prevents a user from marking a
// notification that isn't theirs as "read" other than it simply not
// existing for that (id, user_id) pair (repository.MarkRead is a no-op if
// the row doesn't match both).
type MarkNotificationReadUseCase struct {
	notifications repository.NotificationRepository
}

func NewMarkNotificationReadUseCase(notifications repository.NotificationRepository) *MarkNotificationReadUseCase {
	return &MarkNotificationReadUseCase{notifications: notifications}
}

func (uc *MarkNotificationReadUseCase) Execute(ctx context.Context, id, userID string) error {
	if id == "" || userID == "" {
		return domainerrors.InvalidArgument("id and user_id are required")
	}
	return uc.notifications.MarkRead(ctx, id, userID, time.Now().UTC())
}
