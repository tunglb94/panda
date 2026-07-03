# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

---

## [Unreleased]

---

## [0.1.0] — MVP Milestone 01 — 2026-07-03

### Added

#### Infrastructure & Shared (Phase 1 / 1.5)
- Go workspace (`backend/go.work`) with 16 service modules
- `backend/shared` package: config, logger, errors, database (pgx), Redis, Kafka, gRPC server
- 14 microservice skeletons — all compile clean
- Docker Compose for PostgreSQL, Redis, Kafka, kafka-ui
- `Makefile` with standard targets
- GitHub Actions CI workflow (build + test)
- Multi-stage `backend/Dockerfile`
- Architecture audit: resolved all Critical + High findings
  - gRPC reflection gated behind env flag
  - Real readiness probe (503 on not-ready)
  - DB pool default MaxConns 5 (was 25 × 14 = too many)
  - gRPC message size limits wired from config
  - Kafka `RequireAllAcks` flag for financial messages
  - `DispatchEngine` goroutine timeout auto-retry
  - Linter config (`.golangci.yml`)

#### Identity Service (Phase 2.1 – 2.5)
- Permission entity (23 named constants) + Role entity (6 system roles)
- User entity: `NewUser`, status lifecycle (`Activate` / `Suspend` / `Deactivate`)
- PostgreSQL repositories: `PermissionRepository`, `RoleRepository`, `UserRepository`
- JWT foundation: pure stdlib HS256, separate access/refresh secrets, JTI (`crypto/rand`), `RefreshToken.Family` for future rotation

#### User Profile Service (Phase 3)
- `services/user` microservice — gRPC transport, Clean Architecture
- `UserProfile` entity with gender, date-of-birth, optional email + avatar
- PostgreSQL `ProfileRepository` (upsert, nullable date_of_birth)
- `GetProfileUseCase`, `UpdateProfileUseCase`
- 10 gRPC handler tests

#### Driver Service (Phase 4 – 6)
- `services/driver` microservice — gRPC transport, Clean Architecture
- `DriverProfile` entity: `OnlineStatus` + `VerificationStatus` state machines
- `Vehicle` entity with year bounds (1900 … current+1)
- `AvailabilityState` — Redis heartbeat presence (TTL=5 min, `GEOADD`/`GEOSEARCH`)
- Use cases: get/update driver, CRUD vehicles, go-online/offline/heartbeat
- PostgreSQL `DriverRepository` + `VehicleRepository`
- Redis `AvailabilityRepository` (pipeline, `GEOSEARCH ASC` for nearest-first)
- 130 unit tests across entity / app / grpc layers

#### Trip Service (Phase 7)
- `services/trip` microservice — gRPC transport, Clean Architecture
- `Trip` aggregate: 7-status state machine (`pending` → `completed` / `cancelled`)
- `IsCancellable()` — prevents cancel of in-progress / completed trips
- `StartTrip`, `CompleteTrip` with `finalFareTotal` + `fareCurrency` storage
- PostgreSQL `TripRepository` with index on `(rider_id)`
- 57 unit tests

#### Dispatch Service (Phase 8)
- `services/dispatch` microservice — gRPC transport, Clean Architecture
- `DispatchJob` aggregate: nearest-available-driver algorithm
- `DispatchEngine` — background goroutine auto-retries expired offers (5 s tick)
- Redis GEO: `GEOADD` on location update, `GEOSEARCH ASC` for candidate ranking
- `TripUpdater` writes directly to `trips` table (MVP trade-off, documented)
- Partial index on `dispatch_jobs(offer_expires_at) WHERE status='searching'`
- 56 unit tests

#### Pricing Service (Phase 9)
- `services/pricing` — pure compute, no DB, no Redis
- `FareCalculator`: `BaseFare + distance×PerKm + time×PerMin`, floor at `MinimumFare`, `+ BookingFee`
- Default rates for Car / Motorcycle / Van (USD cents)
- Upfront pricing guarantee: Estimate = CalculateFinal formula (only `IsFinal` flag differs)
- 33 unit tests

