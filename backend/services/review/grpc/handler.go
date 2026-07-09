// Package grpc contains the gRPC handler for the Review service.
package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/fairride/review/app"
	"github.com/fairride/review/domain/entity"
	"github.com/fairride/review/grpc/reviewpb"
	domainerrors "github.com/fairride/shared/errors"
)

// Handler implements reviewpb.ReviewServiceServer.
type Handler struct {
	reviewpb.UnimplementedReviewServiceServer
	submitRating   *app.SubmitRatingUseCase
	getTripRating  *app.GetTripRatingUseCase
}

func NewHandler(submit *app.SubmitRatingUseCase, get *app.GetTripRatingUseCase) *Handler {
	if submit == nil {
		panic("review grpc: SubmitRatingUseCase must not be nil")
	}
	if get == nil {
		panic("review grpc: GetTripRatingUseCase must not be nil")
	}
	return &Handler{submitRating: submit, getTripRating: get}
}

// SubmitRating implements ReviewServiceServer.SubmitRating.
func (h *Handler) SubmitRating(ctx context.Context, req *reviewpb.SubmitRatingRequest) (*reviewpb.RatingProto, error) {
	rating, err := h.submitRating.Execute(ctx, app.SubmitRatingInput{
		TripID:  req.GetTripId(),
		RaterID: req.GetRaterId(),
		RateeID: req.GetRateeId(),
		Role:    req.GetRole(),
		Stars:   req.GetStars(),
		Comment: req.GetComment(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(rating), nil
}

// GetTripRating implements ReviewServiceServer.GetTripRating.
func (h *Handler) GetTripRating(ctx context.Context, req *reviewpb.GetRatingRequest) (*reviewpb.RatingProto, error) {
	if req.GetTripId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trip_id is required")
	}
	if req.GetRole() == "" {
		return nil, status.Error(codes.InvalidArgument, "role is required")
	}
	rating, err := h.getTripRating.Execute(ctx, req.GetTripId(), req.GetRole())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(rating), nil
}

func toProto(r *entity.Rating) *reviewpb.RatingProto {
	return &reviewpb.RatingProto{
		RatingId:  r.RatingID,
		TripId:    r.TripID,
		RaterId:   r.RaterID,
		RateeId:   r.RateeID,
		Role:      string(r.Role),
		Stars:     r.Stars,
		Comment:   r.Comment,
		CreatedAt: timestamppb.New(r.CreatedAt),
	}
}

func toGRPCError(err error) error {
	code := domainerrors.GetCode(err)
	var grpcCode codes.Code
	switch code {
	case domainerrors.CodeNotFound:
		grpcCode = codes.NotFound
	case domainerrors.CodeAlreadyExists:
		grpcCode = codes.AlreadyExists
	case domainerrors.CodeInvalidArgument:
		grpcCode = codes.InvalidArgument
	default:
		grpcCode = codes.Internal
	}
	return status.Error(grpcCode, err.Error())
}
