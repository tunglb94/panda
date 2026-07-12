package handlers

import (
	"net/http"
	"time"

	"github.com/fairride/gateway/http/middleware"
	notificationapp "github.com/fairride/notification/app"
	notificationentity "github.com/fairride/notification/domain/entity"
)

// NotificationHandler exposes the in-app notification feed (Part 3) over HTTP.
type NotificationHandler struct {
	list     *notificationapp.ListNotificationsUseCase
	markRead *notificationapp.MarkNotificationReadUseCase
}

func NewNotificationHandler(list *notificationapp.ListNotificationsUseCase, markRead *notificationapp.MarkNotificationReadUseCase) *NotificationHandler {
	return &NotificationHandler{list: list, markRead: markRead}
}

func (h *NotificationHandler) configured() bool { return h != nil && h.list != nil }

// ListNotifications handles GET /api/v1/notifications.
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	if !h.configured() {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "notification service not configured"})
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	result, err := h.list.Execute(r.Context(), claims.UserID, 50)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	items := make([]map[string]any, len(result.Items))
	for i, n := range result.Items {
		items[i] = notificationJSON(n)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"notifications": items,
		"unread_count":  result.UnreadCount,
	})
}

// MarkRead handles POST /api/v1/notifications/{id}/read.
func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	if !h.configured() {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "notification service not configured"})
		return
	}
	id := r.PathValue("id")
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeBadRequest(w, "missing auth claims")
		return
	}
	if err := h.markRead.Execute(r.Context(), id, claims.UserID); err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func notificationJSON(n *notificationentity.Notification) map[string]any {
	body := map[string]any{
		"id":              n.ID,
		"category":        string(n.Category),
		"title":           n.Title,
		"body":            n.Body,
		"trip_id":         n.TripID,
		"conversation_id": n.ConversationID,
		"created_at":      n.CreatedAt.UTC().Format(time.RFC3339),
		"is_read":         n.ReadAt != nil,
	}
	if n.ReadAt != nil {
		body["read_at"] = n.ReadAt.UTC().Format(time.RFC3339)
	}
	return body
}