#### Booking Service (Phase 10)
- `services/booking` — orchestration only; no DB, no Redis
- Full booking flow: `BookRide → AcceptDispatchOffer → RejectDispatchOffer → StartTrip → FinishTrip`
- `GetBookingDetails` with graceful dispatch-404 handling
- gRPC adapters wrapping Trip / Dispatch / Pricing clients
- `TestFullBookingFlow` integration test (all 5 steps in sequence)
- 30 unit tests

#### API Gateway (Phase 11)
- `services/gateway` — HTTP-only gateway, no gRPC server
- Go 1.22 `net/http` ServeMux with method+path patterns
- JWT auth middleware (Bearer extraction → `ValidateAccessToken` → claims in context)
- Request/response logging middleware (`zerolog`)
- gRPC→HTTP error code mapping (NotFound→404, InvalidArgument→400, etc.)
- REST endpoints: `POST /api/v1/rides`, `GET /api/v1/rides/{tripID}`, `/accept`, `/reject`, `/start`, `/finish`, `GET /health`
- 18 unit tests (5 middleware + 13 handler)

#### Rider Flutter App — Skeleton (Phase 12)
- `apps/rider` — Flutter 3.44.4 / Dart 3.12.2
- Material 3 theme, FAIRRIDE green `#1A8C4E` as primary
- `go_router` `StatefulShellRoute.indexedStack` — 3 bottom-nav tabs (Home / Booking / Profile)
- `BookingPage` — vehicle selector (Car / Moto / Van), fare breakdown, confirm bottom sheet
- `ProfilePage` — header + rating badge + settings tiles + sign-out

#### Rider Flutter App — Map Foundation (Phase 14)
- Full-screen Google Maps (`google_maps_flutter ^2.10.0`)
- GPS permission lifecycle: check → request → handle denied / permanently denied / GPS off
- `getCurrentPosition` with 10 s timeout; `_LocationStatus` enum state machine
- Blue dot + My Location button + compass + zoom controls
- Android: `ACCESS_FINE_LOCATION` + `ACCESS_COARSE_LOCATION` + Google Maps API key meta-data
- iOS: `NSLocationWhenInUseUsageDescription` + `GMSServices.provideAPIKey` in AppDelegate
- `MapPage` replaces `HomePage` on the Home tab

#### Rider Flutter App — Location Engine (Phase 15)
- `lib/core/location/location_engine.dart` — reusable, stream-based GPS service (no UI dependency)
- Three broadcast streams: `locationStream`, `gpsStatusStream`, `permissionStream`
- `start()` / `stop()` / `pause()` / `resume()` / `updateConfig()` / `dispose()` lifecycle
- Platform-specific settings: `AndroidSettings` (honours `updateIntervalMs`), `AppleSettings`
- Auto-restart position stream when GPS re-enables mid-run
- `PermissionDeniedException` caught from position stream → emits `LocationPermissionStatus.denied`

#### Rider Flutter App — Pickup & Destination Selection (Phase 17)
- Three-state selection UX: `pickupPending → destinationPending → confirmed`
- `_CenterPin` overlay — pin tip geometrically aligned to map centre via `Padding(bottom: 48)`
- `_SelectionPanel` bottom sheet with live coordinates (5 decimal places ≈ 1 m precision)
- Green marker = confirmed pickup, Red marker = confirmed destination
- Edit Pickup preserves destination point; Edit Destination preserves pickup point
- `_tripSelection` getter returns `TripSelection(pickup, destination)` for future booking phase
- `GoogleMap.padding: bottom 240` keeps native controls above panel

### Technical Debt / Known Items
- `flutter pub get` + `flutter analyze` not yet run on home machine (Flutter broken on work machine)
- Google Maps API key placeholder (`YOUR_GOOGLE_MAPS_API_KEY`) in AndroidManifest + AppDelegate
- Map Abstraction Layer (`MapProvider` interface) queued before Route Engine phase
- DOC-0003 (System Architecture) + DOC-0004 (Implementation Master Plan) still pending
- 16 ADRs pending formal write-up
- Integration tests (Postgres / Redis) require running infrastructure — skipped in CI

### Test Summary
- Backend unit tests: **460 passing** across 24 packages (no DB/Redis required)
- Flutter: pending `flutter analyze` on home machine

---

[0.1.0]: https://github.com/fairride/fairride-eos/releases/tag/v0.1.0
