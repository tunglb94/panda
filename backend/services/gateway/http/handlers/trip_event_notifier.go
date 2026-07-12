package handlers

import (
	"context"

	notificationapp "github.com/fairride/notification/app"
	notificationentity "github.com/fairride/notification/domain/entity"
)

// TripEventNotifier is Part 3's "additive side-effect at the gateway layer"
// — after an existing Booking/Delivery use case already succeeded, this
// fires one best-effort in-app Notification to the affected rider/driver.
// It never runs before or instead of the original action, only after, and
// its own failure is always swallowed (see Notify) — a notification is a
// courtesy, never a condition for the underlying trip/delivery action.
type TripEventNotifier struct {
	trips  TripStatusClient
	create *notificationapp.CreateNotificationUseCase
}

func NewTripEventNotifier(trips TripStatusClient, create *notificationapp.CreateNotificationUseCase) *TripEventNotifier {
	return &TripEventNotifier{trips: trips, create: create}
}

func (n *TripEventNotifier) configured() bool {
	return n != nil && n.trips != nil && n.create != nil
}

// Notify looks up the trip fresh and fires the notification for event.
// Best-effort only: any lookup/save failure is silently dropped, exactly
// like this file's sibling best-effort enrichment (enrichTripDetails).
func (n *TripEventNotifier) Notify(ctx context.Context, tripID, event string) {
	if !n.configured() {
		return
	}
	trip, err := (tripReaderAdapter{client: n.trips}).GetTrip(ctx, tripID)
	if err != nil {
		return
	}
	recipient, category, title, body := eventNotificationContent(trip, event)
	if recipient == "" {
		return
	}
	_, _ = n.create.Execute(ctx, notificationapp.CreateNotificationInput{
		UserID:   recipient,
		Category: category,
		Title:    title,
		Body:     body,
		TripID:   tripID,
	})
}

// NotifyFromSnapshot is the same as Notify but skips the extra GetTrip
// round-trip when the caller already has trip data on hand (Delivery's
// three lifecycle RPCs return the full TripProto directly).
func (n *TripEventNotifier) NotifyFromSnapshot(ctx context.Context, trip notificationapp.TripSnapshot, event string) {
	if !n.configured() {
		return
	}
	recipient, category, title, body := eventNotificationContent(trip, event)
	if recipient == "" {
		return
	}
	_, _ = n.create.Execute(ctx, notificationapp.CreateNotificationInput{
		UserID:   recipient,
		Category: category,
		Title:    title,
		Body:     body,
		TripID:   trip.TripID,
	})
}

func eventNotificationContent(trip notificationapp.TripSnapshot, event string) (recipient string, category notificationentity.Category, title, body string) {
	category = notificationentity.CategoryTrip
	if trip.TripType == "delivery" {
		category = notificationentity.CategoryDelivery
	}
	switch event {
	case "accepted":
		return trip.RiderID, category, "Tài xế đã nhận chuyến", "Tài xế đang trên đường đến điểm đón."
	case "arrived":
		return trip.RiderID, category, "Tài xế đã đến điểm đón", "Vui lòng ra điểm đón."
	case "started":
		return trip.RiderID, category, "Chuyến đi đã bắt đầu", ""
	case "finished":
		return trip.RiderID, category, "Chuyến đi đã hoàn tất", "Cảm ơn bạn đã sử dụng Panda."
	case "cancelled":
		return trip.DriverID, category, "Chuyến đi đã bị huỷ", "Người dùng đã huỷ chuyến."
	case "pickup_parcel":
		return trip.RiderID, category, "Tài xế đã lấy hàng", ""
	case "start_delivery":
		return trip.RiderID, category, "Đơn hàng đang được giao", ""
	case "complete_delivery":
		return trip.RiderID, category, "Đơn hàng đã được giao", ""
	default:
		return "", category, "", ""
	}
}
