package memory_test

import (
	"context"
	"testing"
	"time"

	domainerrors "github.com/fairride/shared/errors"
	"github.com/fairride/trip/domain/entity"
	"github.com/fairride/trip/infrastructure/memory"
)

var testNow = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func newTestDelivery(t *testing.T, id string) *entity.Delivery {
	t.Helper()
	d, err := entity.NewDelivery(
		id, "Nguyen Van A", "0912345678", "Tran Thi B", "0987654321",
		"gate 1", "leave at desk", entity.PackageTypeSmall, 2.0, false, 500000, testNow,
	)
	if err != nil {
		t.Fatalf("failed to build test delivery: %v", err)
	}
	return d
}

func TestDeliveryRepository_SaveAndFindByID(t *testing.T) {
	repo := memory.NewDeliveryRepository()
	ctx := context.Background()
	d := newTestDelivery(t, "d1")

	if err := repo.Save(ctx, d); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, err := repo.FindByID(ctx, "d1")
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if got.DeliveryID != "d1" {
		t.Errorf("DeliveryID = %q, want %q", got.DeliveryID, "d1")
	}
	if got.SenderName != d.SenderName {
		t.Errorf("SenderName = %q, want %q", got.SenderName, d.SenderName)
	}
}

func TestDeliveryRepository_FindByID_NotFound(t *testing.T) {
	repo := memory.NewDeliveryRepository()
	_, err := repo.FindByID(context.Background(), "does-not-exist")
	if !domainerrors.IsCode(err, domainerrors.CodeNotFound) {
		t.Errorf("expected NotFound, got %v", err)
	}
}

func TestDeliveryRepository_Save_NilDelivery(t *testing.T) {
	repo := memory.NewDeliveryRepository()
	err := repo.Save(context.Background(), nil)
	if !domainerrors.IsCode(err, domainerrors.CodeInvalidArgument) {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestDeliveryRepository_Save_UpsertOverwritesExisting(t *testing.T) {
	repo := memory.NewDeliveryRepository()
	ctx := context.Background()
	d := newTestDelivery(t, "d1")
	if err := repo.Save(ctx, d); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if err := d.AcceptByDriver(testNow); err != nil {
		t.Fatalf("AcceptByDriver failed: %v", err)
	}
	if err := repo.Save(ctx, d); err != nil {
		t.Fatalf("Save (update) failed: %v", err)
	}

	got, err := repo.FindByID(ctx, "d1")
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if got.Status != entity.DeliveryStatusAccepted {
		t.Errorf("Status = %q, want ACCEPTED after update", got.Status)
	}
}

func TestDeliveryRepository_Save_StoresDefensiveCopy(t *testing.T) {
	repo := memory.NewDeliveryRepository()
	ctx := context.Background()
	d := newTestDelivery(t, "d1")
	if err := repo.Save(ctx, d); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Mutating the caller's original after Save must not affect the stored copy.
	d.SenderName = "mutated after save"

	got, err := repo.FindByID(ctx, "d1")
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if got.SenderName == "mutated after save" {
		t.Error("repository did not store a defensive copy on Save")
	}
}

func TestDeliveryRepository_FindByID_ReturnsDefensiveCopy(t *testing.T) {
	repo := memory.NewDeliveryRepository()
	ctx := context.Background()
	d := newTestDelivery(t, "d1")
	if err := repo.Save(ctx, d); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, err := repo.FindByID(ctx, "d1")
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	got.SenderName = "mutated after find"

	got2, err := repo.FindByID(ctx, "d1")
	if err != nil {
		t.Fatalf("FindByID (second read) failed: %v", err)
	}
	if got2.SenderName == "mutated after find" {
		t.Error("repository did not return a defensive copy on FindByID")
	}
}
