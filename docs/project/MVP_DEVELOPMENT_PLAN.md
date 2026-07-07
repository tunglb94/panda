# FAIRRIDE — MVP Development Plan

**Status:** Living document — Single Source of Truth (SSOT) for MVP implementation planning
**Owner:** Engineering (CTO / Principal Engineer)
**Created:** 2026-07-06
**Last updated:** 2026-07-07 (v1.11 — Phase D-06 complete: Driver App Arrived at Pickup implemented with mock data — `TripOfferState` extended to 9 values (+`arrivedAtPickup`, reached when the mock route progress finishes, not via a hard timer), `RouteProgressModel`/`RouteProgressIndicator` now go all the way to 0%/"Arrived" instead of flooring at 20%, new `WaitingTimer` (self-ticking mm:ss, no package, `onMinutePassed` callback) and `WaitingFeeCard` (mock rule: free 5 min then 2.000đ/phút) and `PassengerActionPanel` (Passenger On Board/Contact Rider/Cancel Trip, all plain callbacks) widgets, new `DriverArrivalCard` composing `DriverStatusBanner`/`TripAddressRow`/`RouteStatTile`/`FareEstimateCard` (all reused, none duplicated), Arrival Preview (repository-free, fixed waiting durations); v1.10 Phase D-05 complete: Driver App Assigned & Navigation Ready implemented with mock data — `TripOfferState` extended to 8 values (+`navigatingToPickup`), `RouteProgressModel`/`TrafficLevel` (Normal/Slow/Heavy), `DriverTripOfferRepository.fetchRouteProgress()`, new `DriverNavigationCard`/`DriverStatusBanner`/`RouteProgressIndicator`/`RouteStatTile` widgets (the latter also de-duplicated into `FareEstimateCard`), `TripAssignedCard` gained Pickup ETA/Distance/status, Start Navigation now drives a real state transition into a self-ticking mock route-progress screen (100%→20%, floor — no `arrived` yet), Contact Rider/Cancel Trip placeholder callbacks, Navigation Preview (repository-free); v1.9 Phase D-04 complete: Driver App Accept Flow & Dispatch Session implemented with mock data — `TripOfferState` extended to 7 values, `DriverTripOfferRepository.acceptOffer()` + `DispatchAcceptResult`, `TripAssignedCard`/`DispatchStatusBanner`/`AcceptLoadingButton`, Retry flow, Dispatch Session Preview, race-condition guard on the countdown; v1.8 Phase D-03 complete: Driver App Incoming Trip module implemented with mock data — Trip Offer/Fare Estimate cards, 15s animated countdown with auto-expiry, Accept/Reject actions, all 4 trip-offer states independently previewable via a dev menu; v1.7 Phase D-02 complete: Driver App Home dashboard implemented with mock data — driver summary/stats, Online/Offline toggle (4 states), status card (4 messages), Quick Actions (placeholder nav), reusing a hand-mirrored `AsyncStateView`; v1.6 Phase D-01 complete: Driver App initial project scaffold created (`apps/driver`, separate Flutter project) — Material 3 shell, named routes, 5-tab bottom nav, Developer page; decided no shared package with `apps/rider` for now; v1.5 Phase R-04 complete: Rider App Ride History module implemented with mock data — Trip History (search + filters, grouped by day), Trip Detail, Receipt, added as new Roadmap stage R9 (not in the original plan); v1.4 Phase R-03 complete: Rider App Profile module implemented with mock data — Profile Screen, Settings (9 reusable entries), Notification Center (unread badge, Loading/Success/Empty/Error states); v1.3 Phase R-02 complete: Rider App Trip Lifecycle UI implemented with mock data, reusing the Booking UI; v1.2 Phase R-01 complete: Rider App Booking UI module implemented with mock data; toolchain verified (`flutter pub get`/`analyze`/`test` now run clean); v1.1 revised after team review: Identity reframed from a blanket blocker to an Integration-Testing blocker; added Section 9 Parallel Development Matrix; MockAuth strategy formalized)
**Scope:** Planning only. Contains no source code, no API contracts, no protobuf definitions.
**Basis:** Generated entirely from inspection of the repository at commit `94321a3` (branch `main`). Where the repository is silent or ambiguous, this document says so rather than inventing detail.

> This document does not replace DOC-0001 (Project Constitution), DOC-0001A (AI Development Governance), or DOC-0002 (Product Vision). It sits below them: DOC-0002 defines *what* FAIRRIDE is: this document defines *in what order the MVP gets built*, given what already exists in the repository today.

---

## 1. Overview

### 1.1 Project Objective

Build the FAIRRIDE MVP: a two-sided ride-hailing marketplace (Rider app, Driver app, backend platform) that lets a rider book a car/motorcycle/van ride, get matched to a nearby driver, complete a trip, and pay for it — operated through a basic admin portal.

### 1.2 MVP Scope (per DOC-0002 §6.19, unchanged by this document)

**In scope:** standard ride booking (Economy tier), real-time driver tracking, in-app payment (card + wallet), driver onboarding with KYC, driver earnings dashboard, admin portal (city ops, disputes, refunds), dispatch engine, dynamic pricing engine, in-app wallet, push/SMS notifications, baseline rule-based fraud detection, basic analytics.

**Out of scope for MVP:** multiple vehicle categories beyond Economy, ride pooling, scheduled rides, corporate billing, delivery/food/logistics, open API, ML-based fraud detection, multi-city operations, multi-currency wallet.

### 1.3 Definition of MVP

The MVP is complete when a rider can, end-to-end, on a real device:
1. Register/log in
2. Request a ride and see fare estimate
3. Get matched to an available driver in real time
4. Track the driver approaching
5. Complete the trip and pay
6. Rate the trip

...and a driver can, on a real device:
1. Register/log in, complete onboarding/KYC
2. Go online and receive a trip offer
3. Accept/reject, navigate to pickup, start and complete the trip
4. See earnings

...and an admin can view trips, resolve a basic dispute, and issue a refund.

**As of this document's writing, none of the three device-facing halves (rider end-to-end, driver app, admin portal) are complete.** See Section 2.

### 1.4 Success Criteria

- All flows in 1.3 work against the real backend (not stubs) on at least one physical/emulated device per app.
- Backend services in the critical path (Identity, Gateway, Booking, Trip, Dispatch, Pricing, Driver, User) pass their existing test suites in CI, not just locally.
- No component required for the end-to-end flow is left as a skeleton (see Section 2.1 for what still is).

---

## 2. Current Project Status

*(Inspected directly from the repository. "Completed" means real domain logic + tests exist, not just a compiling skeleton.)*

### 2.1 Backend

**Completed (domain logic + persistence + tests present):**
| Service | What exists |
|---|---|
| `shared` | config, logger, errors, Postgres pool, Redis client, Kafka producer/consumer, gRPC server+interceptors, service bootstrap, idempotency store |
| `identity` | Role/Permission/User entities, Postgres repositories, JWT (HS256) issue/validate. **No register/login use case, no gRPC/HTTP endpoint — nothing can call this today.** |
| `user` | UserProfile entity, gRPC service, Postgres persistence |
| `driver` | DriverProfile + Vehicle entities, gRPC service, Postgres persistence, Redis-based availability heartbeat (GEO) |
| `trip` | Trip aggregate (7-status state machine incl. Start/Complete), gRPC service, Postgres persistence |
| `dispatch` | Nearest-driver matching engine, Redis GEO, Postgres persistence, atomic transactions (H2), hardened background engine lifecycle (H3-H4) |
| `pricing` | Pure-compute fare calculator (car/motorcycle/van), gRPC service, no DB |
| `booking` | Orchestrates Trip + Dispatch + Pricing over gRPC, saga compensation for orphaned trips, idempotency keys (H3-H4) |
| `gateway` | HTTP REST façade over `booking` only, JWT auth middleware, gRPC↔HTTP error mapping |

**Skeleton only (compiles, serves `/health` + `/ready`, zero domain code):**
`geo`, `wallet`, `payment`, `promotion`, `notification`, `review`, `analytics`, `admin` — 8 of 16 services.

