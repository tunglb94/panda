package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	driverentity "github.com/fairride/driver/domain/entity"
	"github.com/fairride/gateway/http/handlers"
	identityentity "github.com/fairride/identity/domain/entity"
	"github.com/fairride/identity/infrastructure/jwt"
	notificationapp "github.com/fairride/notification/app"
	notificationentity "github.com/fairride/notification/domain/entity"
	reviewentity "github.com/fairride/review/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
	"github.com/fairride/trip/grpc/trippb"
	"google.golang.org/grpc"
)

// ─── in-memory fakes shared by chat/call/notification handler tests ───────

type fakeTripReader struct {
	trips map[string]notificationapp.TripSnapshot
}

func (f *fakeTripReader) GetTrip(_ context.Context, tripID string) (notificationapp.TripSnapshot, error) {
	t, ok := f.trips[tripID]
	if !ok {
		return notificationapp.TripSnapshot{}, domainerrors.NotFound("trip not found")
	}
	return t, nil
}

type fakeConvRepo struct {
	mu    sync.Mutex
	byID  map[string]*notificationentity.Conversation
	byTrp map[string]string
}

func newFakeConvRepo() *fakeConvRepo {
	return &fakeConvRepo{byID: map[string]*notificationentity.Conversation{}, byTrp: map[string]string{}}
}
func (r *fakeConvRepo) Save(_ context.Context, c *notificationentity.Conversation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byTrp[c.TripID]; exists {
		return domainerrors.AlreadyExists("conversation exists")
	}
	cp := *c
	r.byID[c.ID] = &cp
	r.byTrp[c.TripID] = c.ID
	return nil
}
func (r *fakeConvRepo) FindByTripID(_ context.Context, tripID string) (*notificationentity.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.byTrp[tripID]
	if !ok {
		return nil, domainerrors.NotFound("conversation not found")
	}
	cp := *r.byID[id]
	return &cp, nil
}
func (r *fakeConvRepo) FindByID(_ context.Context, id string) (*notificationentity.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.byID[id]
	if !ok {
		return nil, domainerrors.NotFound("conversation not found")
	}
	cp := *c
	return &cp, nil
}
func (r *fakeConvRepo) Update(_ context.Context, c *notificationentity.Conversation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *c
	r.byID[c.ID] = &cp
	return nil
}

type fakeMsgRepo struct {
	mu   sync.Mutex
	msgs []*notificationentity.Message
	seq  int64
}

func (r *fakeMsgRepo) Save(_ context.Context, m *notificationentity.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	m.Seq = r.seq
	cp := *m
	r.msgs = append(r.msgs, &cp)
	return nil
}
func (r *fakeMsgRepo) ListSince(_ context.Context, conversationID string, sinceSeq int64, limit int) ([]*notificationentity.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*notificationentity.Message
	for _, m := range r.msgs {
		if m.ConversationID == conversationID && m.Seq > sinceSeq {
			out = append(out, m)
		}
	}
	return out, nil
}
func (r *fakeMsgRepo) MarkReadByRecipient(_ context.Context, conversationID, recipientID string, now time.Time) error {
	return nil
}
func (r *fakeMsgRepo) CountUnread(_ context.Context, conversationID, recipientID string) (int, error) {
	return 0, nil
}

type fakeNotifRepo struct {
	mu    sync.Mutex
	items []*notificationentity.Notification
}

func (r *fakeNotifRepo) Save(_ context.Context, n *notificationentity.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, n)
	return nil
}
func (r *fakeNotifRepo) ListByUser(_ context.Context, userID string, limit int) ([]*notificationentity.Notification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*notificationentity.Notification
	for _, n := range r.items {
		if n.UserID == userID {
			out = append(out, n)
		}
	}
	return out, nil
}
func (r *fakeNotifRepo) MarkRead(_ context.Context, id, userID string, now time.Time) error {
	return nil
}
func (r *fakeNotifRepo) CountUnread(_ context.Context, userID string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, n := range r.items {
		if n.UserID == userID && n.ReadAt == nil {
			count++
		}
	}
	return count, nil
}

// ─── ChatHandler ────────────────────────────────────────────────────────────

func TestChatHandler_ServiceUnavailableWhenNotConfigured(t *testing.T) {
	h := handlers.NewChatHandler(nil, nil, nil, nil, nil, nil)
	w := httptest.NewRecorder()
	r := authedRequest(t, http.MethodGet, "/api/v1/rides/trip1/conversation", nil)
	r.SetPathValue("tripID", "trip1")
	h.GetConversation(w, r)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", w.Code)
	}
}

