// Package grpc contains the gRPC handler for the Wallet service.
package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/fairride/wallet/app"
	"github.com/fairride/wallet/domain/entity"
	"github.com/fairride/wallet/grpc/walletpb"
	domainerrors "github.com/fairride/shared/errors"
)

// Handler implements walletpb.WalletServiceServer.
type Handler struct {
	walletpb.UnimplementedWalletServiceServer
	getWallet      *app.GetWalletUseCase
	getBalance     *app.GetBalanceUseCase
	getLedger      *app.GetLedgerUseCase
	getTransaction *app.GetTransactionUseCase
}

// NewHandler wires all wallet use cases into a gRPC handler.
func NewHandler(
	getWallet *app.GetWalletUseCase,
	getBalance *app.GetBalanceUseCase,
	getLedger *app.GetLedgerUseCase,
	getTransaction *app.GetTransactionUseCase,
) *Handler {
	if getWallet == nil {
		panic("wallet grpc: GetWalletUseCase must not be nil")
	}
	if getBalance == nil {
		panic("wallet grpc: GetBalanceUseCase must not be nil")
	}
	if getLedger == nil {
		panic("wallet grpc: GetLedgerUseCase must not be nil")
	}
	if getTransaction == nil {
		panic("wallet grpc: GetTransactionUseCase must not be nil")
	}
	return &Handler{
		getWallet:      getWallet,
		getBalance:     getBalance,
		getLedger:      getLedger,
		getTransaction: getTransaction,
	}
}

// ─── RPCs ─────────────────────────────────────────────────────────────────────

func (h *Handler) GetWallet(ctx context.Context, req *walletpb.GetWalletRequest) (*walletpb.WalletResponse, error) {
	w, err := h.getWallet.Execute(ctx, req.GetOwnerId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &walletpb.WalletResponse{Wallet: toWalletProto(w)}, nil
}

func (h *Handler) GetBalance(ctx context.Context, req *walletpb.GetBalanceRequest) (*walletpb.BalanceResponse, error) {
	result, err := h.getBalance.Execute(ctx, req.GetOwnerId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &walletpb.BalanceResponse{
		WalletId:     result.WalletID,
		OwnerId:      result.OwnerID,
		BalanceCents: result.BalanceCents,
		Currency:     result.Currency,
	}, nil
}

func (h *Handler) GetLedger(ctx context.Context, req *walletpb.GetLedgerRequest) (*walletpb.LedgerResponse, error) {
	entries, err := h.getLedger.Execute(ctx, req.GetWalletId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	protos := make([]*walletpb.LedgerEntryProto, len(entries))
	for i := range entries {
		protos[i] = toLedgerEntryProto(&entries[i])
	}
	return &walletpb.LedgerResponse{Entries: protos}, nil
}

func (h *Handler) GetTransaction(ctx context.Context, req *walletpb.GetTransactionRequest) (*walletpb.TransactionResponse, error) {
	result, err := h.getTransaction.Execute(ctx, req.GetTransactionId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	protos := make([]*walletpb.LedgerEntryProto, len(result.Entries))
	for i := range result.Entries {
		protos[i] = toLedgerEntryProto(&result.Entries[i])
	}
	return &walletpb.TransactionResponse{
		Transaction: toTransactionProto(result.Transaction),
		Entries:     protos,
	}, nil
}

// ─── converters ───────────────────────────────────────────────────────────────

func toWalletProto(w *entity.Wallet) *walletpb.WalletProto {
	return &walletpb.WalletProto{
		WalletId:   w.WalletID,
		OwnerId:    w.OwnerID,
		WalletType: string(w.WalletType),
		Currency:   w.Currency,
		CreatedAt:  timestamppb.New(w.CreatedAt),
		UpdatedAt:  timestamppb.New(w.UpdatedAt),
	}
}

func toLedgerEntryProto(e *entity.LedgerEntry) *walletpb.LedgerEntryProto {
	return &walletpb.LedgerEntryProto{
		EntryId:       e.EntryID,
		WalletId:      e.WalletID,
		TransactionId: e.TransactionID,
		Direction:     string(e.Direction),
		AmountCents:   e.AmountCents,
		Currency:      e.Currency,
		Description:   e.Description,
		CreatedAt:     timestamppb.New(e.CreatedAt),
	}
}

func toTransactionProto(tx *entity.Transaction) *walletpb.TransactionProto {
	return &walletpb.TransactionProto{
		TransactionId: tx.TransactionID,
		Type:          string(tx.Type),
		ReferenceId:   tx.ReferenceID,
		Currency:      tx.Currency,
		Description:   tx.Description,
		CreatedAt:     timestamppb.New(tx.CreatedAt),
	}
}

func toGRPCError(err error) error {
	de, ok := err.(*domainerrors.DomainError)
	if !ok {
		return status.Error(codes.Internal, err.Error())
	}
	switch de.Code {
	case domainerrors.CodeNotFound:
		return status.Error(codes.NotFound, de.Message)
	case domainerrors.CodeInvalidArgument:
		return status.Error(codes.InvalidArgument, de.Message)
	case domainerrors.CodeAlreadyExists:
		return status.Error(codes.AlreadyExists, de.Message)
	case domainerrors.CodePreconditionFailed:
		return status.Error(codes.FailedPrecondition, de.Message)
	case domainerrors.CodePermissionDenied:
		return status.Error(codes.PermissionDenied, de.Message)
	default:
		return status.Error(codes.Internal, de.Message)
	}
}