**Missing / not started:**
- Identity register/login/OTP use cases and any transport (gRPC or HTTP) for Identity
- Any wallet or payment logic (blocked on ADR-0014, MVP payment methods, still pending)
- Any notification delivery (push/SMS)
- Any review/rating persistence
- Any admin-facing read/write API
- Any analytics aggregation
- Database migration tooling (schemas today are created only by Go test helpers per service — there is no script that provisions a real running Postgres instance outside of tests)
- Kafka is fully wired at the infrastructure level but **used by zero services** — all inter-service communication today is synchronous gRPC

### 2.2 Rider App (`apps/rider`, Flutter)

**Estimated completion:** UI construction ~92% of MVP rider screens now exist (Map/Location, Booking, Trip lifecycle, Profile/Settings/Notifications, Ride History/Detail/Receipt — every Section 1.3 flow step except auth now has *some* screen, plus a post-trip review surface DOC-0002 expects). **Functional completion remains ~10–15%** — every screen still runs on mock data; nothing is wired to Identity, Gateway, Pricing, Booking, Dispatch, Wallet, Payment, User, or Notification. Treat the two numbers separately: a complete-looking screen is not a working feature until Roadmap R2/R4/R6/R8/R9 land.

**Completed:**
- App shell: `go_router` 3-tab bottom nav (Home / Booking / Profile), Material 3 theme
- Full-screen Google Maps on the Home tab with GPS permission lifecycle (granted/denied/permanently-denied/GPS-off)
- Reusable stream-based `LocationEngine` (start/stop/pause/resume, GPS + permission change handling)
- Pickup/destination selection UX on the map (center-pin, confirm sheet, edit pickup/destination), producing a `TripSelection` value object
- **Booking UI module (Phase R-01, 2026-07-06):** full booking configuration screen — Pickup Card, Destination Card, Vehicle Selector (car/moto/van), Fare Summary (animated total), Payment Method Card (mock picker), Promo Code Entry (mock validation), and Book Ride Button (loading/success animation) — composed into a shared `BookingFormBody` used both by the Booking tab (`BookingPage`) and a new `BookingBottomSheet` invoked from the Map's "Book this ride" CTA once pickup/destination are confirmed. **UI only — mock data throughout (fare estimate uses a straight-line distance calc + the Pricing service's default rate shape, no API call is made). Not yet wired to any backend.**
- **Trip lifecycle UI (Phase R-02, 2026-07-06):** all five rider-facing trip states — Searching Driver, Driver Assigned, Driver Arriving, Trip In Progress, Trip Completed — as independently previewable views (`features/trip/presentation/pages/`), composed by a `TripLifecyclePage` that animates between them on a `MockTripRepository` timer. Reuses the Booking UI's `PickupCard`/`DestinationCard`/`FareSummaryCard`. Pushed automatically after a mock "Book Ride" success. Also reachable via a "Trip UI Preview (dev)" entry (now living under Settings, see below). **UI only — mock driver, mock ETA, no live location, no real cancellation/contact/emergency action.**
- **Profile module (Phase R-03, 2026-07-06):** `ProfilePage` rewritten as a real screen fetching a mock `RiderProfile` (avatar initial, full name, phone, member level badge, rating, total completed trips) from `MockProfileRepository` with a genuine Loading→Success transition; a new `SettingsPage` lists all 9 required entries (Personal Information, Payment Methods, Notifications, Privacy, Security, Language, Help Center, About, Logout) grouped into reusable `SettingsSection`/`SettingsTile` components — every entry except Notifications/Logout shows an explicit "placeholder, not yet implemented" message; a new `NotificationCenterPage` fetches mock notifications from `MockNotificationRepository` with tap-to-read, a bell icon + `UnreadBadge` on the Profile app bar, and a dev "Preview state" menu that demonstrates all four required states (Loading/Success/Empty/Error) via a generic reusable `AsyncStateView<T>`. **UI only — mock data throughout, no User/Notification backend calls.**
- **Ride History module (Phase R-04, 2026-07-06):** a new "Trip History" entry on the Profile screen opens `TripHistoryPage` — mock trips grouped by day, with a search box and status (All/Completed/Cancelled) + date-range (All time/Today/This Week/This Month) filter chips, all client-side over `MockTripHistoryRepository`; tapping a trip opens `TripDetailPage` (route summary, driver, vehicle, timeline, fare breakdown incl. promo discount, payment method, distance/duration — reusing `PickupCard`/`DestinationCard`/`FareSummaryCard` from Booking and `DriverInfoCard` from Trip directly) with a "View Receipt" button to `ReceiptPage` (Trip ID, rider, driver, vehicle, date, payment, fare breakdown, mock 8% tax, total). Loading/Success/Empty/Error again via the reused `AsyncStateView<T>`, with a dev "Preview state" menu matching the Notification Center's convention. **UI only — mock data throughout, no Trip/Review backend calls.**
- `flutter pub get` / `flutter analyze` now run clean on a working Flutter 3.35.4 / Dart 3.9.2 install (see Section 12 R-01 entry) — the SDK constraint in `pubspec.yaml` was lowered from `^3.12.2` to `^3.9.2` to match the installed toolchain; `flutter test` passes (15 smoke tests as of R-04)

