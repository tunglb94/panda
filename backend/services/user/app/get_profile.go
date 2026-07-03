// Package app is the application layer of the User service.
// Use cases orchestrate domain entities and repository interfaces.
package app

import (
	"context"

	"github.com/fairride/shared/errors"
	"github.com/fairride/user/domain/entity"
	"github.com/fairride/user/domain/repository"
)

// GetProfileUseCase fetches a user's profile by ID.
type GetProfileUseCase struct {
	repo repository.ProfileRepository
}

// NewGetProfileUseCase constructs a GetProfileUseCase.
func NewGetProfileUseCase(repo repository.ProfileRepository) *GetProfileUseCase {
	if repo == nil {
		panic("user: ProfileRepository must not be nil")
	}
	return &GetProfileUseCase{repo: repo}
}

// Execute returns the profile for userID.
// Returns CodeInvalidArgument for an empty userID.
// Returns CodeNotFound when no profile exists.
func (uc *GetProfileUseCase) Execute(ctx context.Context, userID string) (*entity.UserProfile, error) {
	if userID == "" {
		return nil, errors.InvalidArgument("user id must not be empty")
	}
	return uc.repo.FindByID(ctx, userID)
}
