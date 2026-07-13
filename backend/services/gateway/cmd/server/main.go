package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/fairride/booking/grpc/bookingpb"
	"github.com/fairride/dispatch/grpc/dispatchpb"
	driverapp "github.com/fairride/driver/app"
	"github.com/fairride/driver/grpc/driverpb"
	driverlocalstore "github.com/fairride/driver/infrastructure/localstore"
	driverpostgres "github.com/fairride/driver/infrastructure/postgres"
	httpgateway "github.com/fairride/gateway/http"
	"github.com/fairride/gateway/http/handlers"
	"github.com/fairride/gateway/http/middleware"
	"github.com/fairride/identity/infrastructure/jwt"
	identitypostgres "github.com/fairride/identity/infrastructure/postgres"
	notificationapp "github.com/fairride/notification/app"
	notificationpostgres "github.com/fairride/notification/infrastructure/postgres"
	"github.com/fairride/pricing/grpc/pricingpb"
	reviewapp "github.com/fairride/review/app"
	"github.com/fairride/review/grpc/reviewpb"
	reviewpostgres "github.com/fairride/review/infrastructure/postgres"
	sharedconfig "github.com/fairride/shared/config"
	"github.com/fairride/shared/database"
	"github.com/fairride/shared/logger"
	"github.com/fairride/trip/grpc/trippb"
	walletapp "github.com/fairride/wallet/app"
	walletpostgres "github.com/fairride/wallet/infrastructure/postgres"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := sharedconfig.Load("gateway")
	log := logger.FromConfig(cfg.LogLevel, "gateway", cfg.Environment)

	jwtCfg := jwt.Config{
		AccessSecret:    mustEnv("JWT_ACCESS_SECRET"),
		RefreshSecret:   mustEnv("JWT_REFRESH_SECRET"),
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}
	tokenSvc, err := jwt.NewTokenService(jwtCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid JWT config")
	}

	// Booking service.
	bookingAddr := envOrDefault("BOOKING_ADDR", cfg.GRPC.Addr)
	bookingConn, err := grpc.NewClient(bookingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().Err(err).Str("addr", bookingAddr).Msg("failed to connect to booking service")
	}
	defer bookingConn.Close()

	// Trip service: delivery-lifecycle actions (pickup-parcel/start-delivery/
	// complete-delivery) and delivery-status enrichment on GetBooking live
	// only on Trip's own gRPC surface — Booking's proto was never extended
	// with equivalent RPCs/fields. If TRIP_ADDR is unset, delivery status
	// enrichment is silently skipped and the delivery action routes return
	// 503, but Ride booking/polling is entirely unaffected.
	var tripClient trippb.TripServiceClient
	if tripAddr := os.Getenv("TRIP_ADDR"); tripAddr != "" {
		tripConn, connErr := grpc.NewClient(tripAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if connErr != nil {
			log.Warn().Err(connErr).Str("addr", tripAddr).Msg("gateway: trip service connection failed — delivery actions will return 503")
		} else {
			defer tripConn.Close()
			tripClient = trippb.NewTripServiceClient(tripConn)
		}
	}
	dh := handlers.NewDeliveryHandler(tripClient)

	bookingClient := bookingpb.NewBookingServiceClient(bookingConn)
	bh := handlers.NewBookingHandler(bookingClient, tripClient)

	// Shared Postgres pool: Auth (identity+driver reads) and the entire
	// Communication Module (chat/call/notification) + Review's new
	// average-rating read all run in-process against this one pool — no
	// separate gRPC hop for any of them (see notification module's report,
	// "Kien truc": this environment has no protoc/buf toolchain to generate
	// a brand-new proto package, so these are imported as Go libraries
	// directly, mirroring the existing identity-service precedent, which
	// has never had a gRPC surface either). If DB_URL is unset or the
	// connection fails, every dependent handler gracefully returns 503.
	var pool *database.Pool
	if dbURL := os.Getenv("DB_URL"); dbURL != "" {
		p, dbErr := database.Connect(context.Background(), database.Config{
			URL:      dbURL,
			MaxConns: 5,
			MinConns: 1,
		})
		if dbErr != nil {
			log.Warn().Err(dbErr).Msg("gateway: DB connection failed — auth/chat/call/notification will return 503")
		} else {
			pool = p
		}
	}

	var ah *handlers.AuthHandler
	if pool != nil {
		ah = handlers.NewAuthHandler(
			identitypostgres.NewUserRepository(pool),
			driverpostgres.NewDriverRepository(pool),
			tokenSvc,
		)
	}
	if ah == nil {
		ah = handlers.NewAuthHandler(nil, nil, tokenSvc)
	}

	// Communication Module — Phone Call / In-App Chat / Notification /
	// Contact Card. Needs both `pool` (its own tables) and `tripClient`
	// (participant/status authorization) — degrades gracefully if either is absent.
	var ch *handlers.ChatHandler
	var cah *handlers.CallHandler
	var nh *handlers.NotificationHandler
	var tripNotifier *handlers.TripEventNotifier
	if pool != nil && tripClient != nil {
		convRepo := notificationpostgres.NewConversationRepository(pool)
		msgRepo := notificationpostgres.NewMessageRepository(pool)
		notifRepo := notificationpostgres.NewNotificationRepository(pool)
		callRepo := notificationpostgres.NewCallSessionRepository(pool)
		broadcaster := notificationapp.NewBroadcaster()
		pushSender := notificationapp.NewNoopPushSender(func(userID, title, body string) {
			log.Debug().Str("user_id", userID).Str("title", title).Msg("push: no FCM provider configured — in-app notification only")
		})
		createNotify := notificationapp.NewCreateNotificationUseCase(notifRepo, pushSender)
		tripReader := handlers.NewTripReader(tripClient)

		ch = handlers.NewChatHandler(
			notificationapp.NewGetOrCreateConversationUseCase(convRepo, tripReader),
			notificationapp.NewSendMessageUseCase(convRepo, msgRepo, createNotify, broadcaster),
			notificationapp.NewListMessagesUseCase(convRepo, msgRepo),
			notificationapp.NewPollMessagesUseCase(convRepo, msgRepo, broadcaster),
			notificationapp.NewMarkReadUseCase(convRepo, msgRepo),
			msgRepo,
		)

		avgRating := reviewapp.NewGetAverageRatingUseCase(reviewpostgres.NewRatingRepository(pool))
		recordCall := notificationapp.NewRecordCallUseCase(callRepo, createNotify)
		cah = handlers.NewCallHandler(
			tripClient, identitypostgres.NewUserRepository(pool), driverpostgres.NewDriverRepository(pool), avgRating, recordCall,
			driverapp.NewGetDriverVerificationUseCase(driverpostgres.NewDriverVerificationRepository(pool)),
			driverapp.NewGetVehicleVerificationUseCase(driverpostgres.NewVehicleVerificationRepository(pool)),
			bookingClient,
		)

		nh = handlers.NewNotificationHandler(
			notificationapp.NewListNotificationsUseCase(notifRepo),
			notificationapp.NewMarkNotificationReadUseCase(notifRepo),
		)

		tripNotifier = handlers.NewTripEventNotifier(tripClient, createNotify)
	}
	if ch == nil {
		ch = handlers.NewChatHandler(nil, nil, nil, nil, nil, nil)
	}
	if cah == nil {
		cah = handlers.NewCallHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	}
	if nh == nil {
		nh = handlers.NewNotificationHandler(nil, nil)
	}
	bh.SetNotifier(tripNotifier)
	dh.SetNotifier(tripNotifier)

	// Driver KYC + Vehicle Verification (driver-facing + admin review).
	// Local disk document storage only (Phần 4 — no cloud upload); base
	// directory configurable via KYC_STORAGE_DIR, defaulting to a relative
	// ./data/kyc so a fresh dev checkout works with zero extra setup.
	var kh *handlers.KYCHandler
	var akh *handlers.AdminKYCHandler
	if pool != nil {
		driverVerificationRepo := driverpostgres.NewDriverVerificationRepository(pool)
		vehicleVerificationRepo := driverpostgres.NewVehicleVerificationRepository(pool)
		documentRepo := driverpostgres.NewKYCDocumentRepository(pool)
		licenseCapabilityRepo := driverpostgres.NewLicenseCapabilityRepository(pool)
		auditLogRepo := driverpostgres.NewAuditLogRepository(pool)
		storageDir := envOrDefault("KYC_STORAGE_DIR", "./data/kyc")
		documentStore := driverlocalstore.NewDocumentStore(storageDir)

		kh = handlers.NewKYCHandler(
			driverapp.NewSubmitDriverVerificationUseCase(driverVerificationRepo, documentRepo, auditLogRepo),
			driverapp.NewUpdateDriverVerificationUseCase(driverVerificationRepo, auditLogRepo),
			driverapp.NewGetDriverVerificationUseCase(driverVerificationRepo),
			driverapp.NewSubmitVehicleVerificationUseCase(vehicleVerificationRepo, documentRepo, licenseCapabilityRepo, auditLogRepo),
			driverapp.NewUpdateVehicleVerificationUseCase(vehicleVerificationRepo, licenseCapabilityRepo, auditLogRepo),
			driverapp.NewGetVehicleVerificationUseCase(vehicleVerificationRepo),
			driverapp.NewUploadKYCDocumentUseCase(documentRepo, driverVerificationRepo, vehicleVerificationRepo, auditLogRepo, documentStore),
			driverapp.NewListKYCDocumentsUseCase(documentRepo),
			driverapp.NewListKYCDocumentVersionsUseCase(documentRepo),
		)
		akh = handlers.NewAdminKYCHandler(
			driverapp.NewListDriverVerificationsUseCase(driverVerificationRepo),
			driverapp.NewReviewDriverVerificationUseCase(driverVerificationRepo, auditLogRepo),
			driverapp.NewGetDriverVerificationUseCase(driverVerificationRepo),
			driverapp.NewListVehicleVerificationsUseCase(vehicleVerificationRepo),
			driverapp.NewReviewVehicleVerificationUseCase(vehicleVerificationRepo, auditLogRepo),
			driverapp.NewGetVehicleVerificationUseCase(vehicleVerificationRepo),
			driverapp.NewListKYCDocumentsUseCase(documentRepo),
			driverapp.NewGetKYCDocumentUseCase(documentRepo),
			documentStore,
			driverapp.NewListAuditLogsUseCase(auditLogRepo),
			driverapp.NewGetKYCSummaryUseCase(driverVerificationRepo),
			driverpostgres.NewDriverRepository(pool),
			identitypostgres.NewUserRepository(pool),
		)
	}
	if kh == nil {
		kh = handlers.NewKYCHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	}
	if akh == nil {
		akh = handlers.NewAdminKYCHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	}

	// Driver Finance / Settlement Engine (Financial Core). Shares the same
	// pool as KYC — in-process, no gRPC surface (same reasoning as above).
	// SETTLEMENT_COMMISSION_RATE/PAYOUT_MIN_AMOUNT_CENTS are optional env
	// overrides for the Financial Core's own placeholder commission rate
	// and minimum payout floor (see wallet/app/settlement_engine.go and
	// payout_request.go's doc comments for why these are not read from
	// Pricing).
	var wh *handlers.WalletHandler
	var awh *handlers.AdminWalletHandler
	var settlementEngine *handlers.SettlementEngine
	if pool != nil {
		walletRepo := walletpostgres.NewWalletRepository(pool)
		ledgerRepo := walletpostgres.NewLedgerEntryRepository(pool)
		txRepo := walletpostgres.NewTransactionRepository(pool)
		settlementRepo := walletpostgres.NewSettlementRepository(pool)
		bankAccountRepo := walletpostgres.NewBankAccountRepository(pool)
		payoutRequestRepo := walletpostgres.NewPayoutRequestRepository(pool)
		walletAuditRepo := walletpostgres.NewAuditLogRepository(pool)

		getOrCreateWallet := walletapp.NewGetOrCreateWalletUseCase(walletRepo)
		walletSummary := walletapp.NewGetWalletSummaryUseCase(walletRepo, ledgerRepo, txRepo, payoutRequestRepo)
		commissionRate := envFloatOrDefault("SETTLEMENT_COMMISSION_RATE", walletapp.DefaultCommissionRate)
		minPayoutCents := envInt64OrDefault("PAYOUT_MIN_AMOUNT_CENTS", walletapp.DefaultMinimumPayoutCents)

		createSettlement := walletapp.NewCreateSettlementUseCase(settlementRepo, getOrCreateWallet, ledgerRepo, txRepo, walletAuditRepo, commissionRate)
		settlementEngine = handlers.NewSettlementEngine(bookingClient, tripClient, createSettlement)

		wh = handlers.NewWalletHandler(
			walletSummary,
			walletapp.NewGetStatementUseCase(settlementRepo, walletSummary),
			walletapp.NewListWalletTransactionsUseCase(walletRepo, ledgerRepo, txRepo),
			walletapp.NewGetBankAccountUseCase(bankAccountRepo),
			walletapp.NewSetBankAccountUseCase(bankAccountRepo, walletAuditRepo),
			walletapp.NewCreatePayoutRequestUseCase(payoutRequestRepo, bankAccountRepo, walletSummary, getOrCreateWallet, ledgerRepo, txRepo, walletAuditRepo, minPayoutCents),
			walletapp.NewListMyPayoutRequestsUseCase(payoutRequestRepo),
			driverapp.NewGetDriverVerificationUseCase(driverpostgres.NewDriverVerificationRepository(pool)),
		)
		awh = handlers.NewAdminWalletHandler(
			walletapp.NewListSettlementsUseCase(settlementRepo),
			walletapp.NewGetSettlementDetailUseCase(settlementRepo),
			walletapp.NewListOutstandingDriversUseCase(ledgerRepo),
			walletapp.NewListPayoutRequestsUseCase(payoutRequestRepo),
			walletapp.NewApprovePayoutRequestUseCase(payoutRequestRepo, walletAuditRepo),
			walletapp.NewRejectPayoutRequestUseCase(payoutRequestRepo, getOrCreateWallet, ledgerRepo, txRepo, walletAuditRepo),
			walletapp.NewMarkPayoutPaidUseCase(payoutRequestRepo, getOrCreateWallet, ledgerRepo, txRepo, walletAuditRepo),
			walletapp.NewManualAdjustmentUseCase(getOrCreateWallet, ledgerRepo, txRepo, walletAuditRepo),
		)
	}
	if wh == nil {
		wh = handlers.NewWalletHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	}
	if awh == nil {
		awh = handlers.NewAdminWalletHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	}
	bh.SetSettlementEngine(settlementEngine)

	// Driver availability: proxies to the driver gRPC service.
	// If DRIVER_ADDR is unset or the connection fails, availability returns 503 gracefully.
	var avh *handlers.AvailabilityHandler
	if driverAddr := os.Getenv("DRIVER_ADDR"); driverAddr != "" {
		driverConn, connErr := grpc.NewClient(driverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if connErr != nil {
			log.Warn().Err(connErr).Str("addr", driverAddr).Msg("gateway: driver service connection failed — availability will return 503")
		} else {
			defer driverConn.Close()
			avh = handlers.NewAvailabilityHandler(driverpb.NewDriverAvailabilityServiceClient(driverConn))
		}
	}
	if avh == nil {
		avh = handlers.NewAvailabilityHandler(nil)
	}

	// Dispatch service: driver location upload + rider location query.
	// If DISPATCH_ADDR is unset, location endpoints return 503 gracefully.
	var lh *handlers.LocationHandler
	if dispatchAddr := os.Getenv("DISPATCH_ADDR"); dispatchAddr != "" {
		dispatchConn, connErr := grpc.NewClient(dispatchAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if connErr != nil {
			log.Warn().Err(connErr).Str("addr", dispatchAddr).Msg("gateway: dispatch service connection failed — location will return 503")
		} else {
			defer dispatchConn.Close()
			lh = handlers.NewLocationHandler(dispatchpb.NewDispatchServiceClient(dispatchConn))
		}
	}
	if lh == nil {
		lh = handlers.NewLocationHandler(nil)
	}

	// Driver profile service: rider reads assigned driver's profile.
	// Shares the same connection as the availability handler when DRIVER_ADDR is set.
	var dph *handlers.DriverProfileHandler
	if driverAddr := os.Getenv("DRIVER_ADDR"); driverAddr != "" {
		driverProfileConn, connErr := grpc.NewClient(driverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if connErr != nil {
			log.Warn().Err(connErr).Str("addr", driverAddr).Msg("gateway: driver profile connection failed — profile will return 503")
		} else {
			defer driverProfileConn.Close()
			dph = handlers.NewDriverProfileHandler(driverpb.NewDriverProfileServiceClient(driverProfileConn))
		}
	}
	if dph == nil {
		dph = handlers.NewDriverProfileHandler(nil)
	}

	// Pricing service: pre-booking fare estimate (Rider's BookingForm sends
	// only pickup/destination/service_type — Backend is the fare Single
	// Source of Truth, no client-side fare math).
	// If PRICING_ADDR is unset, estimate-fare returns 503 gracefully.
	var ph *handlers.PricingHandler
	if pricingAddr := os.Getenv("PRICING_ADDR"); pricingAddr != "" {
		pricingConn, connErr := grpc.NewClient(pricingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if connErr != nil {
			log.Warn().Err(connErr).Str("addr", pricingAddr).Msg("gateway: pricing service connection failed — estimate-fare will return 503")
		} else {
			defer pricingConn.Close()
			ph = handlers.NewPricingHandler(pricingpb.NewPricingServiceClient(pricingConn))
		}
	}
	if ph == nil {
		ph = handlers.NewPricingHandler(nil)
	}

	// Review service: submit and fetch trip ratings.
	// If REVIEW_ADDR is unset, rating endpoints return 503 gracefully.
	var rh *handlers.RatingHandler
	if reviewAddr := os.Getenv("REVIEW_ADDR"); reviewAddr != "" {
		reviewConn, connErr := grpc.NewClient(reviewAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if connErr != nil {
			log.Warn().Err(connErr).Str("addr", reviewAddr).Msg("gateway: review service connection failed — ratings will return 503")
		} else {
			defer reviewConn.Close()
			rh = handlers.NewRatingHandler(reviewpb.NewReviewServiceClient(reviewConn))
		}
	}
	if rh == nil {
		rh = handlers.NewRatingHandler(nil)
	}

	authMW := middleware.Auth(tokenSvc)
	router := httpgateway.NewRouter(bh, ah, avh, lh, dph, rh, dh, ch, cah, nh, kh, akh, wh, awh, ph, authMW, log)

	addr := cfg.HTTP.Addr
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}
	log.Info().Str("addr", addr).Msg("gateway listening")
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal().Err(err).Msg("gateway exited with error")
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("required environment variable not set: " + key)
	}
	return v
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envFloatOrDefault(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func envInt64OrDefault(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}
