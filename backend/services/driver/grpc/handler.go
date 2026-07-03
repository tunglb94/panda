// Package grpc contains the gRPC handler for the Driver Profile service.
package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/fairride/driver/app"
	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/grpc/driverpb"
	domainerrors "github.com/fairride/shared/errors"
)

// Handler implements driverpb.DriverProfileServiceServer.
type Handler struct {
	driverpb.UnimplementedDriverProfileServiceServer
	getDriver              *app.GetDriverProfileUseCase
	getDriverByUser        *app.GetDriverProfileByUserIDUseCase
	updateDriver           *app.UpdateDriverProfileUseCase
	updateOnlineStatus     *app.UpdateOnlineStatusUseCase
	updateVerification     *app.UpdateVerificationStatusUseCase
}

// NewHandler constructs a Handler wired to all five use cases.
func NewHandler(
	getDriver *app.GetDriverProfileUseCase,
	getDriverByUser *app.GetDriverProfileByUserIDUseCase,
	updateDriver *app.UpdateDriverProfileUseCase,
	updateOnlineStatus *app.UpdateOnlineStatusUseCase,
	updateVerification *app.UpdateVerificationStatusUseCase,
) *Handler {
	if getDriver == nil {
		panic("driver grpc: GetDriverProfileUseCase must not be nil")
	}
	if getDriverByUser == nil {
		panic("driver grpc: GetDriverProfileByUserIDUseCase must not be nil")
	}
	if updateDriver == nil {
		panic("driver grpc: UpdateDriverProfileUseCase must not be nil")
	}
	if updateOnlineStatus == nil {
		panic("driver grpc: UpdateOnlineStatusUseCase must not be nil")
	}
	if updateVerification == nil {
		panic("driver grpc: UpdateVerificationStatusUseCase must not be nil")
	}
	return &Handler{
		getDriver:          getDriver,
		getDriverByUser:    getDriverByUser,
		updateDriver:       updateDriver,
		updateOnlineStatus: updateOnlineStatus,
		updateVerification: updateVerification,
	}
}

// GetDriverProfile implements DriverProfileServiceServer.GetDriverProfile.
func (h *Handler) GetDriverProfile(ctx context.Context, req *driverpb.GetDriverProfileRequest) (*driverpb.GetDriverProfileResponse, error) {
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	d, err := h.getDriver.Execute(ctx, req.GetDriverId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &driverpb.GetDriverProfileResponse{Profile: toProto(d)}, nil
}

// GetDriverProfileByUserID implements DriverProfileServiceServer.GetDriverProfileByUserID.
func (h *Handler) GetDriverProfileByUserID(ctx context.Context, req *driverpb.GetDriverProfileByUserIDRequest) (*driverpb.GetDriverProfileResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	d, err := h.getDriverByUser.Execute(ctx, req.GetUserId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &driverpb.GetDriverProfileResponse{Profile: toProto(d)}, nil
}

// UpdateDriverProfile implements DriverProfileServiceServer.UpdateDriverProfile.
func (h *Handler) UpdateDriverProfile(ctx context.Context, req *driverpb.UpdateDriverProfileRequest) (*driverpb.UpdateDriverProfileResponse, error) {
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	in := app.UpdateDriverProfileInput{
		DriverID:      req.GetDriverId(),
		LicenseNumber: req.GetLicenseNumber(),
		VehicleType:   entity.VehicleType(req.GetVehicleType()),
		VehicleBrand:  req.GetVehicleBrand(),
		VehicleModel:  req.GetVehicleModel(),
		VehicleColor:  req.GetVehicleColor(),
		PlateNumber:   req.GetPlateNumber(),
	}
	d, err := h.updateDriver.Execute(ctx, in)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &driverpb.UpdateDriverProfileResponse{Profile: toProto(d)}, nil
}

// UpdateOnlineStatus implements DriverProfileServiceServer.UpdateOnlineStatus.
func (h *Handler) UpdateOnlineStatus(ctx context.Context, req *driverpb.UpdateOnlineStatusRequest) (*driverpb.UpdateOnlineStatusResponse, error) {
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	d, err := h.updateOnlineStatus.Execute(ctx, req.GetDriverId(), entity.OnlineStatus(req.GetOnlineStatus()))
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &driverpb.UpdateOnlineStatusResponse{Profile: toProto(d)}, nil
}

// UpdateVerificationStatus implements DriverProfileServiceServer.UpdateVerificationStatus.
func (h *Handler) UpdateVerificationStatus(ctx context.Context, req *driverpb.UpdateVerificationStatusRequest) (*driverpb.UpdateVerificationStatusResponse, error) {
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	d, err := h.updateVerification.Execute(ctx, req.GetDriverId(), app.VerificationAction(req.GetVerificationStatus()))
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &driverpb.UpdateVerificationStatusResponse{Profile: toProto(d)}, nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func toProto(d *entity.DriverProfile) *driverpb.DriverProfileProto {
	return &driverpb.DriverProfileProto{
		DriverId:           d.DriverID,
		UserId:             d.UserID,
		LicenseNumber:      d.LicenseNumber,
		VehicleType:        string(d.VehicleType),
		VehicleBrand:       d.VehicleBrand,
		VehicleModel:       d.VehicleModel,
		VehicleColor:       d.VehicleColor,
		PlateNumber:        d.PlateNumber,
		OnlineStatus:       string(d.OnlineStatus),
		VerificationStatus: string(d.VerificationStatus),
		CreatedAt:          timestamppb.New(d.CreatedAt),
		UpdatedAt:          timestamppb.New(d.UpdatedAt),
	}
}

func toGRPCError(err error) error {
	code := domainerrors.GetCode(err)
	var grpcCode codes.Code
	switch code {
	case domainerrors.CodeNotFound:
		grpcCode = codes.NotFound
	case domainerrors.CodeInvalidArgument:
		grpcCode = codes.InvalidArgument
	case domainerrors.CodeAlreadyExists:
		grpcCode = codes.AlreadyExists
	case domainerrors.CodeUnauthenticated:
		grpcCode = codes.Unauthenticated
	case domainerrors.CodePermissionDenied:
		grpcCode = codes.PermissionDenied
	case domainerrors.CodePreconditionFailed:
		grpcCode = codes.FailedPrecondition
	case domainerrors.CodeUnavailable:
		grpcCode = codes.Unavailable
	default:
		grpcCode = codes.Internal
	}
	return status.Error(grpcCode, err.Error())
}
