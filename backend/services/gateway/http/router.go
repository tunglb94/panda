package httpgateway

import (
	"net/http"

	"github.com/fairride/gateway/http/handlers"
	"github.com/fairride/gateway/http/middleware"
	"github.com/rs/zerolog"
)

// NewRouter constructs the gateway HTTP mux with all routes and middleware wired up.
// /health and /api/v1/auth/* are unauthenticated. All other /api/v1/* routes require
// a valid Bearer JWT. Every request is wrapped in the logging middleware.
func NewRouter(
	bh *handlers.BookingHandler,
	ah *handlers.AuthHandler,
	avh *handlers.AvailabilityHandler,
	lh *handlers.LocationHandler,
	dph *handlers.DriverProfileHandler,
	rh *handlers.RatingHandler,
	dh *handlers.DeliveryHandler,
	ch *handlers.ChatHandler,
	cah *handlers.CallHandler,
	nh *handlers.NotificationHandler,
	kh *handlers.KYCHandler,
	akh *handlers.AdminKYCHandler,
	wh *handlers.WalletHandler,
	awh *handlers.AdminWalletHandler,
	ph *handlers.PricingHandler,
	authMiddleware func(http.Handler) http.Handler,
	log zerolog.Logger,
) http.Handler {
	mux := http.NewServeMux()

	// Health — no auth required.
	mux.HandleFunc("GET /health", handleHealth)

	// Auth — no JWT required (issues the token).
	mux.HandleFunc("POST /api/v1/auth/login", ah.Login)
	mux.HandleFunc("POST /api/v1/auth/rider/login", ah.RiderLogin)
	mux.HandleFunc("POST /api/v1/auth/admin/login", ah.AdminLogin)
	mux.HandleFunc("POST /api/v1/auth/refresh", ah.Refresh)

	auth := authMiddleware
	requireAdmin := func(h http.Handler) http.Handler { return auth(middleware.RequireAdmin(h)) }

	// Driver availability — auth required.
	mux.Handle("POST /api/v1/driver/go-online", auth(http.HandlerFunc(avh.GoOnline)))
	mux.Handle("POST /api/v1/driver/go-offline", auth(http.HandlerFunc(avh.GoOffline)))
	mux.Handle("GET /api/v1/driver/availability", auth(http.HandlerFunc(avh.GetAvailability)))

	// Driver location — auth required.
	mux.Handle("POST /api/v1/driver/location", auth(http.HandlerFunc(lh.UpdateLocation)))
	mux.Handle("GET /api/v1/driver/{driverID}/location", auth(http.HandlerFunc(lh.GetLocation)))

	// Driver profile — auth required (rider reads assigned driver's profile).
	mux.Handle("GET /api/v1/drivers/{driverID}/profile", auth(http.HandlerFunc(dph.GetDriverProfile)))

	// Driver trip offer — auth required (driver polls this endpoint).
	mux.Handle("GET /api/v1/driver/current-offer", auth(http.HandlerFunc(bh.GetDriverOffer)))

	// Trip history — auth required.
	mux.Handle("GET /api/v1/rider/trips", auth(http.HandlerFunc(bh.ListRiderTrips)))
	mux.Handle("GET /api/v1/driver/trips", auth(http.HandlerFunc(bh.ListDriverTrips)))

	// Booking API — all routes require authentication.
	mux.Handle("POST /api/v1/rides/estimate-fare", auth(http.HandlerFunc(ph.EstimateFare)))
	mux.Handle("POST /api/v1/rides", auth(http.HandlerFunc(bh.BookRide)))
	mux.Handle("GET /api/v1/rides/{tripID}", auth(http.HandlerFunc(bh.GetBooking)))
	mux.Handle("POST /api/v1/rides/{tripID}/accept", auth(http.HandlerFunc(bh.AcceptDispatchOffer)))
	mux.Handle("POST /api/v1/rides/{tripID}/reject", auth(http.HandlerFunc(bh.RejectDispatchOffer)))
	mux.Handle("POST /api/v1/rides/{tripID}/arrive", auth(http.HandlerFunc(bh.ArriveAtPickup)))
	mux.Handle("POST /api/v1/rides/{tripID}/start", auth(http.HandlerFunc(bh.StartTrip)))
	mux.Handle("POST /api/v1/rides/{tripID}/finish", auth(http.HandlerFunc(bh.FinishTrip)))

	// Delivery lifecycle — driver-only actions, auth required. Trip service
	// owns these RPCs directly (Booking's proto has no delivery-lifecycle
	// equivalent); see DeliveryHandler.
	mux.Handle("POST /api/v1/rides/{tripID}/pickup-parcel", auth(http.HandlerFunc(dh.PickupParcel)))
	mux.Handle("POST /api/v1/rides/{tripID}/start-delivery", auth(http.HandlerFunc(dh.StartDelivery)))
	mux.Handle("POST /api/v1/rides/{tripID}/complete-delivery", auth(http.HandlerFunc(dh.CompleteDelivery)))
	mux.Handle("POST /api/v1/rides/{tripID}/pay", auth(http.HandlerFunc(bh.PayRide)))
	mux.Handle("POST /api/v1/rides/{tripID}/cancel", auth(http.HandlerFunc(bh.CancelRide)))
	mux.Handle("POST /api/v1/rides/{tripID}/rate", auth(http.HandlerFunc(rh.SubmitRating)))
	mux.Handle("GET /api/v1/rides/{tripID}/rating", auth(http.HandlerFunc(rh.GetRating)))

	// Communication Module — Phone Call, In-App Chat, Notification (auth required).
	mux.Handle("GET /api/v1/rides/{tripID}/contact", auth(http.HandlerFunc(cah.GetContact)))
	mux.Handle("POST /api/v1/rides/{tripID}/call", auth(http.HandlerFunc(cah.Call)))
	mux.Handle("GET /api/v1/rides/{tripID}/conversation", auth(http.HandlerFunc(ch.GetConversation)))
	mux.Handle("GET /api/v1/conversations/{id}/messages", auth(http.HandlerFunc(ch.ListOrPollMessages)))
	mux.Handle("POST /api/v1/conversations/{id}/messages", auth(http.HandlerFunc(ch.SendMessage)))
	mux.Handle("POST /api/v1/conversations/{id}/read", auth(http.HandlerFunc(ch.MarkConversationRead)))
	mux.Handle("GET /api/v1/notifications", auth(http.HandlerFunc(nh.ListNotifications)))
	mux.Handle("POST /api/v1/notifications/{id}/read", auth(http.HandlerFunc(nh.MarkRead)))

	// Driver KYC + Vehicle Verification — driver-facing (own record only, auth required).
	mux.Handle("POST /api/v1/driver/verification", auth(http.HandlerFunc(kh.SubmitDriverVerification)))
	mux.Handle("PUT /api/v1/driver/verification", auth(http.HandlerFunc(kh.UpdateDriverVerification)))
	mux.Handle("GET /api/v1/driver/verification", auth(http.HandlerFunc(kh.GetDriverVerification)))
	mux.Handle("POST /api/v1/vehicle/verification", auth(http.HandlerFunc(kh.SubmitVehicleVerification)))
	mux.Handle("PUT /api/v1/vehicle/verification", auth(http.HandlerFunc(kh.UpdateVehicleVerification)))
	mux.Handle("GET /api/v1/vehicle/verification", auth(http.HandlerFunc(kh.GetVehicleVerification)))
	mux.Handle("POST /api/v1/driver/verification/documents", auth(http.HandlerFunc(kh.UploadDocument)))
	mux.Handle("GET /api/v1/driver/verification/documents", auth(http.HandlerFunc(kh.ListDocuments)))
	mux.Handle("GET /api/v1/driver/verification/documents/{documentType}/versions", auth(http.HandlerFunc(kh.ListDocumentVersions)))

	// KYC review dashboard — admin-only (Phần 12).
	mux.Handle("GET /api/v1/admin/verifications/drivers", requireAdmin(http.HandlerFunc(akh.ListDriverVerifications)))
	mux.Handle("POST /api/v1/admin/verifications/drivers/{driverID}/approve", requireAdmin(http.HandlerFunc(akh.ApproveDriverVerification)))
	mux.Handle("POST /api/v1/admin/verifications/drivers/{driverID}/reject", requireAdmin(http.HandlerFunc(akh.RejectDriverVerification)))
	mux.Handle("GET /api/v1/admin/verifications/vehicles", requireAdmin(http.HandlerFunc(akh.ListVehicleVerifications)))
	mux.Handle("POST /api/v1/admin/verifications/vehicles/{driverID}/approve", requireAdmin(http.HandlerFunc(akh.ApproveVehicleVerification)))
	mux.Handle("POST /api/v1/admin/verifications/vehicles/{driverID}/reject", requireAdmin(http.HandlerFunc(akh.RejectVehicleVerification)))
	mux.Handle("GET /api/v1/admin/verifications/drivers/{driverID}/documents", requireAdmin(http.HandlerFunc(akh.ListDriverDocuments)))
	mux.Handle("GET /api/v1/admin/verifications/documents/{documentID}", requireAdmin(http.HandlerFunc(akh.GetDocument)))
	mux.Handle("GET /api/v1/admin/verifications/drivers/{driverID}/detail", requireAdmin(http.HandlerFunc(akh.GetDriverKYCDetail)))
	mux.Handle("GET /api/v1/admin/verifications/drivers/{driverID}/documents.zip", requireAdmin(http.HandlerFunc(akh.DownloadDriverDocumentsZip)))
	mux.Handle("GET /api/v1/admin/verifications/summary", requireAdmin(http.HandlerFunc(akh.GetKYCSummary)))

	// Driver Finance / Settlement Engine — driver-facing (own wallet only, auth required).
	mux.Handle("GET /api/v1/driver/wallet/summary", auth(http.HandlerFunc(wh.GetSummary)))
	mux.Handle("GET /api/v1/driver/wallet/statement", auth(http.HandlerFunc(wh.GetStatement)))
	mux.Handle("GET /api/v1/driver/wallet/transactions", auth(http.HandlerFunc(wh.ListTransactions)))
	mux.Handle("GET /api/v1/driver/wallet/bank-account", auth(http.HandlerFunc(wh.GetBankAccount)))
	mux.Handle("POST /api/v1/driver/wallet/bank-account", auth(http.HandlerFunc(wh.SetBankAccount)))
	mux.Handle("PUT /api/v1/driver/wallet/bank-account", auth(http.HandlerFunc(wh.SetBankAccount)))
	mux.Handle("POST /api/v1/driver/wallet/payouts", auth(http.HandlerFunc(wh.CreatePayoutRequest)))
	mux.Handle("GET /api/v1/driver/wallet/payouts", auth(http.HandlerFunc(wh.ListMyPayoutRequests)))

	// Driver Finance — admin-only (Phần 10). "Không cần UI. Chỉ API."
	mux.Handle("GET /api/v1/admin/settlements", requireAdmin(http.HandlerFunc(awh.ListSettlements)))
	mux.Handle("GET /api/v1/admin/settlements/outstanding", requireAdmin(http.HandlerFunc(awh.ListOutstandingDrivers)))
	mux.Handle("GET /api/v1/admin/settlements/{settlementID}", requireAdmin(http.HandlerFunc(awh.GetSettlementDetail)))
	mux.Handle("GET /api/v1/admin/payouts", requireAdmin(http.HandlerFunc(awh.ListPayoutRequests)))
	mux.Handle("POST /api/v1/admin/payouts/{payoutRequestID}/approve", requireAdmin(http.HandlerFunc(awh.ApprovePayoutRequest)))
	mux.Handle("POST /api/v1/admin/payouts/{payoutRequestID}/reject", requireAdmin(http.HandlerFunc(awh.RejectPayoutRequest)))
	mux.Handle("POST /api/v1/admin/payouts/{payoutRequestID}/paid", requireAdmin(http.HandlerFunc(awh.MarkPayoutPaid)))
	mux.Handle("POST /api/v1/admin/wallet/adjustments", requireAdmin(http.HandlerFunc(awh.ManualAdjustment)))

	return middleware.CORS(middleware.Logging(log)(mux))
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
