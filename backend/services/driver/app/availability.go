package app

import (
	"context"
	"time"

	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/domain/repository"
	"github.com/fairride/shared/errors"
)

// ─── GoOnlineUseCase ─────────────────────────────────────────────────────────

// GoOnlineUseCase marks a driver as online in Redis — gated by the Online
// Guard (Phần 9 of the Driver KYC Hardening spec): a driver may not go
// online unless DriverVerification is Approved, VehicleVerification is
// Approved, their license class still permits their registered ServiceType
// (via the Phần 1 Rule Engine), none of the vehicle's expiry-tracked
// documents (GPLX/registration/insurance/inspection — Phần 2) have passed
// their expiry date, and the vehicle carries at least one ServicePermission
// (Phần 8). Every failure returns a specific, actionable Vietnamese message.
// This is the ONLY enforcement point — no change was made to Dispatch's
// matching/ranking algorithm; a driver who can never successfully go online
// never enters Dispatch's online-driver pool in the first place.
type GoOnlineUseCase struct {
	repo                 repository.AvailabilityRepository
	driverVerifications  repository.DriverVerificationRepository
	vehicleVerifications repository.VehicleVerificationRepository
	documents            repository.KYCDocumentRepository
	rules                repository.LicenseCapabilityRepository
	audit                repository.AuditLogRepository
}

func NewGoOnlineUseCase(
	repo repository.AvailabilityRepository,
	driverVerifications repository.DriverVerificationRepository,
	vehicleVerifications repository.VehicleVerificationRepository,
	documents repository.KYCDocumentRepository,
	rules repository.LicenseCapabilityRepository,
	audit repository.AuditLogRepository,
) *GoOnlineUseCase {
	return &GoOnlineUseCase{
		repo: repo, driverVerifications: driverVerifications, vehicleVerifications: vehicleVerifications,
		documents: documents, rules: rules, audit: audit,
	}
}

func (uc *GoOnlineUseCase) Execute(ctx context.Context, driverID string) (*entity.AvailabilityState, error) {
	if driverID == "" {
		return nil, errors.InvalidArgument("driver id must not be empty")
	}
	if err := uc.checkEligible(ctx, driverID); err != nil {
		return nil, err
	}
	if err := uc.repo.SetOnline(ctx, driverID, time.Now()); err != nil {
		return nil, err
	}
	return uc.repo.GetAvailability(ctx, driverID)
}

// expiringVehicleDocTypes are checked, in order, against the vehicle's
// current capabilities — License only matters when RideEnabled (Phần 2:
// Delivery never depends on GPLX).
var expiringVehicleDocTypesAlways = []entity.DocumentType{
	entity.DocumentVehicleRegistration,
	entity.DocumentVehicleInsurance,
	entity.DocumentVehicleInspection,
}

// vietnameseExpiredMessage names which document expired, in Vietnamese —
// Phần 9's own examples ("GPLX đã hết hạn").
func vietnameseExpiredMessage(dt entity.DocumentType) string {
	switch dt {
	case entity.DocumentLicense:
		return "GPLX đã hết hạn"
	case entity.DocumentVehicleRegistration:
		return "Đăng ký xe đã hết hạn"
	case entity.DocumentVehicleInsurance:
		return "Bảo hiểm xe đã hết hạn"
	case entity.DocumentVehicleInspection:
		return "Đăng kiểm xe đã hết hạn"
	default:
		return "Một trong các giấy tờ của xe đã hết hạn"
	}
}

// checkEligible is the Online Guard. Returns CodePreconditionFailed (never
// silently no-ops) the moment any requirement isn't met, so the gateway can
// surface a specific, actionable reason to the driver.
func (uc *GoOnlineUseCase) checkEligible(ctx context.Context, driverID string) error {
	dv, err := uc.driverVerifications.FindByDriverID(ctx, driverID)
	if err != nil {
		if errors.GetCode(err) == errors.CodeNotFound {
			return errors.PreconditionFailed("Bạn chưa gửi hồ sơ xác minh cá nhân")
		}
		return err
	}
	if msg, ok := driverVerificationStatusMessage(dv.Status); !ok {
		return errors.PreconditionFailed(msg)
	}

	vv, err := uc.vehicleVerifications.FindByDriverID(ctx, driverID)
	if err != nil {
		if errors.GetCode(err) == errors.CodeNotFound {
			return errors.PreconditionFailed("Bạn chưa đăng ký xe")
		}
		return err
	}
	if msg, ok := vehicleVerificationStatusMessage(vv.Status); !ok {
		return errors.PreconditionFailed(msg)
	}

	if vv.RideEnabled {
		allowed, err := uc.rules.IsAllowed(ctx, vv.LicenseClass, vv.ServiceType)
		if err != nil {
			return err
		}
		if !allowed {
			return errors.PreconditionFailed("Hạng bằng lái không phù hợp với loại xe đã đăng ký")
		}
	}

	if !vv.HasAnyPermission() {
		return errors.PreconditionFailed("Xe chưa được cấp quyền hoạt động")
	}

	return uc.checkDocumentExpiry(ctx, vv)
}