func buildChatHandler(trips *fakeTripReader) (*handlers.ChatHandler, *fakeMsgRepo) {
	convRepo := newFakeConvRepo()
	msgRepo := &fakeMsgRepo{}
	notifRepo := &fakeNotifRepo{}
	notify := notificationapp.NewCreateNotificationUseCase(notifRepo, nil)
	broadcaster := notificationapp.NewBroadcaster()
	h := handlers.NewChatHandler(
		notificationapp.NewGetOrCreateConversationUseCase(convRepo, trips),
		notificationapp.NewSendMessageUseCase(convRepo, msgRepo, notify, broadcaster),
		notificationapp.NewListMessagesUseCase(convRepo, msgRepo),
		notificationapp.NewPollMessagesUseCase(convRepo, msgRepo, broadcaster),
		notificationapp.NewMarkReadUseCase(convRepo, msgRepo),
		msgRepo,
	)
	return h, msgRepo
}

func TestChatHandler_GetConversation_CreatesOnFirstAccess(t *testing.T) {
	trips := &fakeTripReader{trips: map[string]notificationapp.TripSnapshot{
		"trip1": {TripID: "trip1", RiderID: "d1", DriverID: "driver1", Status: "in_progress", TripType: "ride"},
	}}
	h, _ := buildChatHandler(trips)

	w := httptest.NewRecorder()
	r := authedRequest(t, http.MethodGet, "/api/v1/rides/trip1/conversation", nil)
	r.SetPathValue("tripID", "trip1")
	h.GetConversation(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["trip_id"] != "trip1" || body["status"] != "open" {
		t.Errorf("unexpected body: %+v", body)
	}
}

func TestChatHandler_GetConversation_ForbiddenForNonParticipant(t *testing.T) {
	trips := &fakeTripReader{trips: map[string]notificationapp.TripSnapshot{
		"trip1": {TripID: "trip1", RiderID: "someone-else", DriverID: "driver1", Status: "in_progress", TripType: "ride"},
	}}
	h, _ := buildChatHandler(trips)

	w := httptest.NewRecorder()
	// authedRequest injects claims.UserID = "d1", which is neither rider nor driver here.
	r := authedRequest(t, http.MethodGet, "/api/v1/rides/trip1/conversation", nil)
	r.SetPathValue("tripID", "trip1")
	h.GetConversation(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestChatHandler_SendMessage_ThenListMessages(t *testing.T) {
	trips := &fakeTripReader{trips: map[string]notificationapp.TripSnapshot{
		"trip1": {TripID: "trip1", RiderID: "d1", DriverID: "driver1", Status: "in_progress", TripType: "ride"},
	}}
	h, _ := buildChatHandler(trips)

	// Create the conversation first.
	w := httptest.NewRecorder()
	r := authedRequest(t, http.MethodGet, "/api/v1/rides/trip1/conversation", nil)
	r.SetPathValue("tripID", "trip1")
	h.GetConversation(w, r)
	var conv map[string]any
	_ = json.NewDecoder(w.Body).Decode(&conv)
	convID := conv["id"].(string)

	// Send a message as the rider (claims.UserID = "d1" from authedRequest).
	sendW := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"text": "Tôi tới rồi"})
	sendR := httptest.NewRequest(http.MethodPost, "/api/v1/conversations/"+convID+"/messages", bytes.NewReader(body))
	sendR = injectClaims(sendR, &jwt.AccessClaims{UserID: "d1"})
	sendR.SetPathValue("id", convID)
	h.SendMessage(sendW, sendR)
	if sendW.Code != http.StatusCreated {
		t.Fatalf("send status = %d, want 201, body=%s", sendW.Code, sendW.Body.String())
	}

	// List messages as the driver.
	listW := httptest.NewRecorder()
	listR := httptest.NewRequest(http.MethodGet, "/api/v1/conversations/"+convID+"/messages", nil)
	listR = injectClaims(listR, &jwt.AccessClaims{UserID: "driver1"})
	listR.SetPathValue("id", convID)
	h.ListOrPollMessages(listW, listR)
	if listW.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200", listW.Code)
	}
	var listBody map[string]any
	_ = json.NewDecoder(listW.Body).Decode(&listBody)
	msgs, _ := listBody["messages"].([]any)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
}

// ─── CallHandler ────────────────────────────────────────────────────────────

type fakeUserByID struct {
	users map[string]*identityentity.User
}

func (f *fakeUserByID) FindByID(_ context.Context, id string) (*identityentity.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, domainerrors.NotFound("user not found")
	}
	return u, nil
}

type fakeDriverByID struct {
	profiles map[string]*driverentity.DriverProfile
}

