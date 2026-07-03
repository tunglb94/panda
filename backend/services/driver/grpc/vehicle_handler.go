package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/fairride/driver/app"
	"github.com/fairride/driver/domain/entity"
	"github.com/fairride/driver/grpc/driverpb"
)

// VehicleHandler implements driverpb.VehicleServiceServer.
type VehicleHandler struct {
	driverpb.UnimplementedVehicleServiceServer
	create *app.CreateVehicleUseCase
	update *app.UpdateVehicleUseCase
	delete *app.DeleteVehicleUseCase
	list   *app.ListVehiclesUseCase
}

// NewVehicleHandler constructs a VehicleHandler wired to all four use cases.
func NewVehicleHandler(
	create *app.CreateVehicleUseCase,
	update *app.UpdateVehicleUseCase,
	del *app.DeleteVehicleUseCase,
	list *app.ListVehiclesUseCase,
) *VehicleHandler {
	if create == nil {
		panic("vehicle grpc: CreateVehicleUseCase must not be nil")
	}
	if update == nil {
		panic("vehicle grpc: UpdateVehicleUseCase must not be nil")
	}
	if del == nil {
		panic("vehicle grpc: DeleteVehicleUseCase must not be nil")
	}
	if list == nil {
		panic("vehicle grpc: ListVehiclesUseCase must not be nil")
	}
	return &VehicleHandler{create: create, update: update, delete: del, list: list}
}

// CreateVehicle implements VehicleServiceServer.CreateVehicle.
func (h *VehicleHandler) CreateVehicle(ctx context.Context, req *driverpb.CreateVehicleRequest) (*driverpb.CreateVehicleResponse, error) {
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	in := app.CreateVehicleInput{
		DriverID:    req.GetDriverId(),
		VehicleType: entity.VehicleType(req.GetType()),
		Brand:       req.GetBrand(),
		Model:       req.GetModel(),
		Color:       req.GetColor(),
		PlateNumber: req.GetPlateNumber(),
		Year:        int(req.GetYear()),
	}
	v, err := h.create.Execute(ctx, in)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &driverpb.CreateVehicleResponse{Vehicle: vehicleToProto(v)}, nil
}

// UpdateVehicle implements VehicleServiceServer.UpdateVehicle.
func (h *VehicleHandler) UpdateVehicle(ctx context.Context, req *driverpb.UpdateVehicleRequest) (*driverpb.UpdateVehicleResponse, error) {
	if req.GetVehicleId() == "" {
		return nil, status.Error(codes.InvalidArgument, "vehicle_id is required")
	}
	in := app.UpdateVehicleInput{
		VehicleID:   req.GetVehicleId(),
		VehicleType: entity.VehicleType(req.GetType()),
		Brand:       req.GetBrand(),
		Model:       req.GetModel(),
		Color:       req.GetColor(),
		PlateNumber: req.GetPlateNumber(),
		Year:        int(req.GetYear()),
	}
	v, err := h.update.Execute(ctx, in)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &driverpb.UpdateVehicleResponse{Vehicle: vehicleToProto(v)}, nil
}

// DeleteVehicle implements VehicleServiceServer.DeleteVehicle.
func (h *VehicleHandler) DeleteVehicle(ctx context.Context, req *driverpb.DeleteVehicleRequest) (*driverpb.DeleteVehicleResponse, error) {
	if req.GetVehicleId() == "" {
		return nil, status.Error(codes.InvalidArgument, "vehicle_id is required")
	}
	if err := h.delete.Execute(ctx, req.GetVehicleId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &driverpb.DeleteVehicleResponse{}, nil
}

// ListVehicles implements VehicleServiceServer.ListVehicles.
func (h *VehicleHandler) ListVehicles(ctx context.Context, req *driverpb.ListVehiclesRequest) (*driverpb.ListVehiclesResponse, error) {
	if req.GetDriverId() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}
	vehicles, err := h.list.Execute(ctx, req.GetDriverId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	protos := make([]*driverpb.VehicleProto, 0, len(vehicles))
	for _, v := range vehicles {
		protos = append(protos, vehicleToProto(v))
	}
	return &driverpb.ListVehiclesResponse{Vehicles: protos}, nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

func vehicleToProto(v *entity.Vehicle) *driverpb.VehicleProto {
	return &driverpb.VehicleProto{
		VehicleId:   v.VehicleID,
		DriverId:    v.DriverID,
		Type:        string(v.Type),
		Brand:       v.Brand,
		Model:       v.Model,
		Color:       v.Color,
		PlateNumber: v.PlateNumber,
		Year:        int32(v.Year),
		CreatedAt:   timestamppb.New(v.CreatedAt),
		UpdatedAt:   timestamppb.New(v.UpdatedAt),
	}
}
