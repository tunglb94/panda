package app

import "context"

// PushSender delivers a real device push (FCM/APNs) for a Notification.
// Called every time CreateNotificationUseCase creates a row, in addition to
// the in-app feed write, so wiring up a real implementation later needs no
// call-site changes — only NewNoopPushSender's replacement at composition
// root.
//
// No Firebase project exists for this environment (confirmed — no
// google-services.json, no firebase_messaging dependency in either Flutter
// app); NoopPushSender is the only implementation for now. See the module
// report's Known Gap.
type PushSender interface {
	Send(ctx context.Context, userID, title, body string) error
}

// NoopPushSender logs and returns nil — a real push provider isn't
// configured yet, but every call site behaves as if one might be at any time.
type NoopPushSender struct {
	Log func(userID, title, body string)
}

func NewNoopPushSender(log func(userID, title, body string)) *NoopPushSender {
	return &NoopPushSender{Log: log}
}

func (s *NoopPushSender) Send(_ context.Context, userID, title, body string) error {
	if s.Log != nil {
		s.Log(userID, title, body)
	}
	return nil
}