func (f *fakeDriverByID) FindByID(_ context.Context, driverID string) (*driverentity.DriverProfile, error) {
	p, ok := f.profiles[driverID]
	if !ok {
		return nil, domainerrors.NotFound("driver not found")
	}
	return p, nil
}

// stubTripStatusClientForCall implements TripStatusClient for CallHandler tests.
type stubTripStatusClientForCall struct {
	trip *trippb.TripProto
	err  error
}

func (s *stubTripStatusClientForCall) GetTrip(_ context.Context, _ *trippb.GetTripRequest, _ ...grpc.CallOption) (*trippb.TripResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &trippb.TripResponse{Trip: s.trip}, nil
}

type fakeAvgRating struct {
	avg   float64
	count int32
}

func (f fakeAvgRating) Execute(_ context.Context, _ string, _ reviewentity.Role) (float64, int32, error) {
	return f.avg, f.count, nil
}

type fakeApprovedDriverVerification struct{}

func (fakeApprovedDriverVerification) Execute(_ context.Context, driverID string) (*driverentity.DriverVerification, error) {
	return &driverentity.DriverVerification{DriverID: driverID, Status: driverentity.KYCApproved}, nil
}

type fakeApprovedVehicleVerification struct{}

func (fakeApprovedVehicleVerification) Execute(_ context.Context, driverID string) (*driverentity.VehicleVerification, error) {
	return &driverentity.VehicleVerification{DriverID: driverID, Status: driverentity.KYCApproved}, nil
}

func TestCallHandler_ServiceUnavailableWhenNotConfigured(t *testing.T) {
	h := handlers.NewCallHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	w := httptest.NewRecorder()
	r := authedRequest(t, http.MethodGet, "/api/v1/rides/trip1/contact", nil)
	r.SetPathValue("tripID", "trip1")
	h.GetContact(w, r)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", w.Code)
	}
}

func TestCallHandler_GetContact_ReturnsMaskedPhoneNotRealNumber(t *testing.T) {
	tripStub := &stubTripStatusClientForCall{trip: &trippb.TripProto{
		TripId: "trip1", RiderId: "d1", DriverId: "driverX", Status: "in_progress", TripType: "ride",
	}}
	users := &fakeUserByID{users: map[string]*identityentity.User{
		"user-of-driverX": {ID: "user-of-driverX", Name: "Anh Tài", PhoneNumber: "0901234567"},
	}}
	drivers := &fakeDriverByID{profiles: map[string]*driverentity.DriverProfile{
		"driverX": {DriverID: "driverX", UserID: "user-of-driverX", VehicleType: driverentity.VehicleTypeMotorcycle, PlateNumber: "59-A1 12345"},
	}}
	h := handlers.NewCallHandler(
		tripStub, users, drivers, fakeAvgRating{avg: 4.8, count: 12}, nil,
		fakeApprovedDriverVerification{}, fakeApprovedVehicleVerification{}, nil,
	)

	w := httptest.NewRecorder()
	r := authedRequest(t, http.MethodGet, "/api/v1/rides/trip1/contact", nil)
	r.SetPathValue("tripID", "trip1")
	h.GetContact(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["masked_phone"] == "0901234567" {
		t.Error("masked_phone must never equal the real phone number")
	}
	if body["masked_phone"] != "090****567" {
		t.Errorf("masked_phone = %v, want 090****567", body["masked_phone"])
	}
	if body["name"] != "Anh Tài" {
		t.Errorf("name = %v, want Anh Tài", body["name"])
	}
}

// ─── NotificationHandler ────────────────────────────────────────────────────

func TestNotificationHandler_ServiceUnavailableWhenNotConfigured(t *testing.T) {
	h := handlers.NewNotificationHandler(nil, nil)
	w := httptest.NewRecorder()
	h.ListNotifications(w, authedRequest(t, http.MethodGet, "/api/v1/notifications", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", w.Code)
	}
}

func TestNotificationHandler_ListNotifications(t *testing.T) {
	repo := &fakeNotifRepo{}
	n, _ := notificationentity.NewNotification("n1", "d1", notificationentity.CategoryChat, "Tin nhắn mới", "hello", "trip1", "conv1", time.Now().UTC())
	_ = repo.Save(context.Background(), n)
	h := handlers.NewNotificationHandler(
		notificationapp.NewListNotificationsUseCase(repo),
		notificationapp.NewMarkNotificationReadUseCase(repo),
	)

	w := httptest.NewRecorder()
	h.ListNotifications(w, authedRequest(t, http.MethodGet, "/api/v1/notifications", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]any
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["unread_count"].(float64) != 1 {
		t.Errorf("unread_count = %v, want 1", body["unread_count"])
	}
}
