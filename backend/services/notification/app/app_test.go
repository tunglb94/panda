package app_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/fairride/notification/app"
	"github.com/fairride/notification/domain/entity"
	domainerrors "github.com/fairride/shared/errors"
)

// ─── in-memory fakes ──────────────────────────────────────────────────────

type fakeTripReader struct {
	trips map[string]app.TripSnapshot
}

func (f *fakeTripReader) GetTrip(_ context.Context, tripID string) (app.TripSnapshot, error) {
	t, ok := f.trips[tripID]
	if !ok {
		return app.TripSnapshot{}, domainerrors.NotFound("trip not found")
	}
	return t, nil
}

type fakeConversationRepo struct {
	mu    sync.Mutex
	byID  map[string]*entity.Conversation
	byTrp map[string]string // tripID -> conversationID
}

func newFakeConversationRepo() *fakeConversationRepo {
	return &fakeConversationRepo{byID: map[string]*entity.Conversation{}, byTrp: map[string]string{}}
}

func (r *fakeConversationRepo) Save(_ context.Context, c *entity.Conversation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byTrp[c.TripID]; exists {
		return domainerrors.AlreadyExists("conversation already exists for this trip")
	}
	cp := *c
	r.byID[c.ID] = &cp
	r.byTrp[c.TripID] = c.ID
	return nil
}

func (r *fakeConversationRepo) FindByTripID(_ context.Context, tripID string) (*entity.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.byTrp[tripID]
	if !ok {
		return nil, domainerrors.NotFound("conversation not found")
	}
	cp := *r.byID[id]
	return &cp, nil
}

func (r *fakeConversationRepo) FindByID(_ context.Context, id string) (*entity.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.byID[id]
	if !ok {
		return nil, domainerrors.NotFound("conversation not found")
	}
	cp := *c
	return &cp, nil
}

func (r *fakeConversationRepo) Update(_ context.Context, c *entity.Conversation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byID[c.ID]; !ok {
		return domainerrors.NotFound("conversation not found")
	}
	cp := *c
	r.byID[c.ID] = &cp
	return nil
}

type fakeMessageRepo struct {
	mu   sync.Mutex
	msgs []*entity.Message
	seq  int64
}

func newFakeMessageRepo() *fakeMessageRepo { return &fakeMessageRepo{} }

func (r *fakeMessageRepo) Save(_ context.Context, m *entity.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	m.Seq = r.seq
	cp := *m
	r.msgs = append(r.msgs, &cp)
	return nil
}

