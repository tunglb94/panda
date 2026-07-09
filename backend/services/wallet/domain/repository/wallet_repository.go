// Package repository defines the persistence interfaces for the Wallet domain.
package repository

import (
	"context"

	"github.com/fairride/wallet/domain/entity"
)

// WalletRepository persists and retrieves Wallet aggregates.
// Each owner has exactly one wallet per currency.
type WalletRepository interface {
	// Save upserts a wallet. created_at is immutable after first insert.
	Save(ctx context.Context, wallet *entity.Wallet) error

	// FindByID returns the wallet with the given ID.
	// Returns CodeNotFound if no wallet has that ID.
	FindByID(ctx context.Context, walletID string) (*entity.Wallet, error)

	// FindByOwnerID returns the wallet owned by the given owner.
	// Returns CodeNotFound if the owner has no wallet.
	FindByOwnerID(ctx context.Context, ownerID string) (*entity.Wallet, error)
}
