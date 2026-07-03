// Package grpc contains the gRPC handler for the User Profile service.
package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/fairride/user/app"
	"github.com/fairride/user/domain/entity"
	"github.com/fairride/user/grpc/userpb"
	domainerrors "github.com/fairride/shared/errors"
)

// Handler implements userpb.UserProfileServiceServer.
type Handler struct {
	userpb.UnimplementedUserProfileServiceServer
	getProfile    *app.GetProfileUseCase
	updateProfile *app.UpdateProfileUseCase
}

// NewHandler constructs a Handler wired to the two use cases.
func NewHandler(get *app.GetProfileUseCase, update *app.UpdateProfileUseCase) *Handler {
	if get == nil {
		panic("user grpc: GetProfileUseCase must not be nil")
	}
	if update == nil {
		panic("user grpc: UpdateProfileUseCase must not be nil")
	}
	return &Handler{getProfile: get, updateProfile: update}
}

// GetProfile implements UserProfileServiceServer.GetProfile.
func (h *Handler) GetProfile(ctx context.Context, req *userpb.GetProfileRequest) (*userpb.GetProfileResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	profile, err := h.getProfile.Execute(ctx, req.GetUserId())
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &userpb.GetProfileResponse{Profile: toProto(profile)}, nil
}

// UpdateProfile implements UserProfileServiceServer.UpdateProfile.
func (h *Handler) UpdateProfile(ctx context.Context, req *userpb.UpdateProfileRequest) (*userpb.UpdateProfileResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	var dob entity.UserProfile // dummy for field access — use time directly below
	_ = dob

	input := app.UpdateProfileInput{
		UserID:   req.GetUserId(),
		FullName: req.GetFullName(),
		Email:    req.GetEmail(),
		Avatar:   req.GetAvatar(),
		Gender:   entity.Gender(req.GetGender()),
	}
	if req.DateOfBirth != nil {
		input.DateOfBirth = req.DateOfBirth.AsTime().UTC()
	}

	profile, err := h.updateProfile.Execute(ctx, input)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &userpb.UpdateProfileResponse{Profile: toProto(profile)}, nil
}

// ─── private helpers ──────────────────────────────────────────────────────────

// toProto converts a domain UserProfile to the proto message.
func toProto(p *entity.UserProfile) *userpb.UserProfileProto {
	proto := &userpb.UserProfileProto{
		Id:        p.ID,
		FullName:  p.FullName,
		Phone:     p.Phone,
		Email:     p.Email,
		Avatar:    p.Avatar,
		Gender:    string(p.Gender),
		Status:    string(p.Status),
		CreatedAt: timestamppb.New(p.CreatedAt),
		UpdatedAt: timestamppb.New(p.UpdatedAt),
	}
	if !p.DateOfBirth.IsZero() {
		proto.DateOfBirth = timestamppb.New(p.DateOfBirth)
	}
	return proto
}

// toGRPCError maps a DomainError to an appropriate gRPC status error.
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