func (r *fakeMessageRepo) ListSince(_ context.Context, conversationID string, sinceSeq int64, limit int) ([]*entity.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*entity.Message
	for _, m := range r.msgs {
		if m.ConversationID == conversationID && m.Seq > sinceSeq {
			out = append(out, m)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *fakeMessageRepo) MarkReadByRecipient(_ context.Context, conversationID, recipientID string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, m := range r.msgs {
		if m.ConversationID == conversationID && m.SenderID != recipientID && m.ReadAt == nil {
			t := now
			m.ReadAt = &t
		}
	}
	return nil
}

func (r *fakeMessageRepo) CountUnread(_ context.Context, conversationID, recipientID string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, m := range r.msgs {
		if m.ConversationID == conversationID && m.SenderID != recipientID && m.ReadAt == nil {
			count++
		}
	}
	return count, nil
}

type fakeNotificationRepo struct {
	mu    sync.Mutex
	items []*entity.Notification
}

func (r *fakeNotificationRepo) Save(_ context.Context, n *entity.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, n)
	return nil
}
func (r *fakeNotificationRepo) ListByUser(_ context.Context, userID string, limit int) ([]*entity.Notification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []*entity.Notification
	for _, n := range r.items {
		if n.UserID == userID {
			out = append(out, n)
		}
	}
	return out, nil
}
func (r *fakeNotificationRepo) MarkRead(_ context.Context, id, userID string, now time.Time) error {
	return nil
}
func (r *fakeNotificationRepo) CountUnread(_ context.Context, userID string) (int, error) {
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

// ─── GetOrCreateConversationUseCase ───────────────────────────────────────

func TestGetOrCreateConversation_CreatesThenReuses(t *testing.T) {
	trips := &fakeTripReader{trips: map[string]app.TripSnapshot{
		"trip1": {TripID: "trip1", RiderID: "rider1", DriverID: "driver1", Status: "in_progress", TripType: "ride"},
	}}
	convRepo := newFakeConversationRepo()
	uc := app.NewGetOrCreateConversationUseCase(convRepo, trips)

	c1, err := uc.Execute(context.Background(), "trip1", "rider1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c1.Status != entity.ConversationOpen {
		t.Errorf("status = %v, want open", c1.Status)
	}

	c2, err := uc.Execute(context.Background(), "trip1", "driver1")
	if err != nil {
		t.Fatalf("unexpected error on reuse: %v", err)
	}
	if c2.ID != c1.ID {
		t.Errorf("expected same conversation reused, got %s vs %s", c2.ID, c1.ID)
	}
}

func TestGetOrCreateConversation_RejectsNonParticipant(t *testing.T) {
	trips := &fakeTripReader{trips: map[string]app.TripSnapshot{
		"trip1": {TripID: "trip1", RiderID: "rider1", DriverID: "driver1", Status: "in_progress", TripType: "ride"},
	}}
	uc := app.NewGetOrCreateConversationUseCase(newFakeConversationRepo(), trips)

	_, err := uc.Execute(context.Background(), "trip1", "stranger")
	if domainerrors.GetCode(err) != domainerrors.CodePermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", err)
	}
}

func TestGetOrCreateConversation_RejectsNoDriverYet(t *testing.T) {
	trips := &fakeTripReader{trips: map[string]app.TripSnapshot{
		"trip1": {TripID: "trip1", RiderID: "rider1", DriverID: "", Status: "searching", TripType: "ride"},
	}}
	uc := app.NewGetOrCreateConversationUseCase(newFakeConversationRepo(), trips)

	_, err := uc.Execute(context.Background(), "trip1", "rider1")
	if domainerrors.GetCode(err) != domainerrors.CodePreconditionFailed {
		t.Fatalf("expected PreconditionFailed, got %v", err)
	}
}

func TestGetOrCreateConversation_LazyClosesOnTerminalTripStatus(t *testing.T) {
	trips := &fakeTripReader{trips: map[string]app.TripSnapshot{
		"trip1": {TripID: "trip1", RiderID: "rider1", DriverID: "driver1", Status: "in_progress", TripType: "ride"},
	}}
	convRepo := newFakeConversationRepo()
	uc := app.NewGetOrCreateConversationUseCase(convRepo, trips)

	c1, err := uc.Execute(context.Background(), "trip1", "rider1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c1.IsOpen() {
		t.Fatal("expected conversation open right after creation")
	}

	trips.trips["trip1"] = app.TripSnapshot{TripID: "trip1", RiderID: "rider1", DriverID: "driver1", Status: "completed", TripType: "ride"}
	c2, err := uc.Execute(context.Background(), "trip1", "rider1")
	if err != nil {
		t.Fatalf("unexpected error on re-access: %v", err)
	}
	if c2.IsOpen() {
		t.Error("expected conversation to be lazily closed once trip is completed")
	}
}

// ─── SendMessageUseCase ────────────────────────────────────────────────────

func setupConversation(t *testing.T, tripType string) (*fakeConversationRepo, *entity.Conversation) {
	t.Helper()
	convRepo := newFakeConversationRepo()
	conv := entity.NewConversation("conv1", "trip1", "rider1", "driver1", tripType, time.Now().UTC())
	if err := convRepo.Save(context.Background(), conv); err != nil {
		t.Fatalf("setup: %v", err)
	}
	return convRepo, conv
}

func TestSendMessage_HappyPathNotifiesOtherParty(t *testing.T) {
	convRepo, _ := setupConversation(t, "ride")
	msgRepo := newFakeMessageRepo()
	notifRepo := &fakeNotificationRepo{}
	notify := app.NewCreateNotificationUseCase(notifRepo, nil)
	uc := app.NewSendMessageUseCase(convRepo, msgRepo, notify, app.NewBroadcaster())

	msg, err := uc.Execute(context.Background(), app.SendMessageInput{
		ConversationID: "conv1", SenderID: "rider1", Text: "Xin chào",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Body != "Xin chào" || msg.SenderRole != entity.SenderRider {
		t.Errorf("unexpected message: %+v", msg)
	}
	unread, _ := notifRepo.CountUnread(context.Background(), "driver1")
	if unread != 1 {
		t.Errorf("expected driver1 to get 1 notification, got %d", unread)
	}
}

func TestSendMessage_QuickReplyResolvesServerSide(t *testing.T) {
	convRepo, _ := setupConversation(t, "delivery")
	msgRepo := newFakeMessageRepo()
	uc := app.NewSendMessageUseCase(convRepo, msgRepo, nil, nil)

	msg, err := uc.Execute(context.Background(), app.SendMessageInput{
		ConversationID: "conv1", SenderID: "driver1", QuickReplyKey: "picked_up",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Body != entity.QuickReplyText["picked_up"] {
		t.Errorf("body = %q, want canonical quick-reply text", msg.Body)
	}
}

func TestSendMessage_RejectsDeliveryOnlyQuickReplyOnRideConversation(t *testing.T) {
	convRepo, _ := setupConversation(t, "ride")
	msgRepo := newFakeMessageRepo()
	uc := app.NewSendMessageUseCase(convRepo, msgRepo, nil, nil)

	_, err := uc.Execute(context.Background(), app.SendMessageInput{
		ConversationID: "conv1", SenderID: "driver1", QuickReplyKey: "picked_up",
	})
	if domainerrors.GetCode(err) != domainerrors.CodeInvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestSendMessage_RejectsNonParticipant(t *testing.T) {
	convRepo, _ := setupConversation(t, "ride")
	msgRepo := newFakeMessageRepo()
	uc := app.NewSendMessageUseCase(convRepo, msgRepo, nil, nil)

	_, err := uc.Execute(context.Background(), app.SendMessageInput{
		ConversationID: "conv1", SenderID: "stranger", Text: "hi",
	})
	if domainerrors.GetCode(err) != domainerrors.CodePermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", err)
	}
}

func TestSendMessage_RejectsClosedConversation(t *testing.T) {
	convRepo, conv := setupConversation(t, "ride")
	conv.Close(time.Now().UTC())
	if err := convRepo.Update(context.Background(), conv); err != nil {
		t.Fatalf("setup: %v", err)
	}
	msgRepo := newFakeMessageRepo()
	uc := app.NewSendMessageUseCase(convRepo, msgRepo, nil, nil)

	_, err := uc.Execute(context.Background(), app.SendMessageInput{
		ConversationID: "conv1", SenderID: "rider1", Text: "hi",
	})
	if domainerrors.GetCode(err) != domainerrors.CodePreconditionFailed {
		t.Fatalf("expected PreconditionFailed, got %v", err)
	}
}

// ─── PollMessagesUseCase ───────────────────────────────────────────────────

func TestPollMessages_ReturnsImmediatelyWhenMessagesAlreadyAvailable(t *testing.T) {
	convRepo, _ := setupConversation(t, "ride")
	msgRepo := newFakeMessageRepo()
	m, _ := entity.NewMessage("m1", "conv1", "driver1", entity.SenderDriver, "hello", "", time.Now().UTC())
	_ = msgRepo.Save(context.Background(), m)

	uc := app.NewPollMessagesUseCase(convRepo, msgRepo, app.NewBroadcaster())
	start := time.Now()
	msgs, err := uc.Execute(context.Background(), "conv1", "rider1", 0, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("expected immediate return, took %v", elapsed)
	}
}

func TestPollMessages_WakesUpOnPublish(t *testing.T) {
	convRepo, _ := setupConversation(t, "ride")
	msgRepo := newFakeMessageRepo()
	broadcaster := app.NewBroadcaster()
	uc := app.NewPollMessagesUseCase(convRepo, msgRepo, broadcaster)

	go func() {
		time.Sleep(50 * time.Millisecond)
		m, _ := entity.NewMessage("m1", "conv1", "driver1", entity.SenderDriver, "hello", "", time.Now().UTC())
		_ = msgRepo.Save(context.Background(), m)
		broadcaster.Publish("conv1")
	}()

	start := time.Now()
	msgs, err := uc.Execute(context.Background(), "conv1", "rider1", 0, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message after wake-up, got %d", len(msgs))
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("expected wake-up well under the 5s timeout, took %v", elapsed)
	}
}

func TestPollMessages_TimesOutWithEmptyResult(t *testing.T) {
	convRepo, _ := setupConversation(t, "ride")
	msgRepo := newFakeMessageRepo()
	uc := app.NewPollMessagesUseCase(convRepo, msgRepo, app.NewBroadcaster())

	msgs, err := uc.Execute(context.Background(), "conv1", "rider1", 0, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("expected empty result on timeout, got %d messages", len(msgs))
	}
}

// ─── MarkReadUseCase ────────────────────────────────────────────────────────

func TestMarkRead_MarksOnlyOtherPartysMessages(t *testing.T) {
	convRepo, _ := setupConversation(t, "ride")
	msgRepo := newFakeMessageRepo()
	m1, _ := entity.NewMessage("m1", "conv1", "driver1", entity.SenderDriver, "from driver", "", time.Now().UTC())
	m2, _ := entity.NewMessage("m2", "conv1", "rider1", entity.SenderRider, "from rider", "", time.Now().UTC())
	_ = msgRepo.Save(context.Background(), m1)
	_ = msgRepo.Save(context.Background(), m2)

	uc := app.NewMarkReadUseCase(convRepo, msgRepo)
	if err := uc.Execute(context.Background(), "conv1", "rider1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	unread, _ := msgRepo.CountUnread(context.Background(), "conv1", "rider1")
	if unread != 0 {
		t.Errorf("expected rider1's unread count 0 after marking read, got %d", unread)
	}
	// rider1's own message (m2) must NOT be marked read by rider1 reading —
	// it was never unread from rider1's perspective; confirm it stays
	// unread from driver1's perspective (driver1 hasn't read it).
	driverUnread, _ := msgRepo.CountUnread(context.Background(), "conv1", "driver1")
	if driverUnread != 1 {
		t.Errorf("expected driver1 to still have 1 unread message, got %d", driverUnread)
	}
}
