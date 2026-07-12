package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/fairride/gateway/http/middleware"
	notificationapp "github.com/fairride/notification/app"
	notificationentity "github.com/fairride/notification/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
	"github.com/fairride/trip/grpc/trippb"
)

// tripReaderAdapter adapts the gateway's existing TripStatusClient (a thin
// wrapper over trippb.TripServiceClient, already used by BookingHandler) to
// notificationapp.TripReader — the notification app layer's own read-only
// trip dependency. Kept here rather than in the notification module so that
// module has zero dependency on any specific gRPC proto package.
type tripReaderAdapter struct {
	client TripStatusClient
}

// NewTripReader adapts a TripStatusClient into notificationapp.TripReader
// for use by the composition root (gateway/cmd/server/main.go), which lives
// in package main and so cannot construct the unexported tripReaderAdapter
// type directly.
func NewTripReader(client TripStatusClient) notificationapp.TripReader {
	return tripReaderAdapter{client: client}
}

func (a tripReaderAdapter) GetTrip(ctx context.Context, tripID string) (notificationapp.TripSnapshot, error) {
	if a.client == nil {
		return notificationapp.TripSnapshot{}, domainerrors.Unavailable("trip service not configured")
	}
	resp, err := a.client.GetTrip(ctx, &trippb.GetTripRequest{TripId: tripID})
	if err != nil {
		return notificationapp.TripSnapshot{}, domainerrors.NotFound("trip not found")
	}
	t := resp.GetTrip()
	if t.GetTripId() == "" {
		return notificationapp.TripSnapshot{}, domainerrors.NotFound("trip not found")
	}
	tripType := t.GetTripType()
	if tripType == "" {
		tripType = "ride"
	}
	return notificationapp.TripSnapshot{
		TripID:   t.GetTripId(),
		RiderID:  t.GetRiderId(),
		DriverID: t.GetDriverId(),
		Status:   t.GetStatus(),
		TripType: tripType,
	}, nil
}

// ChatHandler exposes the Communication Module's in-app chat over HTTP.
// All logic runs in-process (see notification module's report, "Kien truc")
// — there is no separate gRPC hop, so this handler owns the use cases
// directly rather than a thin proxy client.
type ChatHandler struct {
	getOrCreate   *notificationapp.GetOrCreateConversationUseCase
	sendMessage   *notificationapp.SendMessageUseCase
	listMessages  *notificationapp.ListMessagesUseCase
	pollMessages  *notificationapp.PollMessagesUseCase
	markRead      *notificationapp.MarkReadUseCase
	unreadCounter interface {
		CountUnread(ctx context.Context, conversationID, recipientID string) (int, error)
	}
}

func NewChatHandler(
	getOrCreate *notificationapp.GetOrCreateConversationUseCase,
	sendMessage *notificationapp.SendMessageUseCase,
	listMessages *notificationapp.ListMessagesUseCase,
	pollMessages *notificationapp.PollMessagesUseCase,
	markRead *notificationapp.MarkReadUseCase,
	unreadCounter interface {
		CountUnread(ctx context.Context, conversationID, recipientID string) (int, error)
	},
) *ChatHandler {
	return &ChatHandler{
		getOrCreate:   getOrCreate,
		sendMessage:   sendMessage,
		listMessages:  listMessages,
		pollMessages:  pollMessages,
		markRead:      markRead,
		unreadCounter: unreadCounter,
	}
}

func (h *ChatHandler) configured() bool { return h != nil && h.getOrCreate != nil }

// GetConversation handles GET /api/v1/rides/{tripID}/conversation.
// Lazily creates the conversation on first access (Part 2).
func (h *ChatHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	if !h.configured() {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "chat service not configured"})
		return
	}
	tripID := r.PathValue("tripID")
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	conv, err := h.getOrCreate.Execute(r.Context(), tripID, claims.UserID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	unread := 0
	if h.unreadCounter != nil {
		unread, _ = h.unreadCounter.CountUnread(r.Context(), conv.ID, claims.UserID)
	}
	writeJSON(w, http.StatusOK, conversationJSON(conv, unread))
}

// ListOrPollMessages handles GET /api/v1/conversations/{id}/messages?since_id=&poll=true.
// When poll=true, this genuinely holds the request open (long-poll — see
// notificationapp.PollMessagesUseCase) instead of returning immediately.
func (h *ChatHandler) ListOrPollMessages(w http.ResponseWriter, r *http.Request) {
	if !h.configured() {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "chat service not configured"})
		return
	}
	conversationID := r.PathValue("id")
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	sinceSeq, _ := strconv.ParseInt(r.URL.Query().Get("since_id"), 10, 64)

	var (
		msgs []*notificationentity.Message
		err  error
	)
	if r.URL.Query().Get("poll") == "true" {
		msgs, err = h.pollMessages.Execute(r.Context(), conversationID, claims.UserID, sinceSeq, notificationapp.DefaultPollTimeout)
	} else {
		msgs, err = h.listMessages.Execute(r.Context(), conversationID, claims.UserID, sinceSeq)
	}
	if err != nil {
		writeDomainError(w, err)
		return
	}
	items := make([]map[string]any, len(msgs))
	for i, m := range msgs {
		items[i] = messageJSON(m)
	}
	writeJSON(w, http.StatusOK, map[string]any{"messages": items})
}

type sendMessageRequest struct {
	Text          string `json:"text"`
	QuickReplyKey string `json:"quick_reply_key"`
}

// SendMessage handles POST /api/v1/conversations/{id}/messages.
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	if !h.configured() {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "chat service not configured"})
		return
	}
	conversationID := r.PathValue("id")
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	msg, err := h.sendMessage.Execute(r.Context(), notificationapp.SendMessageInput{
		ConversationID: conversationID,
		SenderID:       claims.UserID,
		Text:           req.Text,
		QuickReplyKey:  req.QuickReplyKey,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, messageJSON(msg))
}

// MarkConversationRead handles POST /api/v1/conversations/{id}/read.
func (h *ChatHandler) MarkConversationRead(w http.ResponseWriter, r *http.Request) {
	if !h.configured() {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "chat service not configured"})
		return
	}
	conversationID := r.PathValue("id")
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	if err := h.markRead.Execute(r.Context(), conversationID, claims.UserID); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func conversationJSON(c *notificationentity.Conversation, unreadCount int) map[string]any {
	body := map[string]any{
		"id":           c.ID,
		"trip_id":      c.TripID,
		"rider_id":     c.RiderID,
		"driver_id":    c.DriverID,
		"trip_type":    c.TripType,
		"status":       string(c.Status),
		"created_at":   c.CreatedAt.UTC().Format(time.RFC3339),
		"unread_count": unreadCount,
	}
	if c.ClosedAt != nil {
		body["closed_at"] = c.ClosedAt.UTC().Format(time.RFC3339)
	}
	return body
}

func messageJSON(m *notificationentity.Message) map[string]any {
	body := map[string]any{
		"id":              m.ID,
		"seq":             m.Seq,
		"conversation_id": m.ConversationID,
		"sender_id":       m.SenderID,
		"sender_role":     string(m.SenderRole),
		"body":            m.Body,
		"quick_reply_key": m.QuickReplyKey,
		"created_at":      m.CreatedAt.UTC().Format(time.RFC3339),
	}
	if m.DeliveredAt != nil {
		body["delivered_at"] = m.DeliveredAt.UTC().Format(time.RFC3339)
	}
	if m.ReadAt != nil {
		body["read_at"] = m.ReadAt.UTC().Format(time.RFC3339)
	}
	return body
}
