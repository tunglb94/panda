package app

import (
	"context"
	"time"

	"github.com/fairride/shared/errors"
	"github.com/fairride/user/domain/entity"
	"github.com/fairride/user/domain/repository"
)

// UpdateProfileInput carries the mutable fields accepted by UpdateProfileUseCase.
// Phone and Status are not included — they are owned by other services/flows.
type UpdateProfileInput struct {
	UserID      string
	FullName    string
	Email       string
	Avatar      string
	DateOfBirth time.Time // zero = clear the field
	Gender      entity.Gender
}

// UpdateProfileUseCase fetches, validates, mutates, and persists a user profile.
type UpdateProfileUseCase struct {
	repo repository.ProfileRepository
}

// NewUpdateProfileUseCase constructs an UpdateProfileUseCase.
func NewUpdateProfileUseCase(repo repository.ProfileRepository) *UpdateProfileUseCase {
	if repo == nil {
		panic("user: ProfileRepository must not be nil")
	}
	return &UpdateProfileUseCase{repo: repo}
}

// Execute applies the input to the existing profile and persists it.
// Returns CodeInvalidArgument for an empty UserID or invalid field values.
// Returns CodeNotFound when no profile exists for the given UserID.
func (uc *UpdateProfileUseCase) Execute(ctx context.Context, input UpdateProfileInput) (*entity.UserProfile, error) {
	if input.UserID == "" {
		return nil, errors.InvalidArgument("user id must not be empty")
	}

	profile, err := uc.repo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	if err := profile.Update(
		input.FullName,
		input.Email,
		input.Avatar,
		input.DateOfBirth,
		input.Gender,
		time.Now().UTC(),
	); err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}
