package repository

import (
	"context"

	"github.com/fairride/identity/domain/entity"
)

// LoginHistoryRepository defines persistence operations for LoginRecord
// entities. Append-only — there is deliberately no Update/Delete.
type LoginHistoryRepository interface {
	Append(ctx context.Context, record *entity.LoginRecord) error
}
