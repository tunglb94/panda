package app

import (
	"context"
	"time"

	"github.com/fairride/notification/domain/entity"
	"github.com/fairride/notification/domain/repository"
	"github.com/google/uuid"
)

// RecordCallUseCase writes a CallSession audit row and best-effort notifies
// the callee "cuộc gọi đến". It does NOT resolve or return a phone number —
// that lookup spans identity + driver profile data that this package has no
// reason to depend on; Gateway's call handler resolves the real number
// itself (see Part 1) and calls this use case purely for the audit
// side-effect, after authorizing the call.
type RecordCallUseCase struct {
	calls  repository.CallSessionRepository
	notify *CreateNotificationUseCase
}

func NewRecordCallUseCase(calls repository.CallSessionRepository, notify *CreateNotificationUseCase) *RecordCallUseCase {
	return &RecordCallUseCase{calls: calls, notify: notify}
}

func (uc *RecordCallUseCase) Execute(ctx context.Context, tripID, callerID, calleeID string) (*entity.CallSession, error) {
	cs, err := entity.NewCallSession(uuid.NewString(), tripID, callerID, calleeID, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if err := uc.calls.Save(ctx, cs); err != nil {
		return nil, err
	}
	if uc.notify != nil {
		_, _ = uc.notify.Execute(ctx, CreateNotificationInput{
			UserID:   calleeID,
			Category: entity.CategoryCall,
			Title:    "Cuộc gọi đến",
			Body:     "Bạn có một cuộc gọi liên quan đến chuyến đi.",
			TripID:   tripID,
		})
	}
	return cs, nil
}
