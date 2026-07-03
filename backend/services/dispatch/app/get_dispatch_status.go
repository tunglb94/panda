package app

import (
	"context"

	"github.com/fairride/dispatch/domain/entity"
	"github.com/fairride/dispatch/domain/repository"
	domainerrors "github.com/fairride/shared/errors"
)

// GetDispatchStatusUseCase retrieves the current dispatch job for a trip.
type GetDispatchStatusUseCase struct {
	jobRepo repository.DispatchJobRepository
}

func NewGetDispatchStatusUseCase(jobRepo repository.DispatchJobRepository) *GetDispatchStatusUseCase {
	return &GetDispatchStatusUseCase{jobRepo: jobRepo}
}

func (uc *GetDispatchStatusUseCase) Execute(ctx context.Context, tripID string) (*entity.DispatchJob, error) {
	if tripID == "" {
		return nil, domainerrors.InvalidArgument("trip_id is required")
	}
	return uc.jobRepo.FindByTripID(ctx, tripID)
}
