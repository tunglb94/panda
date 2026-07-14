package app

import (
	"context"
	"time"

	"github.com/fairride/identity/domain/entity"
	"github.com/fairride/identity/domain/repository"
)

// RecordLoginInput carries one login attempt to be appended to history.
// UserID may be empty (attempt failed before an account was resolved).
type RecordLoginInput struct {
	UserID      string
	IP          string
	DeviceID    string
	Platform    string
	LoginMethod entity.LoginMethod
	Success     bool
}

// RecordLoginUseCase appends one row to the login-history audit trail.
type RecordLoginUseCase struct {
	repo repository.LoginHistoryRepository
}

func NewRecordLoginUseCase(repo repository.LoginHistoryRepository) *RecordLoginUseCase {
	return &RecordLoginUseCase{repo: repo}
}

func (uc *RecordLoginUseCase) Execute(ctx context.Context, in RecordLoginInput) error {
	id, err := newID()
	if err != nil {
		return err
	}
	rec, err := entity.NewLoginRecord(id, in.UserID, time.Now(), in.IP, in.DeviceID, in.Platform, in.LoginMethod, in.Success)
	if err != nil {
		return err
	}
	return uc.repo.Append(ctx, rec)
}