**Missing:**
- Any network/API client — the app makes **zero HTTP calls** to the Gateway today
- Authentication/login screens (and nothing to call, since Identity has no login endpoint yet)
- Real booking submission (connecting the now-complete Booking UI to a fare estimate call and `POST /api/v1/rides` through the Gateway)
- Real trip tracking (the Trip lifecycle UI exists but runs on a fixed mock timer, not real Dispatch/Trip status or driver location)
- Real payment processing behind the Payment Method Card (Wallet/Payment backend doesn't exist)
- Rating UI where the rider *submits* a rating (driver rating is only ever *displayed*, both live in the Trip lifecycle UI and historically in Ride History — nothing lets the rider actually rate a driver)
- Real profile data (wiring `MockProfileRepository` to the existing `user` service) and real push notification delivery (`NotificationCenterPage` is UI + mock data only)
- Real trip history data (wiring `MockTripHistoryRepository` to the existing `trip` service) — same for the fare/receipt numbers, which are computed locally, not fetched
- Real screens behind 7 of the 9 Settings entries (Personal Information, Payment Methods, Privacy, Security, Language, Help Center, About are all placeholders; only Notifications and Logout's dialog are functional)
- State management library (none chosen yet — all state is local `StatefulWidget`, including the new Booking, Trip, Profile, and History UIs)
- Google Maps API key is a placeholder in both `AndroidManifest.xml` and `AppDelegate.swift` — the map cannot function on a real device/emulator until replaced

### 2.3 Driver App

**Initial shell exists (Phase D-01, 2026-07-06).** `apps/driver` is a real, separate Flutter project (own `pubspec.yaml`, `com.fairride.driver` bundle ID) — not a shared package with `apps/rider`. It has:
- App entry (`DriverApp`) — Material 3, `MaterialApp.router`, named `go_router` routes, same architecture as `apps/rider`
- 5-tab bottom navigation shell (Home, Trips, Earnings, Notifications, Profile), each a simple placeholder page reusing a shared `PlaceholderTabContent` widget
- A Driver-specific `AppTheme` — same design system as Rider (typography scale, button/card/input shapes) with a distinct deep-orange accent (`#EF6C00`) instead of Rider's green, per this phase's branding requirement
- A "Developer" page (Development Utilities), reachable only from the Profile tab: app version (mock), build mode (real, via `kReleaseMode`/`kDebugMode`), Flutter version and environment (explicit placeholders)
- Feature folder structure (`features/<name>/{domain,data,presentation/{pages,widgets}}`) demonstrated fully in the Profile feature; the 4 placeholder tabs only have `presentation/pages` so far since there is no data to layer yet

**Home dashboard implemented (Phase D-02, 2026-07-06).** The Home tab is no longer a placeholder — `HomePage` now shows:
- `DriverSummaryHeader` (avatar initial, name, rating, vehicle — or an explicit "No vehicle assigned yet" message) and `HomeStatsRow` (today's completed trips, today's earnings, online duration), fetched from `DriverHomeRepository` via a hand-mirrored `AsyncStateView<T>` (same Loading/Success/Empty/Error pattern as `apps/rider`'s Profile module, R-03 — no shared package, so re-implemented in `shared/widgets/`)
- `AvailabilityToggle` — a large primary Online/Offline switch supporting all 4 required states (Offline/Going Online/Online/Going Offline) via a local mock transition delay (~1.2s online, ~0.9s offline), no real availability API call
- `HomeStatusCard` — Offline / Waiting for trips / Searching nearby / Busy (placeholder) messaging, driven by the toggle (auto-advances Waiting→Searching after ~3s once online) plus a dev "Simulate busy" override, since a real Busy state needs an assigned trip (Roadmap D4/D6, not built)
- `QuickActionsSection` — Earnings / Trip History / Support / Vehicle cards, **placeholder navigation only** (SnackBar message, no real screen)

**Incoming Trip module implemented (Phase D-03, 2026-07-07).** The Trips tab is no longer a placeholder — `TripsPage` now shows:
- `TripOfferCard` (rider info, pickup/destination, distance to pickup, surge indicator when applicable) + `FareEstimateCard` (estimated fare/distance/duration), fetched from `DriverTripOfferRepository` via the same hand-mirrored `AsyncStateView<T>` pattern (Loading/Success/Empty/Error)
- `CountdownIndicator` — animated 15-second circular countdown that automatically fires an "Expired" transition when it reaches zero, no backend timeout involved
- `TripActionButtons` (Accept/Reject) driving a local offer-lifecycle state machine (mock transitions only)
- All base offer states independently previewable via a dev menu (`TripOfferPreviewMenuPage` → `TripOfferStatePreviewPage`), same convention as the Home dashboard's and `apps/rider`'s dev preview menus

**Accept Flow & Dispatch Session implemented (Phase D-04, 2026-07-07).** Pressing Accept no longer jumps straight to a result — the offer-lifecycle state machine (renamed `TripOfferState`, now 7 values: `newOffer/accepting/assigned/rejected/expired/failed/timeout`) grew a dispatch-confirmation sub-flow:
- Accept → `accepting` (countdown removed from the tree immediately — belt-and-suspenders guarded in `TripsPage._handleExpired` against a stray `onExpired` firing mid-`AnimatedSwitcher`-crossfade), full-width disabled `AcceptLoadingButton` with an inline spinner replaces the Accept/Reject row entirely (so Reject is unavailable, not just disabled)
- `DriverTripOfferRepository.acceptOffer({delay, outcome})` — new method returning a `DispatchAcceptResult` (`DispatchAcceptStatus`: success/failed/timeout — a dedicated result type, not a `bool`) after a configurable mock delay (default 1200ms)
- Success → `assigned`: new `TripAssignedCard` (✓ Trip Assigned, pickup, destination, estimated fare, "Start Navigation" — `onNavigate` is a mock callback only, no real navigation)
- Failed/Timeout → new `DispatchStatusBanner` with the exact copy requested ("Unable to accept trip. Try again." / "Dispatch timeout. Please retry.") and an explicit **Retry** button back to a fresh `newOffer` — never a silent auto-revert
- New "Dispatch Session Preview" entry inside `TripOfferPreviewMenuPage`, listing Accepting/Assigned/Failed/Timeout for independent preview without touching the repository
- `AsyncStateView` itself was **not modified** — the Loading/Success/Empty/Error fetch machine and the `TripOfferState` offer-lifecycle machine remain deliberately separate, composed (not merged) in `TripsPage`

**Driver Assigned & Navigation Ready implemented (Phase D-05, 2026-07-07).** The Assigned screen and a new Navigating-to-Pickup screen extend the same offer-lifecycle machine (`TripOfferState` now 8 values, +`navigatingToPickup` — `arrived` does not exist yet):
- `TripAssignedCard` (kept, not duplicated) gained Pickup ETA / Distance to Pickup (new `RouteStatTile`, reused from `FareEstimateCard` too — de-duplicated its old private `_Stat`) and a "Current status: Ready to navigate" line
- "Start Navigation" is no longer a placeholder callback — it now drives `TripsPage`'s state machine `assigned → navigatingToPickup` and kicks off `DriverTripOfferRepository.fetchRouteProgress({traffic})`
- New `RouteProgressModel`/`TrafficLevel` (Normal/Slow/Heavy, an enum — not a `String`) domain model; `fetchRouteProgress()` returns a mock 100%-remaining snapshot after a configurable delay (default 600ms), traffic only slows the ETA, not the physical distance
- New `DriverNavigationCard` ("Driving to Pickup" via new reusable `DriverStatusBanner`, pickup address via the existing `TripAddressRow`, Distance Remaining/ETA via `RouteStatTile`, mock `RouteProgressIndicator` with a traffic badge, Contact Rider/Cancel Trip — both plain callbacks, no dialer, no popup)
- Once navigating, a local `_RouteProgressTicker` (inside `TripOfferView`, no repository calls) ticks the mock progress down every 2s — through Phase D-05 it stopped at a 20% floor; Phase D-06 (below) lets it run all the way to 0% — nested behind its own `AsyncStateView<RouteProgressModel>`, kept deliberately separate from the outer offer-fetch `AsyncStateView<TripOffer?>`
- New "Navigation Preview" (`NavigationPreviewPage`, reached from `TripOfferPreviewMenuPage`) steps through Assigned/100%/80%/60%/40%/20% via `RouteProgressModel.mock(...)` directly — no repository, no delay
- `TripOfferPreviewMenuPage`/`TripOfferStatePreviewPage`/`DispatchSessionPreviewPage` were **not modified** beyond the one new menu entry — `navigatingToPickup` isn't added to their flat state lists since it needs the richer step-through UI the dedicated Navigation Preview provides instead

**Arrived at Pickup implemented (Phase D-06, 2026-07-07).** The route-progress ticker now runs to completion instead of flooring at 20%, and reaching 0% drives the offer-lifecycle machine's next transition (`TripOfferState` now 9 values, +`arrivedAtPickup` — **not** a hard timer; the state changes exactly when the mock progress finishes):
- `RouteProgressModel`/`RouteProgressIndicator` floor changed from 20 to 0 — the indicator's trailing label reads "Arrived" instead of "0% remaining"
- `_RouteProgressTicker` gained an `onArrived` callback, fired exactly once when its `stepDown()` chain reaches 0, then stops scheduling — `TripsPage._handleArrived()` (guarded like every other transition) flips `navigatingToPickup → arrivedAtPickup`
- New `WaitingTimer` — self-ticking mm:ss counter (no package, a recursive `Future.delayed(1s)` chain identical in shape to `_RouteProgressTicker`), with an `onMinutePassed(int minute)` callback fired once per whole minute; `initialSeconds` lets the Arrival Preview seed a fixed starting point without waiting in real time
- New `WaitingFeeCard` — pure mock-rule display driven by `elapsedMinutes` (free for the first 5 minutes, then 2.000đ/phút after), no backend billing
- New `PassengerActionPanel` — Passenger On Board / Contact Rider / Cancel Trip, all plain callbacks (no dialer, no popup) — `passengerBoarding` is **not** a state yet, so "Passenger On Board" is a placeholder like the other two
- New `DriverArrivalCard` composes the "Arrived at Pickup" status via the **existing, unmodified** `DriverStatusBanner`, pickup address via `TripAddressRow`, passenger name via `RouteStatTile`, and estimated fare via `FareEstimateCard` (all reused, nothing duplicated), plus `WaitingTimer`/`WaitingFeeCard`/`PassengerActionPanel`
- New "Arrival Preview" (`ArrivalPreviewPage`) steps through Arrived/Waiting 00:00/03:00/08:00 by seeding `DriverArrivalCard.initialWaitingSeconds` directly — no repository call

**Still a near-total gap for actual driver functionality beyond the Home/Trips UI** — no auth, no KYC, no real availability/location API, no real dispatch matching (offers are a single fixed mock, not a live feed), no real map/GPS navigation, no real earnings data. D2, D6, D7 (Driver App Roadmap) remain not started; D3 (Go online/offline), D4 (Offer handling) and D5 (Navigation to pickup, now complete through Arrived-at-Pickup — Phases D-05 + D-06) have their UI built but not real wiring. `passengerBoarding` (picking up the passenger and starting the trip — Roadmap stage D6) is not built yet.

### 2.4 Admin Portal

**Does not exist.** No Next.js (or any web) project is present anywhere in the repository. ADR-0003 (admin web technology: Next.js) is still unratified.

### 2.5 Infrastructure

- `docker-compose.yml` provisions PostgreSQL 16, Redis 7, Kafka 3.7 (KRaft), Kafka UI — functional for local dev
- No database migration tool (golang-migrate/goose/atlas/etc.) is wired in anywhere; `infra/docker/init/postgres/` contains only a placeholder
- No proto build automation exists in the `Makefile` — the 8 `.proto` files under `backend/proto/` were compiled manually per-phase (protoc + protoc-gen-go + protoc-gen-go-grpc), and generated `*pb.go` output is committed directly

### 2.6 Testing

- Backend: ~460+ unit tests recorded as of the last full count (pre-H3/H4; H3-H4 added several more), spread across `identity`, `user`, `driver`, `trip`, `dispatch`, `pricing`, `booking`, `gateway`, and `shared`
- **CI gap:** the GitHub Actions `test` job runs `go test` only inside `backend/shared`. None of the per-service unit test suites (the ~460 tests above) currently execute in CI — they only run when a developer runs them locally
- **CI gap:** the `build` job matrix lists 14 services but is missing `booking` and `gateway` (added after the matrix was last updated) — those two services are not build-verified in CI at all
- Postgres/Redis integration tests exist per service (identity, user, driver, trip, dispatch) but require `DATABASE_URL`/`REDIS_ADDR` env vars; they skip gracefully without them, and CI does not set them for anything but the `shared` job
- Flutter (Rider): `test/widget_test.dart` now holds 15 smoke tests (Booking UI; Trip Preview menu + 2 Trip lifecycle states; Profile page, Settings list, 3 Notification Center states; Trip History list/filter/empty/error + Trip Detail + Receipt — added across R-01 through R-04) — still smoke-level only (render checks + a few tap interactions for filter chips and demo menus), no golden tests yet.
- Flutter (Driver): `apps/driver/test/widget_test.dart` holds 40 smoke tests (3 from D-01: tab shell + all 5 destinations render, tab switching works, Developer page opens with all 4 diagnostic fields; 5 from D-02: Home dashboard loads summary/stats, the toggle walks through all 4 availability states with the status card updating in sync, Empty/Error dev states, Quick Action placeholder; 8 from D-03: Trips tab loads an offer with countdown, Reject transition, the countdown auto-expiring, Empty/Error dev states, the offer-state preview menu + Dispatch Session Preview menu + one state preview page; 8 from D-04: Accept walks through Accepting→Assigned, the loading button is disabled while accepting, the countdown cannot fire after Accept (race-condition guard), Accept→Failed and Accept→Timeout via the dev "Accept outcome" menu, Retry returns to a fresh New Offer, the Assigned screen shows correct data, `onNavigate` fires exactly once per tap; 8 from D-05: Start Navigation transitions Assigned→Navigating, the Navigation screen renders initial distance/ETA/traffic, route progress ticks 100%→80% over simulated time, the traffic badge reflects the dev "Traffic" menu, Cancel Trip and Contact Rider are plain callbacks with no dialog, the preview menu lists "Navigation Preview", and `NavigationPreviewPage` steps through Assigned→100%→20% without touching the repository; 8 new from D-06: route progress completing transitions Navigating→Arrived at Pickup, `WaitingTimer` counts up in mm:ss and fires `onMinutePassed`, `WaitingTimer` stops ticking after being disposed, `WaitingFeeCard` is free for 5 minutes then charges per minute, `PassengerActionPanel`'s three buttons fire their own callbacks, `DriverArrivalCard` renders its status banner/address/passenger via the existing `DriverStatusBanner`/`TripAddressRow`/`RouteStatTile`, the preview menu lists "Arrival Preview", and `ArrivalPreviewPage` steps through Arrived/Waiting 00:00–08:00 without touching the repository). Interacting with the "Accept outcome"/"Traffic" dev menus in tests calls their `onSelected` callback directly rather than tapping through the popup UI — the offer's 15s countdown keeps an `AnimationController` running the whole time `newOffer` is showing, which makes `pumpAndSettle()` unsafe (it never settles) and a fixed-duration `pump()` a hit-test-timing gamble against the popup's own open/close animation. Both the route-progress ticker and the new `WaitingTimer` are bare `Future.delayed` chains (no package, no `AnimationController`), so `pumpAndSettle()` is unsafe around them too — for a different reason: it falsely reports "settled" without the next tick ever firing, since nothing schedules a frame while a `Future.delayed` is merely waiting. Any test that doesn't specifically want to drive one of them to completion instead unmounts the widget tree and pumps past the single already-scheduled tick, so it fires harmlessly (via each ticker's own `_stopped` dispose-guard) instead of leaving a "pending Timer" failure at test end.
- `flutter analyze` and `flutter test` run clean on this machine's Flutter 3.35.4 / Dart 3.9.2 install, for both Flutter projects independently (`apps/rider` and `apps/driver` are separate projects — no shared Dart package between them, see §2.3).

### 2.7 Documentation

- DOC-0001 (Constitution), DOC-0001A (AI Governance), DOC-0002 (Product Vision): all Draft, none approved
- DOC-0003 (System Architecture) and DOC-0004 (Implementation Master Plan): not generated, blocked on unresolved Open Questions (launch market, driver classification, cash collection) and 16 pending ADRs (0 approved)
- `CHANGELOG.md` exists at root and is reasonably current through the H3-H4 hardening commit
- `.ai/memory.md` is the de-facto accurate phase tracker; `.ai/current-phase.md` and `.ai/current-task.md` are **stale** (still describe Phase 1) and should not be trusted over `.ai/memory.md`

---

## 3. Dependency Graph

```
Identity (auth/roles/JWT)
   │   [Integration-Testing blocker only — see note below. NOT a blocker for UI construction.]
   ▼
Gateway (HTTP façade, JWT verification)
   │
   ▼
Booking (orchestrator)
   │
   ├──▶ Trip        (owns trip state machine + trips table)
   ├──▶ Dispatch     (nearest-driver matching, Redis GEO)
   └──▶ Pricing      (pure fare compute, no state)

Dispatch ──(reads driver location)──▶ own Redis GEO layer (independent of Driver's own presence system)
Driver   ──(own Redis GEO heartbeat, Postgres online_status)──▶ NOT currently linked to Dispatch's location layer

User (profile)     — standalone, no other service depends on it yet
Driver (profile/vehicle/KYC) — standalone; Dispatch only consumes driver location, not the Driver Postgres records
Wallet / Payment / Promotion / Notification / Review / Analytics / Admin / Geo — skeletons, nothing depends on them yet
```

**Revised framing (team review, v1.1):** Identity's missing register/login is a blocker for **real Integration Testing / End-to-End verification (Milestone M4)**, not a blocker for **UI development**. Rider UI and Driver UI can — and should — be built against a **MockAuth** layer (a stub that returns a fake-but-structurally-valid token/session) so both app tracks proceed in parallel with Identity's own real implementation. The one condition: MockAuth's interface must mirror the real Gateway auth contract exactly (same token shape, same header, same error cases), so swapping mock→real at Integration Testing time is a config change, not a rewrite. See Section 9 (Parallel Development Matrix) and Section 14 Rule 5.

**Can be developed independently right now (no blocking dependency):**
Wallet, Payment, Promotion, Notification, Review, Analytics, Admin, Geo (all skeleton services — starting any one does not block or get blocked by another), Admin Portal (new project), Driver App (new project, can start against already-complete Driver/Vehicle/Availability backend APIs), Rider App auth/booking/tracking UI (against MockAuth + mock data), Identity register/login (no dependency of its own — can be built in parallel with both UI tracks).

**Must wait on something else:**
- Rider App **real** booking submission (swapping MockAuth for real tokens, mock fare/dispatch responses for real ones) → waits on Identity register/login + Gateway route expansion
- Any real end-to-end trip test → waits on Identity register/login (Section 2.1) and on a Driver App existing (Section 2.3) to have a second party in the marketplace
- Admin Portal disputes/refunds → waits on Wallet + Payment domain logic existing (currently skeletons)
- Driver earnings dashboard → waits on Wallet domain logic existing

**Known architectural note (not a defect — an intentional MVP trade-off documented in `.ai/memory.md`):** Driver's own presence system (Postgres `online_status` + Redis heartbeat, Phase 6) and Dispatch's own driver-location system (Redis GEO, Phase 8) are two independent systems that happen to both track "is this driver available." They are not reconciled today. This is worth resolving before Driver App integration, not after.

---

## 4. MVP Feature Breakdown

### Backend
| Module | Purpose | Dependencies | Status |
|---|---|---|---|
| Identity | Auth, roles, JWT | none | Entities+persistence+JWT done; **register/login/API missing** |
| Gateway | REST façade for mobile apps | Identity (JWT), Booking | Done for Booking only; no Identity/User/Driver routes exposed |
| User | Rider profile | none | Done (gRPC+Postgres) |
| Driver | Driver profile, vehicle, availability | none | Done (gRPC+Postgres+Redis) |
| Trip | Trip state machine | none | Done |
| Dispatch | Driver matching | Trip (shared table write), Driver (location only) | Done, hardened |
| Pricing | Fare calculation | none | Done |
| Booking | Orchestration + saga | Trip, Dispatch, Pricing | Done, hardened |
| Wallet | Rider/driver balance | Payment (eventually) | Not started |
| Payment | Card/wallet payment processing | Wallet | Not started; blocked on ADR-0014 |
| Promotion | Discounts/incentives | Wallet | Not started |
| Notification | Push/SMS | none | Not started |
| Review | Ratings | Trip | Not started |
| Analytics | Reporting | all services (event consumers) | Not started |
| Admin | Admin backend API | Trip, Wallet, Payment, User, Driver | Not started |
| Geo | Standalone geocoding/mapping service | none | Not started; **unclear if still needed** — Driver and Dispatch already implement their own Redis GEO logic independently. Recommend clarifying this service's actual scope before building it. |

### Rider App
| Module | Purpose | Dependencies | Status |
|---|---|---|---|
| Shell/Nav/Theme | App skeleton | none | Done |
| Map + Location Engine | GPS, permissions, live position | none | Done |
| Pickup/Destination selection | Trip origin/destination UX | Map | Done |
| Booking UI (mock data) | Pickup/Destination cards, Vehicle Selector, Fare Summary, Payment Method Card, Promo Code Entry, Book Ride Button, Booking Bottom Sheet | Map (`TripSelection`) | **Done (Phase R-01)** |
| Trip lifecycle UI (mock data) | Searching Driver, Driver Assigned, Driver Arriving, Trip In Progress, Trip Completed views + `TripLifecyclePage` orchestrator + dev preview menu | Booking UI (`TripSelection`, `FareSummaryCard`, `PickupCard`/`DestinationCard`) | **Done (Phase R-02)** |
| Profile module (mock data) | Profile Screen (avatar/name/phone/member level/rating/trips), Settings (9 entries, reusable sections/tiles), Notification Center (unread badge, read/unread, Loading/Success/Empty/Error) | none | **Done (Phase R-03)** |
| Ride History module (mock data) | Trip History (search + status/date filters, grouped by day), Trip Detail (route/driver/vehicle/timeline/fare/payment/distance/duration), Receipt (Trip ID/rider/driver/vehicle/date/payment/fare/taxes/total) | Booking UI (`FareSummaryCard`, `PickupCard`/`DestinationCard`), Trip module (`DriverInfoCard`/`MockDriver`), Profile module (`AsyncStateView`, rider name) | **Done (Phase R-04)** |
| Auth screens | Login/register UI | Identity register/login (missing) | Not started |
| API client | HTTP layer to Gateway | Gateway | Not started |
| Booking submission (real wiring) | Replace mock fare/payment/promo data with real Pricing/Booking API calls via Gateway | API client, Pricing/Booking APIs via Gateway (exist) | Not started |
| Trip tracking (real wiring) | Replace `MockTripRepository`'s fixed timer with real Dispatch/Trip status polling + live driver location | Driver App location broadcast (missing), Dispatch status endpoint | Not started |
| Payment processing | Real charge behind the Payment Method Card | Payment/Wallet backend (missing) | Not started |
| Rating UI | Post-trip review (rider submits a rating — distinct from the driver-rating *display* already in the Trip lifecycle UI and Ride History) | Review backend (missing) | Not started |
| Profile/Settings/Notifications (real wiring) | Wire `MockProfileRepository`/`MockNotificationRepository` to the real User/Notification services; build out the 7 placeholder Settings screens | API client, User service (exists), Notification service (skeleton only) | Not started |
| Ride History (real wiring) | Wire `MockTripHistoryRepository` to the real Trip service; compute fare/receipt numbers server-side instead of locally | API client, Trip service (exists) | Not started |

### Driver App
| Module | Purpose | Dependencies | Status |
|---|---|---|---|
| App shell (D1, mock/no data) | Entry point, Material 3 + distinct-accent theme, named routes, 5-tab bottom nav, Developer page | none | **Done (Phase D-01)** |
| Home dashboard (D2, mock data) | Driver summary header, today's stats, Online/Offline toggle (4 states), status card (4 messages), Quick Actions (placeholder nav), `AsyncStateView` (Loading/Success/Empty/Error) | App shell (D-01) | **Done (Phase D-02)** — UI portion of Roadmap stage D3 |
| Incoming Trip module (D3, mock data) | `TripOfferCard`, `FareEstimateCard`, `CountdownIndicator` (15s), `TripActionButtons` (Accept/Reject), offer-state machine, `AsyncStateView` (Loading/Success/Empty/Error) | App shell (D-01), `AsyncStateView` (D-02) | **Done (Phase D-03)** — UI portion of Roadmap stage D4 |
| Accept Flow & Dispatch Session (D4, mock data) | `TripAssignedCard`, `DispatchStatusBanner`, `AcceptLoadingButton`; `TripOfferState` extended to 7 values (+accepting/assigned/failed/timeout); `DriverTripOfferRepository.acceptOffer()` + `DispatchAcceptResult`/`DispatchAcceptStatus`; Retry flow; Dispatch Session Preview | Incoming Trip module (D-03) | **Done (Phase D-04)** — completes the UI portion of Roadmap stage D4 |
| Assigned & Navigation Ready (D5, mock data) | `TripAssignedCard` +Pickup ETA/Distance/status; `TripOfferState` extended to 8 values (+`navigatingToPickup`); `RouteProgressModel`/`TrafficLevel`; `DriverTripOfferRepository.fetchRouteProgress()`; new `DriverNavigationCard`/`DriverStatusBanner`/`RouteProgressIndicator`/`RouteStatTile`; mock progress ticker (100%→20%); Navigation Preview | Accept Flow & Dispatch Session (D-04) | **Done (Phase D-05)** — UI portion of Roadmap stage D5 |
| Arrived at Pickup (D5, mock data) | Route-progress ticker now runs 100%→0%; `TripOfferState` extended to 9 values (+`arrivedAtPickup`, driven by progress completion, not a timer); new `WaitingTimer`/`WaitingFeeCard`/`PassengerActionPanel`/`DriverArrivalCard`; Arrival Preview | Assigned & Navigation Ready (D-05) | **Done (Phase D-06)** — completes the UI portion of Roadmap stage D5 |
| Auth + onboarding (D2) | Login + KYC document submission | Identity register/login (missing), Driver service (exists) | Not started |
| Go online/offline (real wiring) | Replace `AvailabilityToggle`'s mock transitions with the real Driver Availability API + background location | Driver Availability API (exists) | Not started |
| Offer handling (real wiring) | Replace `DriverTripOfferRepository`'s single mock offer with a live Dispatch offer feed + real accept/reject calls | Dispatch API (exists) | Not started |
| Navigation to pickup (real wiring) | Replace the mock `RouteProgressModel` ticker with a real maps/directions provider + live GPS | a maps/directions provider (missing) | Not started |
| Trip lifecycle, earnings (D6–D7) | Passenger boarding, start/complete trip, real earnings data | See Section 6 | Not started |

### Admin Portal
Entire app: not started. See Section 8.

### Infrastructure
| Item | Status |
|---|---|
| Docker Compose (Postgres/Redis/Kafka) | Done |
| DB migration tooling | Not started |
| Proto build automation | Not started (manual today) |
| Kafka actually used by any service | Not started |

### Testing
| Item | Status |
|---|---|
| Backend unit tests | Substantial coverage exists per-service |
| Backend tests running in CI | Only `shared` — gap |
| CI build coverage of all 16 services | Missing `booking`, `gateway` — gap |
| Integration tests (Postgres/Redis) | Exist, skipped in CI (no env vars set for service jobs) |
| Rider app automated tests | 15 smoke tests (`apps/rider/test/widget_test.dart`, R-01 through R-04) — this row was stale until now; see §2.6 |
| Driver app automated tests | 40 smoke tests (`apps/driver/test/widget_test.dart`, D-01–D-06): tab shell/switching/Developer page (D-01); Home dashboard + toggle states + dev states + Quick Action (D-02); Trips offer load/Reject/countdown-expiry/dev states/preview menus (D-03); Accept→Assigned/Failed/Timeout, countdown-cancel race guard, loading-button disabled, Retry, Assigned data, onNavigate-once (D-04); Assigned→Navigating transition, progress render/update, traffic badge, Cancel/Contact plain callbacks, Navigation Preview menu entry + step-through (D-05); Navigating→Arrived transition, WaitingTimer count-up/minute-callback/dispose-safety, WaitingFeeCard free/charged thresholds, PassengerActionPanel callbacks, DriverArrivalCard composition, Arrival Preview menu entry + step-through (D-06) |

### Deployment
No deployment pipeline, staging environment, or release process exists yet beyond the local Docker Compose dev environment and the CI build/lint/test jobs described above.

---

## 5. Rider App Roadmap

| Stage | Objective | Required Dependencies | Estimated Complexity |
|---|---|---|---|
| R1. Toolchain verification | Run `flutter pub get` + `flutter analyze` on a working Flutter install; fix any surfaced issues | none | **Done — see Phase Registry R-01. Fixed: pubspec SDK constraint `^3.12.2` → `^3.9.2` (installed toolchain is Flutter 3.35.4 / Dart 3.9.2); `ThemeData.cardTheme` `CardTheme`→`CardThemeData`; `unnecessary_underscores` lint; stale `test/widget_test.dart` referencing a nonexistent `MyApp` class replaced with a real smoke test.** |
| R1b. Booking UI (mock data) | Booking Bottom Sheet, Pickup/Destination Cards, Vehicle Selector, Fare Summary, Payment Method Card, Promo Code Entry, Book Ride Button — all mock data, reusing `TripSelection` | R1 | **Done — Phase R-01, 2026-07-06.** This was UI construction only; see R2/R4 below for real wiring, which remains a separate, not-yet-started step (consistent with Section 9: Identity/backend readiness blocks *wiring*, not UI construction). |
| R2. API client layer | Generic HTTP client wired to Gateway base URL, token storage/refresh | Identity login endpoint must exist | Medium |
| R3. Auth screens | Phone entry, OTP (or interim password) login, session persistence | R2, Identity register/login | Medium |
| R4. Booking submission (real wiring) | Connect the Booking UI (R1b) to a real fare estimate call and `POST /api/v1/rides`, replacing `MockFareBreakdown`/`MockBookingCatalog`/`MockPromoValidator` with live Gateway responses | R2, Gateway booking routes (exist) | Medium |
| R5. Trip tracking screen | Show driver location + ETA while `driver_assigned`/`in_progress` | Driver App location broadcast, Dispatch status endpoint | **UI done — Phase R-02, 2026-07-06** (Searching/Assigned/Arriving/InProgress/Completed views, mock driver + mock ETA, `MockTripRepository` timer). **Real wiring remains High complexity / not started** — still needs the Driver App location broadcast and Dispatch status endpoint listed above. |
| R6. Payment UI (real wiring) | Replace the Payment Method Card's mock methods with real Wallet/Payment integration | Wallet/Payment backend | High (backend not started) |
| R7. Rating UI | Post-trip star rating + comment | Review backend | Low-Medium |
| R8. Profile completion | Wire Profile tab to User service | R2, existing User service | **UI done — Phase R-03, 2026-07-06** (Profile Screen, Settings with 9 reusable entries, Notification Center with unread badge + Loading/Success/Empty/Error states, all mock data). **Real wiring remains not started** — still needs R2 (API client) and the existing User service to be called for real. |
| R9. Ride History (new stage, not in the original roadmap) | Trip History list + filters, Trip Detail, Receipt | Booking UI, Trip module, Profile module (all exist) | **UI done — Phase R-04, 2026-07-06** (all mock data, reusing `FareSummaryCard`/`PickupCard`/`DestinationCard`/`DriverInfoCard`/`AsyncStateView`). **Real wiring remains not started** — needs R2 (API client) and the existing Trip service to be called for real. |

*(No implementation details specified — stage descriptions only, per instructions.)*

---

## 6. Driver App Roadmap

| Stage | Objective | Required Dependencies | Estimated Complexity |
|---|---|---|---|
| D1. Project scaffold | Create `apps/driver` Flutter project, decide whether to share code/packages with `apps/rider` | none | **Done — Phase D-01, 2026-07-06.** Decision: no shared package for now — `apps/driver` is a fully separate Flutter project; the same design system is hand-mirrored in its own `AppTheme` (distinct accent color) rather than imported. Delivered: `DriverApp` entry point, named `go_router` routes, 5-tab bottom nav (Home/Trips/Earnings/Notifications/Profile) with placeholder pages, Developer page (Development Utilities) under Profile, `features/<name>/{domain,data,presentation}` folder convention. |
| D2. Auth + onboarding | Login + KYC document submission flow | Identity register/login (missing), Driver service (exists) | Medium |
| D3. Go online/offline | Toggle availability, background location updates | Driver Availability API (exists) | **UI done — Phase D-02, 2026-07-06** (`AvailabilityToggle` — 4 states, mock transition delay; `HomeStatusCard` — 4 messages). **Real wiring remains not started** — still needs the Driver Availability API and background location. |
| D4. Offer handling | Receive dispatch offer, accept/reject with timeout | Dispatch API (exists) | **UI done — Phase D-03 (2026-07-07, offer/countdown/accept-reject) + Phase D-04 (2026-07-07, accept-flow: Accepting → Assigned/Failed/Timeout, `TripAssignedCard`, `DispatchStatusBanner`, Retry)**. **Real wiring remains not started** — still needs the live Dispatch API for the offer feed and real accept/reject/confirm calls. |
| D5. Navigation to pickup/dropoff | Map + directions during active trip | Trip API (exists), a maps/directions provider | **UI done (pickup leg, through Arrived) — Phase D-05, 2026-07-07** (`TripAssignedCard` +Pickup ETA/Distance/status; Start Navigation drives `assigned → navigatingToPickup`; new `DriverNavigationCard`/`DriverStatusBanner`/`RouteProgressIndicator`/`RouteStatTile`; mock `RouteProgressModel`/`TrafficLevel` ticker 100%→20%; Navigation Preview) **+ Phase D-06, 2026-07-07** (ticker now runs 100%→0%, reaching 0 drives `navigatingToPickup → arrivedAtPickup` — not a hard timer; new `WaitingTimer`/`WaitingFeeCard`/`PassengerActionPanel`/`DriverArrivalCard`; Arrival Preview). **The dropoff leg is not built yet (`passengerBoarding` — Roadmap stage D6 — doesn't exist); real wiring remains not started** — still needs a maps/directions provider and live GPS. |
| D6. Trip lifecycle actions | Start trip, complete trip, view fare | Trip/Booking/Pricing APIs (exist) | Low-Medium |
| D7. Earnings dashboard | View completed trips + payout balance | Wallet backend (missing) | High (backend not started) |

---

## 7. Backend Roadmap

| Stage | Objective | Required Dependencies | Estimated Complexity |
|---|---|---|---|
| B1. Identity auth completion | Register/login/OTP use cases + expose via gRPC or Gateway HTTP | Existing Identity domain layer | Medium |
| B2. Gateway route expansion | Expose Identity auth, User profile, Driver profile/availability through Gateway | B1, existing services | Low-Medium |
| B3. Driver-location reconciliation | Decide whether Driver's availability system and Dispatch's location system should merge or stay separate, document the decision | none (design decision, not new code) | Low (decision) / Medium (if merged) |
| B4. Wallet foundation | Balance entity, ledger persistence | none | Medium |
| B5. Payment integration | Card/wallet charge on trip completion | B4, ADR-0014 (payment method decision — pending) | High (external dependency + unresolved ADR) |
| B6. Notification delivery | Push/SMS on trip events | Event source (Kafka, currently unused, or direct calls) | Medium |
| B7. Review/rating persistence | Store ratings post-trip | Trip completion event | Low-Medium |
| B8. Admin backend API | Read trips/users/drivers, issue refunds, resolve disputes | B4/B5 (for refunds) | Medium-High |
| B9. Analytics aggregation | Basic reporting for MVP KPIs (WFM, trip counts) | Event source or scheduled queries | Medium |
| B10. CI hardening | Run per-service tests in CI, add `booking`/`gateway` to build matrix, add migration tooling | none (process/config work) | Low |

---

## 8. Admin Portal Roadmap

| Stage | Objective | Required Dependencies | Estimated Complexity |
|---|---|---|---|
| A1. Technology ratification | Approve ADR-0003 (Next.js) or confirm still the intended stack | none (decision) | Low |
| A2. Project scaffold | Create the admin web project, auth-gated shell | A1, Identity auth (B1) | Low-Medium |
| A3. Trip/ops view | List/search trips, view status | Trip API (exists) | Medium |
| A4. Dispute/refund handling | Resolve disputes, issue refunds | Backend Wallet/Payment (B4/B5) | High (backend not started) |
| A5. Driver ops | View driver KYC status, verify/reject drivers | Driver API (exists) | Medium |

---

## 9. Parallel Development Matrix

Quick-reference table: what can start today, what it's actually blocked by (for *real* wiring/integration, not for UI construction). Consult this before picking up any new work item; update it whenever a module's blocked-by status changes (see Section 14, Rule 5).

| Module | Can develop now in parallel? | Blocked by (for real integration only) | Note |
|---|---|---|---|
| Rider UI — screens (auth, booking, tracking, payment placeholders) | ✅ | — | Build against MockAuth + mock API responses |
| Rider UI — real booking submission | ✅ | Gateway booking routes (already exist) | Not blocked — can wire for real now |
| Rider UI — real auth | ⚠️ shell only | Identity register/login | UI can be built now; real token exchange waits on B1 |
| Driver UI — entire app scaffold + screens | ✅ | — | Build against MockAuth + mock dispatch offers |
| Driver UI — real offer accept/reject | ✅ | Dispatch API (already exists) | Not blocked — can wire for real now |
| Driver UI — real auth | ⚠️ shell only | Identity register/login | Same as Rider |
| Identity — register/login/OTP | ✅ | — | No dependency; build in parallel with both UI tracks |
| Gateway — route expansion (B2) | ✅ | — | No dependency |
| Admin UI — scaffold, trip/driver read views | ✅ | — | Trip and Driver read APIs already exist |
| Admin UI — dispute/refund | ⚠️ shell only | Wallet + Payment backend | Shell can be scaffolded; not functional until backend exists |
| Payment UI (Rider checkout) | ⚠️ shell only | Payment/Wallet backend + ADR-0014 decision | Screen can be scaffolded; not functional until both land |
| Wallet backend | ✅ | — | No dependency |
| Payment backend | ❌ | Wallet backend, ADR-0014 | Needs ledger to exist + payment method decided first |
| Notification backend | ✅ | — | No dependency |
| Review backend | ✅ | — | Trip completion already fires; can be polled |
| Analytics backend (basic/polling) | ✅ | — | Polling approach available now |
| Analytics backend (event-driven) | ❌ | Kafka adoption decision | No service publishes events today |
| Driver-location reconciliation decision (B3) | ✅ | — | A decision task, not code — do before Driver UI D3/D4 are considered final |
| CI hardening (per-service tests, build matrix) | ✅ | — | No dependency; should happen immediately |
| Geo service | ⏸️ paused | Scope decision (tied to B3) | Do not start until clarified whether still needed |
| **Integration Testing / End-to-End (M4)** | ❌ | Identity register/login (real) + Rider UI real-wired + Driver UI real-wired | The actual gate — everything above can run in parallel up to this point |

---

## 10. Recommended Development Order

Goal: unblock the end-to-end flow with minimal wasted work and maximum parallelism. Structure follows the Parallel Development Matrix (Section 9): a **Foundation** stage feeds four parallel tracks, which converge at **Integration**, then **Beta**.

```
Foundation
   (CI hardening — B10; MockAuth contract defined — Rule 5; driver-location
    reconciliation decision — B3)
   │
   ├── Backend Core       (B2 Gateway route expansion → B4 Wallet foundation)
   ├── Rider UI            (R1 → R2 → R4, built against MockAuth + mock data)
   ├── Driver UI            (D1 → D2 → D3 → D4 → D6, built against MockAuth + mock data)
   └── Identity Integration (B1 — register/login/OTP, built independently of the UI tracks)
   │
   ▼
Integration Testing / End-to-End
   (swap MockAuth → real Identity in both apps; swap mock data → real API
    responses; run Milestone M4)
   │
   ▼
Beta
   (B5 Payment + R6 + D7 land together; B7 Review + R7; A1→A5 Admin portal;
    B6 Notifications + B9 Analytics last)
```

**Foundation (start immediately, no dependencies):**
1. CI hardening (B10) — per-service tests in CI, `booking`/`gateway` added to build matrix
2. MockAuth contract definition (Rule 5) — agree the mock token/session shape once, shared by both UI tracks
3. Driver-location reconciliation decision (B3) — needed before Driver UI's D3/D4 are trustworthy

**Four parallel tracks (no file-level overlap, can run concurrently across engineers/agents):**
- **Track A — Rider UI:** R1 → R2 (against MockAuth) → R4 (booking submission, real Gateway routes) → R8 (profile)
- **Track B — Driver UI:** D1 → D2 (against MockAuth) → D3 → D4 (real Dispatch API) → D6 (skip D5/navigation initially)
- **Track C — Backend Core:** B2 (Gateway route expansion) → B4 (Wallet foundation, needed later by D7 and A4)
- **Track D — Identity Integration:** B1 (register/login/OTP) — proceeds independently; does **not** block Tracks A/B, only gates the Integration step below

**Integration Testing / End-to-End (the actual gate — this is where Identity becomes required):**
Once Track D (Identity) and Tracks A/B (UIs) are each individually done, swap MockAuth for real Identity in both apps and verify Milestone M4. This is the only point in the plan where Identity is a hard blocker.

**Beta wave (after M4):**
- B5 (Payment) + R6 (Rider payment UI) + D7 (Driver earnings) — tightly coupled, land together
- B7 (Review) + R7 (Rider rating UI) — small, slot in anytime after Tracks A/B
- A1 → A2 → A3 → A5 (Admin portal core) can start as soon as B2 exists; does not need to wait for payment
- A4 (dispute/refund) waits on B5

**Deliberately last:**
- B6 (Notifications), B9 (Analytics) — MVP-required per DOC-0002 but not on the critical path to a working booking loop
- Geo service — scope unclear (see Section 4); do not start until B3 clarifies whether a standalone Geo service is still justified

---

## 11. Milestones

| Milestone | Objective Completion Criteria |
|---|---|
| **M0 — Foundation Ready** | CI hardened (per-service tests running, build matrix complete); MockAuth contract agreed and available to both UI tracks; driver-location reconciliation decision documented |
| **M1 — Backend Core Complete** | Identity, Gateway, User, Driver, Trip, Dispatch, Pricing, Booking all have CI-verified tests (per-service, not just `shared`) and the build matrix includes all active services |
| **M2 — Rider MVP Complete** | Rider app can log in (MockAuth acceptable), select pickup/destination, get a fare estimate, submit a booking, and see booking status through real Booking/Pricing API calls — no static placeholders remaining. Real Identity login not required yet. |
| **M3 — Driver MVP Complete** | Driver app exists, can log in (MockAuth acceptable), go online, receive and accept/reject a dispatch offer through the real Dispatch API, and complete a trip. Real Identity login not required yet. |
| **M4 — End-to-End Booking** | MockAuth replaced with real Identity login in both apps; a rider and a driver, using their respective real apps against the real backend, complete one full trip from request to completion without manual database intervention. **This is the milestone where Identity is a hard blocker.** |
| **M5 — Payment Integrated** | Wallet + Payment backend exist; Rider can pay and Driver sees the earning reflected |
| **M6 — Admin Core Ready** | Admin portal can view trips, verify a driver, and issue one refund against a real trip |
| **M7 — Closed Beta Ready** | M0–M6 complete; notifications fire on key trip events; basic analytics report the MVP KPIs from DOC-0002 §6.19 (driver count, daily trips, availability) |

---

## 12. Phase Registry

*This table starts empty as of this document's adoption. Historical phases completed before this document existed (Phase 1 through Phase 17, plus Hardening H2–H4) remain recorded in `.ai/memory.md` and `CHANGELOG.md` and are not re-listed here. Every phase completed from this point forward must be appended below.*

| Phase | Name | Status | Depends On | Completed Date |
|---|---|---|---|---|
| R-01 | Rider App — Booking UI Module (Section 5, stages R1 + R1b) | Complete | Map + Location Engine, `TripSelection` (Phase 17) | 2026-07-06 |
| R-02 | Rider App — Trip Lifecycle UI (Section 5, stage R5 — UI portion) | Complete | R-01 (Booking UI: `TripSelection`, `FareSummaryCard`, `PickupCard`/`DestinationCard` reused) | 2026-07-06 |
| R-03 | Rider App — Profile Module (Section 5, stage R8 — UI portion) | Complete | none (standalone; relocates the R-02 dev preview entry point into the new Settings screen) | 2026-07-06 |
| R-04 | Rider App — Ride History Module (Section 5, new stage R9 — UI portion) | Complete | R-01 (`FareSummaryCard`/`PickupCard`/`DestinationCard`), R-02 (`DriverInfoCard`/`MockDriver`), R-03 (`AsyncStateView`, rider name) | 2026-07-06 |
| D-01 | Driver App — Initial Project Scaffold (Section 6, stage D1) | Complete | none (new, separate Flutter project — no shared package with `apps/rider`) | 2026-07-06 |
| D-02 | Driver App — Home Dashboard Module (Section 6, stage D3 — UI portion) | Complete | D-01 (app shell, `AsyncStateView` hand-mirrored from Rider R-03) | 2026-07-06 |
| D-03 | Driver App — Incoming Trip Module (Section 6, stage D4 — UI portion) | Complete | D-01 (app shell), D-02 (`AsyncStateView`, dev-preview-menu convention) | 2026-07-07 |
| D-04 | Driver App — Accept Flow & Dispatch Session (Section 6, stage D4 — UI portion, completes it) | Complete | D-03 (offer state machine, `TripOfferCard`/`FareEstimateCard`, dev-preview convention) | 2026-07-07 |
| D-05 | Driver App — Assigned & Navigation Ready (Section 6, stage D5 — UI portion, pickup leg only) | Complete | D-04 (`TripAssignedCard`, offer state machine, dev-preview convention) | 2026-07-07 |
| D-06 | Driver App — Arrived at Pickup (Section 6, stage D5 — UI portion, completes the pickup leg) | Complete | D-05 (route-progress ticker, `RouteProgressModel`, dev-preview convention) | 2026-07-07 |

---

## 13. Risks

Grounded in Section 2 findings only — no hypothetical risks added.

1. **No functioning real auth flow.** Identity has entities, persistence, and JWT machinery, but no register/login use case and no API surface. This blocks **Integration Testing (Milestone M4)**, not Rider/Driver UI construction — both UI tracks should proceed now against MockAuth (see Section 9, Section 10 Track D, Section 14 Rule 5). The risk to watch is MockAuth drifting from the real Identity/Gateway contract, not the UI tracks being idle.
2. **CI does not test the services it builds.** The `test` job only runs `backend/shared`; ~460 existing service-level tests never run in CI. Regressions in Identity/Driver/Trip/Dispatch/Pricing/Booking/Gateway would currently pass CI even if fully broken.
3. **CI build matrix is out of date.** `booking` and `gateway` are missing from the build matrix — they are not even compile-checked in CI.
4. **No database migration tooling.** Every service's schema exists only inside Go test helper functions. There is no documented, repeatable way to provision a real (non-test) Postgres instance with the correct schema today.
5. **Driver presence is tracked in two independent, unreconciled systems** (Driver service's own Redis heartbeat + Postgres `online_status`, versus Dispatch's own separate Redis GEO layer). This was a deliberate MVP trade-off per `.ai/memory.md` but has not been revisited since, and will get harder to fix the more the Driver App is built on top of it unexamined.
6. **No Driver App exists.** The backend has full Driver/Vehicle/Availability support, but there is no client to exercise it. Two-sided marketplace behavior (the core of the product) cannot be manually or automatically verified until this exists.
7. **No Admin Portal exists.** DOC-0002 lists admin portal (disputes, refunds, city ops) as in-MVP-scope; zero code exists.
8. **Rider app is not connected to any backend.** Beyond GPS/maps, the app makes no network calls. The booking screen is still Phase-12 static placeholder content, not wired to the pickup/destination selection already built in Phase 17.
9. **Kafka is provisioned but unused.** The Constitution (CARCH-004) calls for event-driven communication for significant state changes; today 100% of inter-service communication is synchronous gRPC. This is not necessarily wrong for MVP, but it is a gap between the ratified architectural principle and current implementation that DOC-0003 (still blocked) will eventually need to address.
10. **Payment method is architecturally undecided.** ADR-0014 (MVP payment methods) is still pending. Wallet and Payment services are empty skeletons, and this blocks Milestone M5 and Admin refund handling (M6).
11. **Flutter toolchain state is unverified.** `flutter pub get`/`flutter analyze` have never successfully run against `apps/rider`. Dependency resolution and lint cleanliness are unknown until this happens on a working machine.
12. **Geo service scope is unclear.** A `geo` service skeleton exists in the service catalog, but Driver and Dispatch have each already implemented their own Redis-GEO logic independently. It is not documented whether `geo` is still meant to be built, or whether it was superseded by that in-service logic.
13. **Foundational governance documents are still unapproved.** DOC-0001/0001A/0002 remain Draft; DOC-0003/0004 are blocked. This does not block engineering today, but any open question resolved late (e.g., launch market, driver classification, cash collection) could still retroactively change service boundaries already built.

---

## 14. Future Rules

1. **Every future implementation phase must reference this roadmap.** Any coding prompt or task that adds a backend module, mobile feature, or admin feature should state which roadmap stage (Sections 5–8) it corresponds to.
2. **No feature may be implemented outside this roadmap without explicit approval.** If a requested feature does not map to an existing stage or milestone here, this document should be updated first (or the deviation explicitly approved) rather than building ad hoc.
3. **Completed phases must update the Phase Registry (Section 12).** Every phase that finishes should add a row: Phase number, Name, Status, Depends On, Completed Date. This is in addition to — not a replacement for — the existing `.ai/memory.md` and `CHANGELOG.md` update habits already in use.
4. **This document is the authoritative implementation roadmap for the MVP** until superseded by DOC-0004 (Implementation Master Plan) once that document is unblocked and generated, at which point this document should be reconciled with it rather than left to silently diverge.
5. **MockAuth is a contract, not a shortcut.** Any UI work built against MockAuth (Rider UI, Driver UI — Section 9) must consume a mock auth/session interface that mirrors the real Identity/Gateway auth contract exactly (same token shape, same header convention, same error responses). Swapping MockAuth for real Identity at Integration Testing (Milestone M4) must be a configuration change, not a client-code rewrite. If the real Identity contract changes after MockAuth is defined, MockAuth must be updated in the same change — they do not drift independently.
6. **Consult the Parallel Development Matrix (Section 9) before starting any new work item**, and update it immediately whenever a module's parallel-safe or blocked-by status changes — it is the fastest way for anyone (human or AI) to answer "what can I work on right now."