// checkDocumentExpiry is Phần 2's "mỗi lần Go Online check" — no scheduler.
// The first expired document found auto-transitions VehicleVerification to
// Expired (persisted + audited, actor "system") and blocks Online with a
// specific Vietnamese message.
func (uc *GoOnlineUseCase) checkDocumentExpiry(ctx context.Context, vv *entity.VehicleVerification) error {
	docTypes := expiringVehicleDocTypesAlways
	if vv.RideEnabled {
		docTypes = append([]entity.DocumentType{entity.DocumentLicense}, docTypes...)
	}
	now := time.Now().UTC()
	for _, dt := range docTypes {
		doc, err := uc.documents.FindByDriverAndType(ctx, vv.DriverID, dt)
		if err != nil {
			continue // not uploaded (e.g. optional inspection) — nothing to check
		}
		if !doc.IsExpired(now) {
			continue
		}
		return uc.expireVehicle(ctx, vv, vietnameseExpiredMessage(dt), now)
	}
	return nil
}

func (uc *GoOnlineUseCase) expireVehicle(ctx context.Context, vv *entity.VehicleVerification, reason string, now time.Time) error {
	oldValue := vehicleVerificationSnapshot(vv)
	if err := vv.Expire(reason, now); err != nil {
		// Already not Approved (a concurrent request beat us to it) — the
		// status-message check above already covers this; just surface it.
		return errors.PreconditionFailed(reason)
	}
	if err := uc.vehicleVerifications.Save(ctx, vv); err != nil {
		return err
	}
	_ = recordAudit(ctx, uc.audit, entity.AuditEntityVehicleVerification, vv.ID, vv.DriverID, entity.AuditActionExpire, "system", oldValue, vehicleVerificationSnapshot(vv), reason)
	return errors.PreconditionFailed(reason)
}

// driverVerificationStatusMessage returns (Vietnamese message, ok). ok is
// true only for KYCApproved.
func driverVerificationStatusMessage(status entity.KYCStatus) (string, bool) {
	switch status {
	case entity.KYCApproved:
		return "", true
	case entity.KYCPending, entity.KYCUnderReview:
		return "Hồ sơ cá nhân đang chờ duyệt", false
	case entity.KYCRejected:
		return "Hồ sơ cá nhân đã bị từ chối, vui lòng gửi lại", false
	case entity.KYCExpired:
		return "CCCD cần xác minh lại", false
	default:
		return "Hồ sơ cá nhân chưa được xác minh", false
	}
}

// vehicleVerificationStatusMessage returns (Vietnamese message, ok). ok is
// true only for KYCApproved.
func vehicleVerificationStatusMessage(status entity.KYCStatus) (string, bool) {
	switch status {
	case entity.KYCApproved:
		return "", true
	case entity.KYCPending, entity.KYCUnderReview:
		return "Xe đang chờ duyệt", false
	case entity.KYCRejected:
		return "Hồ sơ xe đã bị từ chối, vui lòng gửi lại", false
	case entity.KYCExpired:
		return "Hồ sơ xe đã hết hạn, cần xác minh lại", false
	default:
		return "Xe chưa được xác minh", false
	}
}

// ─── GoOfflineUseCase ────────────────────────────────────────────────────────

// GoOfflineUseCase marks a driver as offline in Redis.
type GoOfflineUseCase struct {
	repo repository.AvailabilityRepository
}

func NewGoOfflineUseCase(repo repository.AvailabilityRepository) *GoOfflineUseCase {
	return &GoOfflineUseCase{repo: repo}
}

func (uc *GoOfflineUseCase) Execute(ctx context.Context, driverID string) (*entity.AvailabilityState, error) {
	if driverID == "" {
		return nil, errors.InvalidArgument("driver id must not be empty")
	}
	if err := uc.repo.SetOffline(ctx, driverID, time.Now()); err != nil {
		return nil, err
	}
	return uc.repo.GetAvailability(ctx, driverID)
}

// ─── HeartbeatUseCase ────────────────────────────────────────────────────────

// HeartbeatUseCase refreshes the online TTL for an active driver.
// Returns CodePreconditionFailed if the driver is not currently online.
type HeartbeatUseCase struct {
	repo repository.AvailabilityRepository
}

func NewHeartbeatUseCase(repo repository.AvailabilityRepository) *HeartbeatUseCase {
	return &HeartbeatUseCase{repo: repo}
}

func (uc *HeartbeatUseCase) Execute(ctx context.Context, driverID string) (*entity.AvailabilityState, error) {
	if driverID == "" {
		return nil, errors.InvalidArgument("driver id must not be empty")
	}
	if err := uc.repo.RefreshHeartbeat(ctx, driverID, time.Now()); err != nil {
		return nil, err
	}
	return uc.repo.GetAvailability(ctx, driverID)
}

// ─── GetAvailabilityUseCase ───────────────────────────────────────────────────

// GetAvailabilityUseCase returns a driver's current online status and last-seen time.
type GetAvailabilityUseCase struct {
	repo repository.AvailabilityRepository
}

func NewGetAvailabilityUseCase(repo repository.AvailabilityRepository) *GetAvailabilityUseCase {
	return &GetAvailabilityUseCase{repo: repo}
}

func (uc *GetAvailabilityUseCase) Execute(ctx context.Context, driverID string) (*entity.AvailabilityState, error) {
	if driverID == "" {
		return nil, errors.InvalidArgument("driver id must not be empty")
	}
	return uc.repo.GetAvailability(ctx, driverID)
}
