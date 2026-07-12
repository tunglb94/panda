package app

import (
	"context"
	"time"

	"github.com/fairride/notification/domain/entity"
	"github.com/fairride/notification/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
	"github.com/google/uuid"
)

// GetOrCreateConversationUseCase returns the conversation for a trip,
// creating it on first access. This is the single authorization checkpoint
// every chat/notification-badge read goes through (Part 7 — security: only
// the trip's own rider/driver ever reach a Conversation).
type GetOrCreateConversationUseCase struct {
	conversations repository.ConversationRepository
	trips         TripReader
}

func NewGetOrCreateConversationUseCase(conversations repository.ConversationRepository, trips TripReader) *GetOrCreateConversationUseCase {
	return &GetOrCreateConversationUseCase{conversations: conversations, trips: trips}
}

// Execute returns the conversation for tripID. requesterID must be the
// trip's rider or driver, or this returns PermissionDenied. If the trip's
// driver hasn't been assigned yet, returns PreconditionFailed (there is no
// second participant to chat with yet).
func (uc *GetOrCreateConversationUseCase) Execute(ctx context.Context, tripID, requesterID string) (*entity.Conversation, error) {
	if tripID == "" || requesterID == "" {
		return nil, domainerrors.InvalidArgument("trip_id and requester_id are required")
	}
	trip, err := uc.trips.GetTrip(ctx, tripID)
	if err != nil {
		return nil, err
	}
	if requesterID != trip.RiderID && requesterID != trip.DriverID {
		return nil, domainerrors.PermissionDenied("requester is not a participant of this trip")
	}
	if trip.DriverID == "" {
		return nil, domainerrors.PreconditionFailed("trip has no driver assigned yet")
	}

	conv, err := uc.conversations.FindByTripID(ctx, tripID)
	if err != nil {
		if domainerrors.GetCode(err) != domainerrors.CodeNotFound {
			return nil, err
		}
		conv = entity.NewConversation(uuid.NewString(), tripID, trip.RiderID, trip.DriverID, trip.TripType, time.Now().UTC())
		if saveErr := uc.conversations.Save(ctx, conv); saveErr != nil {
			if domainerrors.GetCode(saveErr) == domainerrors.CodeAlreadyExists {
				// Lost a create race — someone else already inserted the
				// row for this trip_id between our FindByTripID and Save.
				return uc.conversations.FindByTripID(ctx, tripID)
			}
			return nil, saveErr
		}
		return conv, nil
	}

	if IsTripStatusClosed(trip.Status) && conv.IsOpen() {
		conv.Close(time.Now().UTC())
		if updErr := uc.conversations.Update(ctx, conv); updErr != nil {
			return nil, updErr
		}
	}
	return conv, nil
}
