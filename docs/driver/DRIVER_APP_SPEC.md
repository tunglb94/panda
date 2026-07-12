# PandaDriver — Driver App Product & Technical Specification

**Status:** Draft v1.0
**Owner:** Product / Flutter Platform
**Scope:** `apps/driver` (Flutter) only. Backend is out of scope except where an endpoint must be *added* to support a screen — those are explicitly flagged.
**Source material:** 8 reference screenshots of a third-party ride-hailing driver app (phone-number + password login, map-centric home screen, wallet, earnings, profile). This document reverse-engineers the *product logic* behind those screens — not the pixels, not the brand — and re-specifies it for PandaDriver's own design system, backend contracts, and current codebase.
**Non-goal:** This is not a clone spec. Every screen, copy string, and color below is PandaDriver's own. Where the reference app's business logic is ride-hailing-industry-standard (surge zones, wallet holds, per-trip fee breakdown), we adopt the *concept*, not the *implementation*.

---

## How to read this document

- ✅ **Implemented** — this exact capability already exists in `apps/driver` and/or the Go backend today. Verified against the current codebase as of this writing.
- 🔶 **Proposed** — needed for the feature to work, but the backend endpoint/table does not exist yet. Treated as a backend work item, not assumed to exist.
- 🧪 **Preview-only** — exists in the codebase as a developer-only mock/preview screen (reachable from `DeveloperPage`), not part of the real user-facing flow.
- 🚫 **Not applicable** — reference-app behavior we are deliberately not adopting (e.g. multi-vertical beRide/beFood/beDelivery switching — PandaDriver is ride-hailing only for now).

Every screen section ends with an **Implementation status** line so a future engineer (or another Claude Code session) can tell at a glance what to build vs. what to wire up.

---

# 1. Project Goal

## 1.1 Purpose of the Driver App

PandaDriver is the counterpart app to Panda (rider). Its job is narrow and specific: let a driver earn money by fulfilling ride requests dispatched by the Panda platform, with full visibility into what they're earning and why. Everything in this document serves one of three outcomes:

1. **Get the driver online, fast.** Every extra tap between "app open" and "receiving trip offers" is lost earning time for the driver and lost supply for the platform.
2. **Make every payout number explainable.** The single biggest driver-trust failure mode in ride-hailing is "I don't understand why I got paid this amount." Every screen that shows money must be traceable to a line-item breakdown.
3. **Never leave the driver stuck.** A driver mid-trip with a dead map, an expired offer, or a frozen "loading" screen is a driver who logs into a competitor's app tomorrow. State transitions must always have a visible next action or an automatic retry.

## 1.2 Primary user

- A driver-partner, not an employee. Independent contractor mental model: they choose when to go online, they see costs (fees, tax withholding) transparently, they are not micromanaged by the app.
- Vietnam market first (confirmed by existing Vietnamese-language UI, VND currency formatting, and phone-number-first auth already in the codebase). Assume Android-first, mid-range devices, intermittent 4G, and a driver who is often holding the phone in a mount, glancing at it while riding/driving — large tap targets, minimal reading required mid-trip.
- Vehicle types: motorbike (`Xe máy`) is the primary vehicle category already modeled in `VehicleCategory` on the rider side; car/van also exist. This spec is vehicle-type-agnostic except where noted (e.g. helmet/2-wheeler-specific safety copy is out of scope for v1).

## 1.3 Main responsibilities of the app

| Responsibility | Already built? |
|---|---|
| Authenticate the driver | ✅ phone-number login, no password (`LoginPage`, `AuthRepository`) |
| Let the driver toggle online/offline | ✅ `AvailabilityRepository`, `MapPage` toggle |
| Stream the driver's live location to dispatch while online | ✅ `LocationUploadService` |
| Receive and respond to trip offers (accept/reject with countdown) | ✅ `TripPage`, `TripOfferRepository` |
| Guide the driver through pickup → in-trip → drop-off | ✅ `TripPage` state machine |
| Show trip history | ✅ `DriverTripHistoryPage` |
| Show earnings (this trip, this day/week/month) | 🔶 `EarningsPage` exists as a stub; no real earnings backend |
| Show wallet balance + withdrawal | 🔶 not built — no wallet service in backend |
| Push notifications for new offers / announcements | 🔶 `NotificationsPage` is a placeholder tab; no push infra |
| Driver profile, rating, vehicle/document info | 🔶 `ProfilePage` shows basic info only |

## 1.4 Driver lifecycle (top level)

```
Install → Sign up (out of scope — dev-seeded accounts only, see §5) → Login
   → Grant permissions → Home (Offline)
   → Go Online → Wait for offer → Accept/Reject
   → Navigate to pickup → Arrive → Start trip → Navigate to drop-off
   → Finish trip → Rider pays → Rate rider → Back to "Wait for offer"
   → (end of shift) Go Offline → Logout (rare — sessions persist)
```

This loop (Offline → Online → Trip → Online → …) is the entire product. Sections 3, 8, and 9 formalize it.

---

# 2. App Architecture

## 2.1 Architectural style actually in use (important — do not introduce Bloc)

The existing codebase is **feature-first, StatefulWidget + Repository**, *not* Bloc/Cubit, *not* Riverpod, *not* Provider. `pubspec.yaml` has no state-management package at all — every page manages its own `State` with `setState`, delegates I/O to a plain-Dart `Repository` class constructed with an `ApiClient`, and passes callbacks down through constructors. This mirrors `apps/rider` exactly (same author, same conventions, confirmed by reading both codebases side by side).

**Decision: keep this pattern.** Do not introduce a state-management library for this spec's proposed screens (Wallet, Earnings, Notifications). Consistency with the existing 30+ files matters more than any Bloc idiom. Section 24 ("Bloc/Cubit Inventory") is retained as a heading for parity with product-spec convention but its content lists **Controller classes** (the `State<T>` + Repository pairing already used) instead of literal Bloc/Cubit types.

## 2.2 Layering

```
lib/
  core/                      # cross-cutting: no feature imports another feature's internals
    auth/                    # AuthState (ChangeNotifier-like, drives GoRouter redirect)
    config/                  # AppConfig — String.fromEnvironment dart-defines
    location/                # LocationEngine — geolocator wrapper, distance-filtered stream
    network/                 # ApiClient — thin http wrapper, throws ApiException
    router/                  # AppRouter — GoRouter, StatefulShellRoute bottom-nav shell
    storage/                 # TokenStorage — shared_preferences-backed JWT persistence
    theme/                   # AppTheme — Material 3 ColorScheme + typography
    trip_metrics/            # TripMetricsEngine — client-side distance/duration accumulation during a trip
  features/
    auth/
      data/                  # AuthRepository — POST /api/v1/auth/login
      presentation/pages/    # LoginPage
    home/
      data/                  # AvailabilityRepository, DriverHomeRepository (mock)
      domain/models/         # DriverActivityStatus, DriverAvailabilityStatus, DriverHomeSummary
      presentation/
        pages/                # HomePage (legacy — superseded by MapPage as the live home screen)
        widgets/              # AvailabilityToggle, DriverSummaryHeader, HomeStatsRow, HomeStatusCard, QuickActionCard
    location/
      data/                  # LocationUploadRepository — POST /api/v1/driver/location
      services/              # LocationUploadService — periodic upload while online
    map/
      presentation/pages/    # MapPage — THE live home screen (see §7)
    trip/                     # THE live trip-lifecycle module (see §9)
      data/                  # ActiveTripRepository, TripOfferRepository
      presentation/pages/    # TripPage — single state-machine page for offer→completion
    trips/                    # legacy/preview module — dev-only, reachable from DeveloperPage
      data/                  # DriverTripOfferRepository (mock catalog)
      domain/models/         # richer trip-offer domain models (kept for future real use)
      presentation/          # *_preview_page.dart — static state previews for design QA
    earnings/
      presentation/pages/    # EarningsPage — stub, see §11
    notifications/
      presentation/pages/    # NotificationsPage — stub, see §13
    profile/
      data/                  # MockAppInfoRepository
      domain/models/         # AppInfo, BuildModeKind
      presentation/pages/    # ProfilePage, DriverTripHistoryPage, DeveloperPage
  shared/
    widgets/                 # ScaffoldWithNav (bottom nav shell), AsyncStateView, PlaceholderTabContent
```

**Naming clarification for future engineers:** `features/trip` (singular) is live production code. `features/trips` (plural) is a design-preview sandbox. Do not add real functionality to `features/trips` — either promote its domain models into `features/trip` when building Wallet-adjacent trip detail views, or delete it once no longer needed for QA.

## 2.3 State management (the actual pattern)

Every screen with async data follows this shape:

```dart
class XPage extends StatefulWidget { ... }

class _XPageState extends State<XPage> {
  late final XRepository _repo;
  late Future<XModel> _future;

  @override
  void initState() {
    super.initState();
    _repo = XRepository(apiClient: widget.apiClient);
    _future = _repo.fetch();
  }

  @override
  Widget build(BuildContext context) => AsyncStateView<XModel>(
    future: _future,
    successBuilder: (context, data) => ...,
  );
}
```

`AsyncStateView<T>` (in `shared/widgets/async_state_view.dart`) is the one reusable primitive that standardizes Loading / Success / Empty / Error rendering — every new screen in this spec (Wallet, Earnings detail, Notifications) must use it rather than hand-rolling a `FutureBuilder`.

## 2.4 Dependency injection

No DI framework. `main.dart` constructs the object graph by hand (`ApiClient`, `AuthState`, `TokenStorage`, `LocationUploadService`) and threads them down through widget constructors. New repositories (WalletRepository, EarningsRepository, NotificationsRepository) follow the same pattern: constructed once in `main.dart` or lazily inside the owning page's `initState`, given an `ApiClient`.

## 2.5 Navigation

`go_router` with a `StatefulShellRoute.indexedStack` for the 5-tab bottom nav (see §3). Auth-gated via `redirect:` reading `AuthState.isLoggedIn`, exactly mirroring the rider app.

## 2.6 Folder structure additions this spec requires

```
lib/features/wallet/
  data/wallet_repository.dart
  domain/models/wallet_summary.dart
  domain/models/wallet_transaction.dart
  presentation/pages/wallet_page.dart
  presentation/pages/withdraw_page.dart
  presentation/pages/wallet_history_page.dart
  presentation/widgets/wallet_balance_card.dart

lib/features/earnings/
  data/earnings_repository.dart              # replaces the current stub
  domain/models/earnings_summary.dart
  domain/models/earnings_entry.dart
  presentation/pages/earnings_page.dart       # rewritten, not a stub
  presentation/pages/earnings_detail_page.dart
  presentation/widgets/earnings_period_tabs.dart
  presentation/widgets/earnings_breakdown_row.dart

lib/features/notifications/
  data/notifications_repository.dart
  domain/models/driver_notification.dart
  presentation/widgets/notification_tile.dart

lib/features/profile/
  presentation/pages/vehicle_info_page.dart
  presentation/pages/documents_page.dart
  presentation/pages/support_page.dart
  presentation/pages/rating_detail_page.dart
```

---

# 3. Complete Navigation Graph

```
Splash (native launch screen, no Flutter route)
  ↓
Login  [/login]
  │  (phone number → OTP-less login per current backend; see §5)
  ↓
Permission Gate  [inline — not a route, triggered from MapPage.initState]
  │  Location permission → GPS enabled check
  ↓
┌─────────────────────────── Bottom Nav Shell (5 tabs) ───────────────────────────┐
│                                                                                    │
│  Tab 1: Home (Map)  [/]                                                          │
│    ├─ Wallet  [/wallet]                        (tap "Xem số dư")                 │
│    │    ├─ Wallet History  [/wallet/history]                                     │
│    │    └─ Withdraw  [/wallet/withdraw]                                          │
│    ├─ Trip Offer (in-place, not a route — TripPage owns this state)              │
│    │    ├─ Offer countdown card (dialog-like, in-page)                           │
│    │    ├─ Active Trip card (in-page)                                            │
│    │    ├─ Awaiting Payment card (in-page)                                       │
│    │    └─ Rating card (in-page)                                                 │
│    ├─ Quick Actions sheet (bottom sheet)                                         │
│    │    ├─ Service configuration  [/settings/services]        🔶                 │
│    │    ├─ Nearby charging stations  [/nearby-charging]        🔶 (EV only)      │
│    │    └─ Scheduled ride request  [/scheduled-rides]          🔶                │
│    └─ Location permission dialog (native + in-app fallback card)                 │
│                                                                                    │
│  Tab 2: Trips  [/trips]                                                          │
│    (Note: this route hosts the LIVE TripPage state machine today — see §9.       │
│     If Home/Map is repurposed as a pure map+status view, Trips tab becomes the   │
│     canonical offer/active-trip surface. Recommendation in §22.)                 │
│                                                                                    │
│  Tab 3: Earnings  [/earnings]                                                    │
│    ├─ Period tabs: Ngày / Tuần / Tháng (in-page, not routes)                     │
│    └─ Earnings Detail  [/earnings/:tripId]        (tap a history row)            │
│                                                                                    │
│  Tab 4: Notifications  [/notifications]                                         │
│    └─ (list only — tapping a notification deep-links into the relevant screen)  │
│                                                                                    │
│  Tab 5: Profile  [/profile]                                                     │
│    ├─ Trip History  [/profile/trips]                            ✅ exists        │
│    ├─ Vehicle Info  [/profile/vehicle]                           🔶              │
│    ├─ Documents  [/profile/documents]                            🔶              │
│    ├─ Activity Rate (Tỉ lệ hoạt động)  [/profile/activity-rate]  🔶              │
│    ├─ Achievements (Thành tích)  [/profile/achievements]         🔶              │
│    ├─ Ratings from riders  [/profile/ratings]                    🔶              │
│    ├─ Rewards program  [/profile/rewards]                        🔶              │
│    ├─ Referral (Giới thiệu)  [/profile/referral]                 🔶              │
│    ├─ Driver Handbook  [/profile/handbook]                       🔶 (static content) │
│    ├─ Support / Help Center  [/profile/support]                  🔶              │
│    ├─ Settings  [/profile/settings]                              🔶              │
│    │    └─ Change password/PIN — 🚫 not applicable (passwordless auth)          │
│    └─ Developer (dev builds only)  [/dev]         ✅ exists (DeveloperPage)      │
└────────────────────────────────────────────────────────────────────────────────┘
      ↓ (logout, from Profile)
Login  [/login]
```

## 3.1 Dialogs, bottom sheets, and popups (every one, enumerated)

| Trigger | Type | Content | Status |
|---|---|---|---|
| First `MapPage` load, permission not yet granted | Native permission dialog + in-app blocked-state card | "PandaDriver cần quyền vị trí…" | ✅ |
| GPS service disabled | Full-screen blocked-state card | "GPS đang tắt" + "Mở Cài đặt định vị" | ✅ |
| Permission permanently denied | Full-screen blocked-state card | "Mở Cài đặt" deep-link to app settings | ✅ |
| Toggle "Bật/Tắt" tapped | Inline loading state on the toggle itself (no modal) | spinner replaces icon during `goOnline`/`goOffline` | ✅ |
| New trip offer arrives | In-page card with countdown ring (not a modal — driver could be looking at the map) | pickup, drop-off, fare estimate, distance, Accept/Reject | ✅ |
| Offer countdown reaches 0 | Auto-dismiss, return to polling state | — | ✅ |
| Driver taps "Kết thúc chuyến đi" | 🔶 confirmation dialog recommended (irreversible action, ends trip early) | "Xác nhận kết thúc chuyến?" | 🔶 proposed |
| Rating submission | In-page card (not modal), star selector + optional comment | — | ✅ |
| Withdraw funds | Full page (not a sheet) — amount entry + bank account confirmation | — | 🔶 |
| Quick actions ("Cấu hình dịch vụ" etc.) | Bottom sheet from Home | icon grid, matches reference app's floating-button cluster | 🔶 |
| Logout | Confirmation dialog | "Đăng xuất khỏi PandaDriver?" | 🔶 (currently instant, no confirm) |
| Notification tapped while trip is active | Blocking dialog *only if* it would navigate away from an active trip | "Bạn đang có chuyến đi. Rời khỏi màn hình?" | 🔶 |

---

# 4. Screen Inventory

For each screen: Purpose, Widgets, Actions, Navigation, State, API, Error handling, Skeleton loading, Empty state, Refresh behavior.

## 4.1 Login Page — ✅ implemented

- **Purpose:** authenticate the driver with a phone number; no password, no OTP (dev-seeded backend accounts).
- **Widgets:** `_Logo` (brand mark + "Ứng dụng dành cho tài xế" tagline), phone `TextField` (numeric keyboard, `+84` implicit), inline error `Text`, `FilledButton` with embedded spinner.
- **Actions:** submit on button tap or keyboard "done".
- **Navigation:** on success, `AuthState.login()` flips `isLoggedIn`, `GoRouter`'s `redirect` sends the driver to `/` (Home/Map).
- **State:** `_isLoading`, `_error` (String?).
- **API:** `POST /api/v1/auth/login` — ✅ implemented, body `{ phone }`, returns `{ access_token, driver_id }`.
- **Error handling:** `ApiException` message shown inline in red under the field; generic catch-all shows "Đăng nhập thất bại. Vui lòng thử lại."
- **Skeleton loading:** none needed (single field, instant transition to spinner-in-button).
- **Empty state:** n/a.
- **Refresh behavior:** n/a — one-shot form.
- **Gap vs. reference app:** the reference app has a phone-number screen *then* a 6-digit password screen (`Nhập mật khẩu`) as two separate steps, with "Quên mật khẩu" recovery. PandaDriver's backend is intentionally simpler (phone-only, no password) for this phase — do not add a password step without a corresponding backend change (would require new `identity` service work, out of scope for this Flutter-only spec).

## 4.2 Home / Map Page — ✅ implemented (this is the real home screen, not `HomePage`)

Covered in full in §7 (dedicated section per the requested outline). Route: `/`.

## 4.3 Trip Page — ✅ implemented

Covered in full in §9. Route: `/trips` today; recommend keeping trip state visible from Home too (see §22 for the consolidation recommendation).

## 4.4 Earnings Page — 🔶 stub today, full spec in §11

- **Current state:** `EarningsPage` is a placeholder (`PlaceholderTabContent` or minimal static content — verify against `lib/features/earnings/presentation/pages/earnings_page.dart` before starting; treat any existing content as disposable).
- **Target Purpose:** show the driver what they've earned, broken down by period and by trip, with full fee transparency (matches the reference app's "Thu nhập và các khoản thanh toán khác" screen).
- Full breakdown in §11.

## 4.5 Notifications Page — 🔶 stub today, full spec in §13

- **Current state:** placeholder tab.
- **Target:** list of driver-relevant notifications (new-offer alerts are usually *not* here — those interrupt via the Trip surface directly — this tab is for account/system/promo messages).

## 4.6 Profile Page — ✅ implemented (basic), full spec in §12

- **Current widgets:** driver name/photo, rating (static/mock), a settings-style list, and a link into `DriverTripHistoryPage`.
- **Target:** expand into the full list seen in §12.

## 4.7 Driver Trip History Page — ✅ implemented

- **Purpose:** flat list of past trips (completed/cancelled), each showing route, date, and status chip.
- **Widgets:** `ListView.separated` of trip tiles, `_StatusChip`, pull-to-refresh via `RefreshIndicator`.
- **API:** `GET /api/v1/driver/trips` — ✅ implemented.
- **Error handling:** retry button on failure (`_ErrorView` inline component, already Vietnamese-localized: "Không thể tải lịch sử chuyến đi." / "Thử lại").
- **Empty state:** "Chưa có chuyến đi nào."
- **Refresh:** pull-to-refresh re-fetches; no pagination today — 🔶 add cursor-based pagination once trip volume grows (not urgent, matches rider app's current non-paginated equivalent).

## 4.8 Developer Page — ✅ implemented (dev builds only)

- **Purpose:** QA/design tool. Links into the `features/trips` preview pages (state previews for every trip-offer/active-trip visual state) and shows `AppInfo` (build mode, version placeholder).
- **Not part of the driver-facing product** — exclude from release builds' navigation if not already gated (verify `ProfilePage` doesn't surface a "Developer" entry point to a non-debug `BuildModeKind`).

## 4.9 Wallet Page — 🔶 new, full spec in §10

## 4.10 Withdraw Page — 🔶 new, full spec in §10

## 4.11 Wallet History Page — 🔶 new, full spec in §10

## 4.12 Vehicle Info / Documents / Support / Settings pages — 🔶 new, full spec in §12

---

# 5. Login Flow

## 5.1 What exists today (✅)

- Single field: phone number.
- `POST /api/v1/auth/login` with `{ "phone": "+84900000002" }` (dev-seeded driver phone number, no self-registration — confirmed: there is no registration endpoint or UI anywhere in either app).
- Response: `{ access_token, driver_id }` — note there is **no refresh token** in the current response shape (unlike the rider app's `AuthState` which is written generically enough to support one). This is a real gap:

### 🔶 Proposed backend change: refresh tokens
The gateway's `jwt.Config` already defines `RefreshTokenTTL: 7 * 24 * time.Hour` (seen in `gateway/cmd/server/main.go`) — the *capability* exists in the JWT service layer but the login handler does not appear to issue or accept a refresh token over HTTP yet. Recommend: `POST /api/v1/auth/login` returns `{ access_token, refresh_token, driver_id }`, and a new `POST /api/v1/auth/refresh` accepts `{ refresh_token }` → new `access_token`. Without this, a driver's 15-minute access token (per `AccessTokenTTL: 15 * time.Minute`) expires mid-shift and they get logged out while driving — a serious UX problem for a driver app specifically, more so than for the rider app.

## 5.2 Session persistence — ✅ implemented

`TokenStorage` (shared_preferences-backed) persists the access token; `AuthState.initialize()` reads it on app boot and treats a present token as "logged in" without re-validating it against the server. This means:

- ⚠️ **Known gap:** if the access token has expired (15 min TTL) and the app was killed/reopened, the driver appears "logged in" client-side but every API call will 401. There is currently no interceptor that catches a 401 and forces re-login. 🔶 Proposed: `ApiClient` should catch a 401 response, clear `TokenStorage`, flip `AuthState.isLoggedIn = false` (triggering the router redirect to `/login`), and surface a one-time toast "Phiên đăng nhập đã hết hạn."

## 5.3 Logout — 🔶 partially implemented

`AuthState.logout(tokenStorage)` exists and is wired from the rider app's Profile "Sign Out" tile; verify the same wiring exists in driver `ProfilePage` (add if missing — trivial, same pattern). No confirmation dialog today; recommend adding one (§3.1) since going offline + logging out mid-shift is a costly mistake to reverse (loses queue position for incoming offers).

## 5.4 What the reference app does that we are NOT adopting

- Password/PIN (6-digit) login step — 🚫 not applicable, backend is passwordless by design for this phase.
- "Quên mật khẩu" (forgot password) flow — 🚫 not applicable, no password exists.
- Biometric unlock — deferred, see §21 (Future Features).

## 5.5 Remember login

Implicit and always-on: as long as `TokenStorage` has a token, the driver stays logged in indefinitely (subject to the 401-handling gap above). There is no "remember me" toggle — this matches ride-hailing driver-app norms (drivers should never have to re-login daily).

---

# 6. Permission Flow

## 6.1 What exists today — ✅

`MapPage._resolveLocation()` implements a clean state machine:

```
loading → (GPS service check) → gpsDisabled [terminal until user acts]
        → (permission check) → permissionDenied [terminal until user acts]
                              → permissionPermanentlyDenied [terminal, deep-links to Settings]
                              → ready
```

Each terminal state renders a full-screen `_LocationErrorView` with icon, title, message, and a single action button — no dead ends. This is good UX and should be the template for any future permission flow (notifications, background location).

**Recently fixed bug (this session):** the driver app's `AndroidManifest.xml` was missing `ACCESS_FINE_LOCATION`/`ACCESS_COARSE_LOCATION` `<uses-permission>` declarations entirely, which meant Android never showed the OS permission dialog and `Geolocator.checkPermission()` would hang/throw, leaving the map stuck on the loading spinner forever. Fixed by adding both permissions to the manifest (mirroring the rider app, which had them correctly). **This is exactly the kind of gap this spec exists to prevent recurring** — any new permission-gated feature (background location, notifications) must have its manifest entry added in the same commit as the Dart code that requests it, and verified on a real device before considering the feature done.

## 6.2 Background location — 🔶 proposed, not yet needed

The current `LocationUploadService` uploads location periodically *while the app is foregrounded and the driver is online* (`POST /api/v1/driver/location`). It does not use a foreground service or background location permission (`ACCESS_BACKGROUND_LOCATION`), meaning: **if the driver backgrounds the app while online, location uploads stop**, dispatch will consider them stale, and they will silently stop receiving new offers.

This is the single most important gap for a driver app specifically (rider app can get away with foreground-only location; a driver app cannot — drivers reasonably expect to be able to glance at another app while waiting for an offer). Recommended fix, phased:

1. **Phase A (cheap):** show a persistent, high-visibility "you're online but the app must stay open" banner. Not a real fix, but stops silent failure.
2. **Phase B (real fix):** Android foreground service (`flutter_foreground_task` or native `ForegroundService` + `ACCESS_BACKGROUND_LOCATION` permission, which requires its own separate runtime permission prompt on Android 10+, requested *after* fine/coarse location is already granted, with Play Store data-safety disclosure). iOS equivalent: `UIBackgroundModes: location` + "Always" location authorization.

## 6.3 Notification permission — 🔶 proposed

Needed once push notifications (§13) ship. Android 13+ requires explicit `POST_NOTIFICATIONS` runtime permission. Request this *after* login, *not* on first launch (best practice: request permissions in context, not upfront — matches the existing pattern of requesting location only when `MapPage` actually needs it).

## 6.4 Battery optimization exemption — 🔶 proposed

Paired with background location (6.2). Android aggressively kills background location on many OEM skins (Xiaomi, Oppo, Samsung) unless the app is exempted from battery optimization. Needs an in-app explainer screen (reference app pattern: "Để nhận cuốc xe ổn định, vui lòng tắt tối ưu hoá pin cho ứng dụng") with a deep-link to the OS battery settings page — same UX pattern already used for the GPS-disabled state.

## 6.5 Permission denied behavior — ✅ implemented, extend the pattern

Current pattern (full-screen blocking view with a single clear action) should be reused verbatim for notification and background-location permission denial, rather than inventing a new pattern.

---

# 7. Home Screen

**Route:** `/` → `MapPage`. This is the screen a driver spends the most cumulative time looking at, so it deserves the most scrutiny.

## 7.1 Reverse-engineering the reference screenshot

The reference Home screen (screenshot 1) is a full-bleed map with floating UI on top:

| Element | Reference app | Inferred purpose | PandaDriver decision |
|---|---|---|---|
| Top banner "Chưa bật hoạt động" | Orange banner, full width | Ambient online/offline status, always visible even when scrolled/panned | ✅ adopt — add as a persistent top banner in `MapPage`, currently the online/offline state is only shown via the toggle button itself; a banner improves glanceability |
| Driver avatar (top-left, circular) | Tappable, small | Quick access to profile without leaving map | 🔶 adopt — small tap target to `/profile` |
| "Xem số dư" pill button | Dark pill, chevron | Wallet balance shortcut from Home | 🔶 adopt — links to new Wallet page (§10) |
| "Bật/Tắt" pill button (top-right) | Power icon + label | The single most important control in the app | ✅ exists as `AvailabilityToggle` — verify visual prominence matches this importance (top-right placement, not buried) |
| Heatmap overlay (colored zones on map) | Yellow/tan shaded polygons | Surge/demand zones — encourages drivers to reposition toward high-demand areas | 🔶 proposed, requires new backend: a zone-demand endpoint. See §7.3 |
| "Điểm nóng" / "Nhân giá" toggle chips | Segmented control above a demand gradient bar | Two views of the same underlying demand data: hotspot zones vs. surge multiplier zones | 🔶 proposed, same backend dependency |
| Demand gradient bar "Nhu cầu thấp → Cao" | Horizontal color gradient with a label | Legend for the heatmap colors | 🔶 proposed (UI-only once data exists) |
| Floating icon buttons (right edge): service config, nearby charging, traffic layer, recenter | Vertical stack of white circular FABs | Secondary map controls/settings, kept out of the main flow | 🔶 proposed — "recenter" (`Icons.my_location` equivalent) is the only one worth building first; the rest (charging stations, traffic layer) are vehicle-type-specific or low-value for v1 |
| Bottom nav bar (5 tabs) | Trang chủ / Thu nhập / Dịch vụ / Hộp thư / Tôi | Matches PandaDriver's existing 5-tab shell almost exactly | ✅ already implemented as Home/Trips/Earnings/Notifications/Profile — note the reference app's "Dịch vụ" (multi-vertical service switcher: ride/food/delivery) is 🚫 not applicable, PandaDriver is single-vertical |

## 7.2 Current PandaDriver implementation (✅ what exists)

`MapPage` today combines:

- Location resolution state machine (§6.1).
- `GoogleMap` widget, `myLocationEnabled: true`.
- `AvailabilityToggle` widget — calls `AvailabilityRepository.goOnline()`/`goOffline()`, shows a loading spinner mid-toggle, disables re-tap during the transition (`_isToggling` guard — good, prevents double-fire).
- Location upload wiring: going online starts `LocationUploadService.start()` (periodic `POST /api/v1/driver/location`), going offline calls `.stop()`.
- `_fetchAvailability()` on `initState` — restores the online/offline toggle state correctly on app relaunch (important: a driver who force-closes the app mid-shift shouldn't have to remember to manually re-toggle).

There is a **separate, older `HomePage`** (`features/home/presentation/pages/home_page.dart`) with its own `DriverSummaryHeader`, `HomeStatsRow`, `HomeStatusCard`, `QuickActionCard` widgets and a `DriverHomeRepository` (currently backed by `MockDriverHomeSummary` data, not real API). **This page is not wired into `AppRouter` today** — `/` routes to `MapPage`, not `HomePage`. Treat `HomePage` and its supporting widgets as either (a) a design library to harvest good components from (the stats-row and status-card widgets are well-built and directly reusable) or (b) dead code to delete once Wallet/Earnings ship, whichever the team decides. **Do not silently leave two competing "home screen" implementations in the codebase long-term** — this is the kind of drift that causes the next engineer to fix a bug in the wrong file.

## 7.3 Heatmap / demand zones — 🔶 fully proposed, no backend today

This is the single biggest net-new backend dependency in this entire spec. To render "Điểm nóng" (hotspot) or "Nhân giá" (surge) zones, the app needs:

```
GET /api/v1/dispatch/demand-zones?lat={}&lon={}&radius_km={}
→ {
    "zones": [
      { "polygon": [[lat,lon], ...], "demand_level": "low"|"medium"|"high", "surge_multiplier": 1.0 }
    ]
  }
```

This does not exist in the `dispatch` service today (confirmed: gateway router has no such route). **Do not build the map heatmap UI before this backend work is scoped** — it is pure decoration without real data, and a fake/static heatmap actively misleads drivers about where demand is, which is worse than no heatmap at all. Mark this entire visual feature as blocked on backend in the roadmap (§22).

## 7.4 Online / Offline / Searching / Busy states on Home

Reusing the existing `DriverActivityStatus` enum (`offline`, `waitingForTrips`, `searchingNearby`, `busy`) which is already modeled (currently only consumed by the unwired `HomePage`) — this is exactly the right shape for the Home screen's status banner once §7.2's dead-code question is resolved. Recommend: **wire `DriverActivityStatus` into `MapPage`'s new top banner**, driven by:
- `offline` → toggle is off.
- `waitingForTrips` → toggle is on, no active offer, no active trip.
- `searchingNearby` → cosmetic-only distinction from `waitingForTrips` (both are "online, idle") — 🔶 low priority, can be merged into a single "online, idle" state for v1 rather than building fake differentiation.
- `busy` → driver has an active trip (`TripPage` state is not `polling`/`offerAvailable`).

## 7.5 Implementation status summary

✅ Core online/offline mechanics, location streaming, map rendering.
🔶 Wallet-balance shortcut, status banner, heatmap (blocked on backend), quick-actions sheet, recenter FAB.

---

# 8. Driver State Machine

This is the authoritative state machine for a driver's shift. It spans both the "ambient" availability state (§7.4/`DriverAvailabilityStatus`) and the "per-trip" state (`TripPage`'s internal `_PageState`, §9). They are two independent-but-coupled state machines; conflating them has caused confusion in the existing code (two enums — `DriverActivityStatus` and `DriverAvailabilityStatus` — model overlapping concepts). This spec formalizes the merge.

## 8.1 States

| State | Meaning | Entered from | Exits to |
|---|---|---|---|
| `Offline` | Driver not receiving offers, not uploading location | initial state; `GoOffline` action from any online state | `Online` (via `GoOnline`) |
| `Online` (= `waitingForTrips`) | Uploading location, polling for offers, no offer/trip active | `Offline` via `GoOnline`; `Completed`/`Rated` via auto-return | `Searching`, `Offline`, `Assigned` |
| `Searching` | Cosmetic sub-state of `Online` — optional, see §7.4 | `Online` | `Online`, `Assigned` |
| `Assigned` (= `offerAvailable`) | An offer is presented with a countdown | `Online`/`Searching` when dispatch sends an offer | `Accepted`, `Online` (on reject/expire) |
| `Accepted` (= `acting`→`activeTrip`, status `driver_assigned`) | Driver accepted, en route to pickup | `Assigned` via Accept | `Arriving`... (same state, sub-labeled by UI) |
| `Arriving` | UI label for `activeTrip` before `hasArrived` flips true | `Accepted` | `Waiting`/`In Trip` |
| `Waiting` (= `activeTrip`, `hasArrived: true`, status `driver_arrived`) | Driver at pickup, rider not yet in vehicle | `Arriving` via "Tôi đã đến điểm đón" | `In Trip` |
| `In Trip` (= `activeTrip`, status `in_progress`) | Trip underway | `Waiting` via "Bắt đầu chuyến đi" | `Completed` |
| `Completed` (= `awaitingPayment` → `completed`) | Trip finished, awaiting/confirmed rider payment | `In Trip` via "Kết thúc chuyến đi" | `Rated` |
| `Rated` | Driver has rated the rider (or skipped) | `Completed` via rating submit/skip | `Online` (automatic) |

## 8.2 Transition diagram

```
                    ┌─────────────────────────────────────────────┐
                    │                                               │
        GoOffline   ▼                                               │ auto after
   ┌──────────► Offline                                             │ rating/skip
   │                │                                                │
   │        GoOnline│                                                │
   │                ▼                                                │
   │             Online ◄──────────────────────────────┐             │
   │             (waiting/searching)                    │             │
   │                │                                    │             │
   │      dispatch offers a trip                  reject/expire       │
   │                ▼                                    │             │
   │             Assigned (countdown) ───────────────────┘             │
   │                │                                                  │
   │            Accept│                                                │
   │                ▼                                                  │
   │             Accepted/Arriving ────────► Waiting ────► In Trip     │
   │                                                              │     │
   │                                                     Finish trip│     │
   │                                                              ▼     │
   │                                                        Completed   │
   │                                                              │     │
   │                                                    Rate/Skip │     │
   │                                                              ▼     │
   │                                                          Rated ────┘
   │
   └── GoOffline can be invoked from Online only in current implementation.
       See §8.3 for the "GoOffline while Assigned/Accepted/In Trip" question.
```

## 8.3 Invalid transitions and edge cases

| Attempted transition | Current behavior | Recommended behavior |
|---|---|---|
| `GoOffline` while an offer is `Assigned` (countdown active) | 🔶 undefined — `AvailabilityToggle` is only rendered/reachable from `MapPage`, not from the offer card in `TripPage`. Effectively **impossible today** since the two screens don't cross-link the toggle. | Once Home and Trip surfaces are consolidated (§22), explicitly disable the offline toggle while `_PageState` is anything other than `polling`/`offerAvailable`-with-reject-available. If a driver *must* go offline mid-offer, treat it as an implicit reject. |
| `GoOffline` while `Accepted`/`Arriving`/`Waiting`/`In Trip` | Should be **blocked entirely** — a driver cannot abandon a rider mid-trip by toggling offline. | Backend should reject `POST /api/v1/driver/go-offline` with 409 if the driver has an active trip; client should not even show the control (remove/hide, don't just disable, to avoid a confusing disabled-button state). Verify current backend behavior — if unenforced server-side, this is a 🔶 backend hardening item, related to the H3/H4 saga-compensation hardening work already done for dispatch (per recent commit history: "saga compensation, idempotency, and dispatch engine lifecycle"). |
| Offer countdown reaches 0 with no driver action | ✅ handled — `_startCountdown` auto-clears the offer and returns to `polling` | — |
| App killed mid-`Accepted`/`In Trip` and reopened | ✅ handled — `TripPage._initialize()` calls `getStoredTripId()` then `fetchTrip()` and restores the correct `_PageState` from server-reported `status` | — |
| Two devices logged in as the same driver simultaneously | 🔶 undefined — no session-single-device enforcement visible in `AuthHandler`. Ride-hailing driver apps universally prevent this (double-booking risk: same driver "accepts" two offers from two devices). Recommend backend enforcement: new login invalidates prior access token, or dispatch treats `driver_id` uniqueness as a hard constraint at offer-assignment time (may already be implicitly enforced by dispatch's driver-lock semantics — verify against dispatch service code before treating as a gap). |
| Rider cancels while driver is `Arriving`/`Waiting` | ✅ backend has a `POST /api/v1/rides/{tripID}/cancel` and saga compensation (per recent hardening work) — client must poll/observe this and transition the driver back to `Online` with a clear "Chuyến đi đã bị hủy bởi khách hàng" message. Verify `TripPage._resumeActiveTrip`/`_paymentPoll` cover the cancelled-by-rider case, not just the settled-payment case — this looks like a gap: `_paymentPoll` only checks `updated.status == 'settled'`, doesn't appear to branch on `cancelled`. 🔶 flag as a bug to verify/fix. |

---

# 9. Trip Lifecycle UI

**Route:** `/trips` → `TripPage`. This is the most complex and most-tested part of the app (per project history: "trips D-03 through D-06 (offer, accept, navigation, arrival)" already shipped).

## 9.1 State-by-state screen spec

### 9.1.1 Incoming request (Offer)

- **Trigger:** `_poll()` (every 5s while in a pollable state) finds a non-null offer from `GET /api/v1/driver/current-offer`.
- **Widgets:** `_OfferCard` — title "Yêu cầu chuyến mới", `_CountdownBadge` (color flips to error/red when ≤10s remaining), pickup/drop-off `_AddressRow`s, two `_InfoChip`s for distance and fare (currently placeholder `—` — 🔶 backend gap: the offer payload doesn't appear to carry a fare estimate or distance-to-pickup number that's actually populated client-side; verify `TripOffer` model and wire real values, since showing "—" for the fare estimate on the single most important decision-making screen (accept or reject this trip?) undermines driver trust).
- **Actions:** Accept (`_onAccept`) / Reject (`_onReject`), both disable further interaction during the in-flight request (`_PageState.acting`).
- **Countdown:** client-side timer synced against `offer.offerExpiresAt` (server-provided expiry, not just a client-side guess) — ✅ correct pattern, avoids clock-drift accept-after-expiry bugs.

**🔶 Missing: audible/haptic alert.** A driver is not staring at the phone at all times. A new offer needs a sound + vibration, not just a silent UI change — this is table-stakes for every ride-hailing driver app and is currently absent (no `vibration`/`audioplayers` package in `pubspec.yaml`). High-priority gap, see §22 Phase 3.

### 9.1.2 Accept

- `POST /api/v1/rides/{tripID}/accept` — ✅ implemented, atomic transaction per project history ("atomic transactions for AcceptTrip").
- On success: `_activeTrip` constructed client-side with `status: 'driver_assigned'` optimistically (before the next poll confirms) — ✅ good optimistic-update pattern, avoids a blank loading flash.
- On failure (e.g. offer already taken by another driver / expired server-side): `_PageState.error`, `_errorMessage` shown with retry → `_initialize()`. 🔶 Recommend a more specific message than the generic `e.message` for this specific case ("Chuyến đi này đã được nhận bởi tài xế khác" is more actionable than a raw backend error string), i.e. map known `statusCode`/error codes to friendly Vietnamese copy at the client boundary.

### 9.1.3 Reject

- `POST /api/v1/rides/{tripID}/reject` — ✅ implemented.
- Returns driver to `_startPolling()` immediately, no confirmation dialog (correct — rejecting should be low-friction, drivers reject offers constantly for legitimate reasons: wrong direction, low fare, etc.)
- 🔶 Consider (backend, not urgent): track rejection rate for the "Tỉ lệ hoạt động" profile metric seen in the reference app (§12), but do **not** show a driver-facing penalty/warning for rejecting — that's an anti-pattern that damages driver trust and is explicitly not adopted here even though some competitor apps do it.

### 9.1.4 Navigation (to pickup)

- **Current implementation:** address rows + a fare/status card; **no turn-by-turn navigation embedded in-app.** The reference app screenshots don't show in-app turn-by-turn either (typical pattern: driver apps deep-link into Google Maps/Waze for actual navigation rather than reimplementing it).
- 🔶 Proposed: add a "Chỉ đường" button that launches `google.navigation:q=lat,lon&mode=d` via `url_launcher` (not yet a dependency — needs adding to `pubspec.yaml`). This is a small, high-value addition — do not attempt to build in-app turn-by-turn navigation, that is a multi-month undertaking with no product justification when Google Maps deep-linking solves it in an afternoon.

### 9.1.5 Arrival

- `_onArrived()` → `POST /api/v1/rides/{tripID}/arrive` — ✅ implemented.
- Status banner switches to "Đã đến điểm đón" (✅ localized), action button switches to "Bắt đầu chuyến đi".
- **No waiting-timer/waiting-fee UI** in the live `TripPage` today, even though `features/trips/presentation/widgets/waiting_timer.dart` and `waiting_fee_card.dart` **already exist** in the legacy/preview module. 🔶 High-value, low-effort: promote these two widgets from `features/trips` into `features/trip` and wire a simple client-side timer starting at `arrive` — this directly maps to the reference app's implied "waiting time affects earnings" concept and is nearly free since the widgets are already built, just not connected to the live page.

### 9.1.6 OTP (pickup verification)

- 🚫 Not present in current backend or UI. Some ride-hailing apps require the rider to read a PIN to the driver before trip start, as an anti-fraud/wrong-pickup measure. **Not recommending for v1** — adds friction with no corresponding backend support (`ArriveAtPickup`/`StartTrip` endpoints take no OTP parameter) and is not evidenced in the reference screenshots either. Listed in §21 (Future Features) as a fraud-prevention option only if trip-hijacking becomes a measured problem.

### 9.1.7 Start

- `_onStartTrip()` → `POST /api/v1/rides/{tripID}/start` — ✅ implemented. Also starts `TripMetricsEngine` (client-side GPS-based distance/duration accumulation, used at Finish to report actual trip metrics — smart design, avoids relying solely on server-side estimation which could differ from actual route taken).

### 9.1.8 Finish

- `_onFinishTrip()` → `POST /api/v1/rides/{tripID}/finish` with `distanceKm`/`durationMin` from the metrics engine — ✅ implemented. Metrics snapshot is memoized (`_finalMetrics ??= ...`) so a failed request + retry doesn't double-count — ✅ correct idempotency-conscious client design.
- Transitions to `awaitingPayment`, starts `_paymentPollTimer` (3s interval) waiting for `status == 'settled'`.

### 9.1.9 Payment

- Driver sees a "Đang chờ thanh toán" spinner card with the fare and route recap — ✅ implemented.
- **Gap identified in §8.3:** poll loop doesn't appear to branch on a rider-side cancellation/failure during this window — needs verification/fix.
- 🔶 No visibility into *which* payment method the rider is using (cash vs. wallet vs. card) — for a cash trip, the driver is physically waiting for cash in hand, not a payment confirmation ping; the current UI doesn't distinguish this, which could leave a driver confused about what "waiting for payment" even means for a cash trip. Recommend surfacing `paymentMethod` on the awaiting-payment card and adjusting copy: cash → "Xác nhận đã nhận tiền mặt" (driver-initiated confirmation) vs. digital → "Đang chờ khách thanh toán" (passive wait).

### 9.1.10 Rating

- `_TripCompletedCard` — 5-star selector + optional comment (200 char max), `POST /api/v1/rides/{tripID}/rate` — ✅ implemented, non-fatal on failure (driver can skip/proceed regardless, correct choice — don't block a driver's next earning opportunity on a rating submission succeeding).
- Auto-returns to `_startPolling()` on submit or skip.

## 9.2 Implementation status summary

✅ Full offer → accept → navigate → arrive → start → finish → payment-wait → rate loop, with good resilience (app-kill recovery, optimistic updates, idempotent metrics).
🔶 Gaps: audible offer alert, real fare/distance estimate on the offer card, waiting-timer wiring (widgets exist, just disconnected), navigation deep-link, rider-cancellation handling during payment wait, payment-method-aware copy, cash-received confirmation.

---

# 10. Wallet Module

**Fully 🔶 proposed — no backend wallet service exists today.** This section specifies both the UI and the minimum backend contract needed to build it, since the reference screenshots make clear this is expected functionality (screenshot 2: balance, linked bank account, pending/withdrawable split, withdraw CTA, footnote rules).

## 10.1 Reverse-engineered business logic (from screenshot 2)

- **"Số dư tạm tính" (provisional balance)** — the operative word is *tạm tính* (provisional/estimated). This implies the platform holds funds in a temporary state before they're fully settled/withdrawable — standard ride-hailing pattern to handle payment-processing float and dispute windows.
- **Linked external account ("Tài khoản Cake liên kết")** — the reference app integrates with a specific e-wallet/bank partner for payouts. PandaDriver should not hard-couple to any single third-party wallet brand; model this generically as a **linked payout method** (bank account or e-wallet), configurable, not vendor-locked.
- **"Tiền đang chờ duyệt" (pending approval)** — funds not yet cleared (likely: trips completed in the last N hours, still in a fraud/dispute review window, or awaiting daily settlement batch).
- **"Có thể rút" (withdrawable)** — the actual liquid balance.
- **Two footnote rules:**
  - A minimum balance (50,000₫ in the screenshot) is held back as an account-maintenance reserve, not withdrawable.
  - A daily withdrawal cap (3,000,000₫/day in the screenshot).
  These are **business/finance-team decisions**, not engineering decisions — PandaDriver's actual numbers must come from Finance/Ops, not be copied from the reference screenshot. The UI must support displaying arbitrary rule text/numbers (don't hardcode "50.000đ" and "3.000.000đ" into the Dart source — these must be server-configured values, likely from a `wallet_config` table, so Ops can tune them without an app release).

## 10.2 Proposed data model

```dart
class WalletSummary {
  final int provisionalBalanceCents;
  final int pendingApprovalCents;
  final int withdrawableCents;
  final String currency;                    // "VND"
  final int minimumReserveCents;             // server-configured, e.g. 50_000 * 100
  final int dailyWithdrawLimitCents;         // server-configured
  final LinkedPayoutAccount? linkedAccount;  // null = not yet linked
}

class LinkedPayoutAccount {
  final String provider;        // e.g. "bank_transfer", "e_wallet"
  final String displayName;     // masked account name/number for display
  final bool isVerified;
}

class WalletTransaction {
  final String id;
  final DateTime timestamp;
  final int amountCents;        // signed: positive = credit, negative = debit/withdrawal
  final WalletTransactionType type;   // tripEarning | withdrawal | adjustment | bonus
  final String? relatedTripId;
}
```

## 10.3 Proposed backend endpoints

```
GET  /api/v1/driver/wallet                          → WalletSummary
GET  /api/v1/driver/wallet/transactions?cursor=&limit=  → paginated WalletTransaction[]
POST /api/v1/driver/wallet/withdraw   { amount_cents }  → { withdrawal_id, status }
GET  /api/v1/driver/wallet/payout-account            → LinkedPayoutAccount | 404
POST /api/v1/driver/wallet/payout-account            { provider, account_details }  → link/update
```

None of these exist in the current gateway router. This is a **new backend service or a significant extension of an existing one** (likely lives alongside `booking`/`pricing` since it's derived from settled trip fares, or as a new `wallet` service if payout/withdrawal has its own compliance/audit requirements — that's an architecture decision for the backend team, out of scope here beyond flagging the dependency).

## 10.4 Screens

### 10.4.1 Wallet Page (`/wallet`)

- **Purpose:** mirror reference screenshot 2 — big balance card, linked-account row, pending/withdrawable rows, withdraw CTA, footnote rules (server-driven text).
- **Widgets:** `WalletBalanceCard` (new, reusable — large amount display + CTA, same visual weight as `FareSummaryCard` in the rider app), plain `ListTile` rows for linked account / pending / withdrawable, footnote `Text`.
- **Actions:** "Nạp tiền" (top-up — 🔶 lower priority than withdrawal; a driver wallet is usually earnings-in, not top-up-in, unless it's used to pay platform fees/penalties directly — clarify with product before building; may be 🚫 not applicable for a pure-earnings wallet), "Rút tiền" → `/wallet/withdraw`, "Lịch sử" → `/wallet/history`.
- **State:** `AsyncStateView<WalletSummary>`.
- **API:** `GET /api/v1/driver/wallet`.
- **Error handling:** standard retry via `AsyncStateView`'s error builder.
- **Skeleton:** shimmer placeholder for the balance card (amount is the single most-anticipated number on this screen — a blank/spinner-only loading state feels slower than it is; a shimmer numeral placeholder is worth the extra widget).
- **Empty state:** n/a (summary always has a value, even if zero).
- **Refresh:** pull-to-refresh; also re-fetch on return-to-foreground after a trip completes (a driver who just finished a trip and taps into Wallet expects to see the new earning reflected — don't serve a stale cached value).

### 10.4.2 Withdraw Page (`/wallet/withdraw`)

- **Purpose:** enter an amount (≤ withdrawable balance, ≤ daily limit), confirm, submit.
- **Widgets:** amount input (numeric, VND formatting, live validation against both caps with inline error text), destination account display (read-only, from `linkedAccount`; if none linked, redirect to a link-account flow first), confirm button.
- **Actions:** submit → `POST /api/v1/driver/wallet/withdraw`.
- **Error handling:** insufficient balance / over daily limit / no linked account — all as inline validation *before* submit where possible (client-side pre-check using the already-fetched `WalletSummary`), with server-side re-validation as the source of truth (client validation is UX sugar, not a security boundary).
- **Success:** confirmation state + return to Wallet page with an updated summary (re-fetch, don't just optimistically decrement — withdrawal processing may not be instant, and `provisionalBalanceCents` vs `withdrawableCents` accounting should come from the server).

### 10.4.3 Wallet History Page (`/wallet/history`)

- **Purpose:** flat, paginated list of `WalletTransaction`s (all types: earnings, withdrawals, adjustments, bonuses) — broader than Earnings' per-trip view, this is the full ledger.
- **Widgets:** `ListView.separated`, signed-amount styling (green credit / red-or-neutral debit), tap-through to Earnings Detail for `tripEarning`-type rows (reuse §11.4's detail page, keyed by `relatedTripId`).
- **API:** `GET /api/v1/driver/wallet/transactions` with cursor pagination.
- **Empty state:** "Chưa có giao dịch nào."

## 10.5 Implementation status summary

🔶 Entirely new — UI and backend. **Do not start Flutter work on this module until the backend wallet contract (§10.3) is confirmed with the backend team** — building against a guessed API shape that later changes is wasted work; this section exists to drive that backend conversation, not to be implemented blind.

---

# 11. Earnings Module

Distinct from Wallet: **Earnings is about understanding how a number was calculated; Wallet is about moving that money.** The reference app (screenshots 6, 7) demonstrates this split clearly — the earnings screen groups by period and lets you drill into a single trip's fee breakdown; it does not handle withdrawal at all (that's the separate Account/Wallet screen).

## 11.1 Daily / Weekly / Monthly summary (screenshot 6)

- **Segmented control:** Ngày / Tuần / Tháng — ✅ directly adoptable pattern, simple `SegmentedButton` or three `ChoiceChip`s.
- **Period navigator:** `< THÁNG 10/2025 >` — arrows to page between periods, current period highlighted.
- **Big total number**, with a parenthetical clarifier ("Thu nhập từ đơn hàng và các khoản thanh toán khác") — good pattern: always caption what a big number includes, don't leave it ambiguous.
- **Per-vertical breakdown** (beRide/beDelivery/beFood counts) — 🚫 not applicable, PandaDriver is single-vertical (ride only). Replace with a more useful single-vertical breakdown: **completed trips count + cancelled trips count + average fare**, which is the equivalent "at a glance"信息 for a ride-only driver.
- **"Lịch sử" section** — flat list of trips within the selected period, timestamp + amount + chevron into detail.

## 11.2 Proposed data model

```dart
enum EarningsPeriod { day, week, month }

class EarningsSummary {
  final EarningsPeriod period;
  final DateTime periodStart;
  final int totalCents;
  final int completedTripCount;
  final int cancelledTripCount;
  final int averageFareCents;
  final List<EarningsEntry> entries;   // trips within this period, newest first
}

class EarningsEntry {
  final String tripId;
  final DateTime timestamp;
  final int amountCents;
  final String pickupAddress;   // for the list row preview, if needed
}
```

## 11.3 Proposed backend endpoints

```
GET /api/v1/driver/earnings?period=day|week|month&date=YYYY-MM-DD  → EarningsSummary
GET /api/v1/driver/earnings/{tripId}                                → EarningsDetail (see §11.4)
```

**Note:** this may be derivable *entirely client-side* by aggregating `GET /api/v1/driver/trips` (already ✅ implemented) if that endpoint's response includes `final_fare`/`currency`/`created_at`/`status` per trip (it does, per `_TripSummary.fromJson` in the existing `DriverTripHistoryPage`). **Recommendation: do not build a new `/earnings` backend endpoint for the summary view initially** — reuse `GET /api/v1/driver/trips` and aggregate client-side (filter by date range, sum `final_fare`). This avoids new backend work for something the data already supports. Only build a dedicated endpoint if trip volume grows large enough that client-side aggregation over the full trip list becomes a performance problem (add pagination/date-range query params to the existing trips endpoint first, as a smaller incremental change, before standing up a parallel earnings-specific endpoint).

The **detail** view (§11.4) genuinely needs new backend data (fee/tax breakdown), since that is not part of the trip summary today.

## 11.4 Earnings Detail Page (screenshot 7)

- **Purpose:** full fee transparency for a single trip — this is the trust-building screen.
- **Reverse-engineered structure:**
  ```
  Mã chuyến đi: {trip_id}          [vehicle icon] {vehicle type label}
  {timestamp}

  Thanh toán cho đối tác (I)+(II)                    {total, bold, green}
  ────────────────────────────────────
  Thu nhập từ đơn hàng (I)                            {subtotal}
    Cước phí đơn hàng trước thuế GTGT                 {gross fare}
    Phí SDUD và phí dịch vụ (ⓘ)                       -{platform commission}
    Thuế TNCN                                          -{personal income tax withheld}
  ────────────────────────────────────
  Các khoản thanh toán khác (II)                      {bonuses/adjustments, 0 if none}
  ────────────────────────────────────
  Khách hàng thanh toán  ▾                             {collapsible: what the rider actually paid, may differ from driver payout due to promos}

  [note] Gói tiết kiệm khách mua kèm (nếu có) sẽ không ảnh hưởng đến thu nhập của Đối tác.

  {pickup pin} {pickup address}
  {dropoff pin} {dropoff address}
  ```
- **Key UX principle reverse-engineered here:** the driver's payout is shown as the *primary* number, and it's explicitly decoupled from "what the rider paid" (which is secondary/collapsed) — a driver should never have to do their own math to figure out "why is my payout less than the fare the app quoted the rider." The line items (gross fare → minus commission → minus tax → equals payout) make the math traceable at every step. **This is the single most important pattern to carry over from the reference app, more so than any visual element.**
- **"(ⓘ)" info icon** next to the commission line → tappable, opens an explainer (bottom sheet or dialog) of what "Phí SDUD và phí dịch vụ" (service/platform usage fee) actually covers. PandaDriver equivalent: a short, honest explanation of the platform commission — do not hide or obscure this, transparency here is a trust feature, not a nice-to-have.

## 11.5 Proposed data model (detail)

```dart
class EarningsDetail {
  final String tripId;
  final DateTime timestamp;
  final String vehicleTypeLabel;
  final int payoutTotalCents;              // (I) + (II)
  final int tripEarningCents;              // (I)
  final int grossFareCents;                // before deductions
  final int platformFeeCents;              // negative in UI
  final int personalIncomeTaxCents;        // negative in UI
  final int otherPaymentsCents;            // (II) — bonuses, adjustments
  final int riderPaidCents;                // what the rider was charged (collapsible section)
  final String pickupAddress;
  final String dropoffAddress;
}
```

## 11.6 Filtering

- Period tabs (day/week/month) double as the primary filter — no separate filter UI needed, matches reference app's simplicity. 🔶 Consider adding a status filter (All / Completed / Cancelled) to the history list within a period, reusing the existing `ChoiceChip` pattern already built for `DriverTripHistoryPage`'s "Completed" filter (`find.widgetWithText(ChoiceChip, 'Completed')` — confirmed this pattern already exists and works well there; extend it here rather than inventing a new filter widget).

## 11.7 Charts

Reference screenshots show no chart — just numbers and lists. **Do not add a chart for v1** — a bar/line chart of daily earnings is a plausible v2 addition (§21) but is not evidenced as necessary by the reference material and adds a charting-library dependency (none exists in `pubspec.yaml` today) for unproven value. Ship the numeric version first; add visualization only if drivers ask for it.

## 11.8 Implementation status summary

🔶 Summary view: buildable now by extending the existing `GET /api/v1/driver/trips` endpoint client-side — no backend blocker.
🔶 Detail view: blocked on backend exposing the fee/tax breakdown per trip (currently the trip model only has `final_fare`, not a decomposition) — needs `pricing`/`booking` service work to persist and expose the commission/tax split at trip-settlement time.

---

# 12. Driver Profile

**Route:** `/profile`. Reverse-engineered from screenshot 8, re-scoped for PandaDriver.

## 12.1 Header

- Avatar, name, star rating — ✅ exists in current `ProfilePage` in some form; verify it's driven by real `GET /api/v1/drivers/{driverID}/profile` data (✅ endpoint exists) rather than mock data.
- QR code icon (reference app) — likely a driver-identification QR for in-person verification (e.g. a rider or ops staff scanning to confirm driver identity). 🔶 Low priority, only build if there's a confirmed use case (in-person driver verification at a hub, fraud-check kiosk, etc.) — do not build a QR code with no defined scanning consumer on the other end.

## 12.2 Menu sections (grouped, matching reference app's visual grouping via dividers)

**Group 1 — Performance & standing:**
| Item | Purpose | Status |
|---|---|---|
| Đơn hàng cần hoàn trả (Orders needing return) | 🚫 not applicable — this is a delivery/food-vertical concept (returning undeliverable goods), irrelevant to ride-hailing | 🚫 |
| Tỉ lệ hoạt động (Activity rate) | Likely: acceptance rate, cancellation rate, online-hours-vs-target — a self-monitoring dashboard so drivers can see what affects their standing | 🔶 needs backend aggregation of accept/reject/cancel history |
| Chương trình thưởng (Rewards program) | Incentive program (e.g. "complete 20 trips this week for a bonus") | 🔶 needs a promotions/incentives backend, likely tied to the same system that would generate Wallet's "bonus" transaction type (§10.2) |
| Thành tích hoạt động (Activity achievements) | Gamified milestones (e.g. "1000 trips completed") | 🔶 low priority, see §21 |
| Đánh giá từ khách hàng (Ratings from riders) | List of individual rider ratings/comments, not just the aggregate star | ✅ backend has `GET /api/v1/rides/{tripID}/rating` per-trip — 🔶 needs a new aggregate list endpoint (`GET /api/v1/driver/ratings?cursor=`) to list them, rather than fetching one-by-one per trip |

**Group 2 — History & growth:**
| Item | Purpose | Status |
|---|---|---|
| Lịch sử (History) | Same as existing `DriverTripHistoryPage` | ✅ already exists, just needs a menu entry point if not already linked |
| Giới thiệu & Nhận ưu đãi (Referral) | Refer-a-driver program with a shareable code/link | 🔶 see §21 |
| Ưu đãi mua xe máy điện (EV purchase incentive) | Vehicle-type-specific partnership offer | 🚫 not applicable unless PandaDriver has a similar partnership — do not build speculative partner-integration UI |

**Group 3 — Learning & support:**
| Item | Purpose | Status |
|---|---|---|
| Sổ tay tài xế (Driver Handbook) | Static content — policies, best practices | 🔶 easy win: a static in-app content page (could even be a `WebView` pointing at a CMS-hosted page, avoids needing app releases to update policy text) |
| Học viện đào tạo (Training academy) | Structured training/onboarding content, possibly video | 🔶 low priority, defer — significant content-production dependency outside engineering scope |
| Hỗ trợ (Support) | Help center / contact support | 🔶 minimum viable version: a support-ticket form or a deep-link to a support phone/chat channel — do not attempt to build a full in-app help-center CMS for v1, link out to a web-based help center instead |
| Kiểm tra tổng quát (General/vehicle check) | Likely a periodic vehicle safety self-checklist | 🚫 not applicable for v1 unless PandaDriver has a compliance requirement mandating it — flag to product/legal, don't build speculatively |
| Cài đặt mật khẩu (Password settings) | 🚫 not applicable — passwordless auth (§5) |

## 12.3 PandaDriver's actual v1 Profile menu (synthesized, not copied)

Given the 🚫/🔶 triage above, the realistic v1 menu is:

1. Header: avatar, name, rating (✅ mostly exists)
2. **Lịch sử chuyến đi** → `DriverTripHistoryPage` (✅ exists)
3. **Ví & thu nhập** → shortcut into Wallet (§10) — avoid duplicating the bottom-nav Earnings tab, this entry is specifically the Wallet/payout shortcut
4. **Thông tin xe** (Vehicle Info) → 🔶 new, simple read-only display of vehicle make/model/plate (data likely already exists server-side per driver record, just needs a display screen — check `driver` service's existing driver profile fields before assuming new backend work)
5. **Giấy tờ** (Documents) → 🔶 new, license/registration document status (view-only for v1; upload/renewal flow is a bigger scope, defer)
6. **Đánh giá từ khách hàng** → 🔶 new, per §12.2 Group 1
7. **Sổ tay tài xế** → 🔶 new, static/WebView content
8. **Hỗ trợ** → 🔶 new, minimum-viable contact/ticket form
9. **Cài đặt** (Settings: notification preferences, language — though app is Vietnamese-only today) → 🔶 new, start minimal (notification toggle only)
10. **Đăng xuất** (Logout, with confirmation per §3.1) → 🔶 wire up if not already present

Deferred to §21 (Future Features), not in v1: Activity rate dashboard, Rewards program, Achievements, Referral, Training academy, QR identification.

## 12.4 Implementation status summary

🔶 Almost entirely new beyond the existing header + trip-history link. Prioritize Wallet shortcut, Vehicle Info, and Support first (§22) — these are the highest driver-value, lowest-backend-dependency items in this section.

---

# 13. Notification System

**Route:** `/notifications`, currently a stub tab.

## 13.1 Two distinct notification concepts (do not conflate)

1. **Time-critical, in-app-only signal: new trip offer.** This is not a "notification" in the inbox sense — it's a real-time state change in `TripPage` (§9.1.1), delivered via polling today. This should **never** be listed in the Notifications tab as a historical item (a driver doesn't want to scroll past 40 "you had a trip offer" entries) — it's ambient/ephemeral UI, not a persisted message.
2. **Everything else** (the reference app's "Hộp thư" with a "99+" badge): account notices, promotions, policy announcements, support replies. **This** is what `/notifications` should show.

## 13.2 Notification types

```dart
enum DriverNotificationType { system, promotion, earnings, support, tripSummary }

class DriverNotification {
  final String id;
  final DriverNotificationType type;
  final String title;
  final String body;
  final DateTime timestamp;
  final bool isRead;
  final String? deepLinkRoute;   // e.g. "/wallet", "/earnings/{tripId}"
}
```

This directly mirrors the rider app's already-built `NotificationItem`/`MockNotificationRepository` pattern (`apps/rider/lib/features/profile/domain/models/notification_item.dart`) — **reuse that exact shape and the `AsyncStateView`-based `NotificationCenterPage` UI pattern** rather than designing a new one from scratch. The two apps should feel like siblings.

## 13.3 Push delivery — 🔶 fully proposed

No push infrastructure (FCM/APNs) exists in either app today. This is a meaningful new dependency:

- `firebase_messaging` (or a self-hosted push gateway, architecture decision for backend) added to `pubspec.yaml`.
- Backend needs a device-token registration endpoint: `POST /api/v1/driver/push-token { token, platform }`.
- Backend needs to actually send push payloads on relevant events (new promo, support reply, etc. — **not** new trip offers, which stay on the polling/in-app model per §13.1, since polling every 5s while online is already near-real-time and push adds complexity for marginal latency gain on the one thing that's already fast).

## 13.4 Unread badge

Reference app shows "99+" on the bottom-nav tab. Reuse the rider app's `UnreadBadge` widget pattern (`apps/rider/lib/features/profile/presentation/widgets/unread_badge.dart`) verbatim — same component, same visual language, just wired to the driver notification count instead.

## 13.5 Screen spec

- **List page:** `ListView.separated` of `NotificationTile`s, tap → mark read + deep-link (if `deepLinkRoute` present) or expand in-place.
- **API:** `GET /api/v1/driver/notifications?cursor=` (paginated), `POST /api/v1/driver/notifications/{id}/read`, `POST /api/v1/driver/notifications/mark-all-read` — all 🔶 new, mirroring rider app's notification endpoints if those already exist server-side (check — if the rider notification center is also still mock-only per earlier project state, this is a shared backend gap for both apps, worth building once and reusing the same service for both).
- **Empty state:** "Chưa có thông báo nào." (exact copy already established in rider app — reuse verbatim for consistency).
- **Error/retry:** standard `AsyncStateView` pattern.

## 13.6 Implementation status summary

🔶 Entirely new. Recommend building the **list UI first against a mock repository** (exactly like the rider app's `MockNotificationRepository` pattern) so the screen is demoable and testable before push infra and backend persistence exist — this unblocks Flutter work immediately without waiting on FCM setup or a new backend service.

---

# 14. API Mapping

One row per screen/feature, endpoint, and status. This is the authoritative cross-reference — if a screen needs an endpoint not listed here, add it here first before writing Dart code against it.

| Screen / Feature | Endpoint | Status |
|---|---|---|
| Login | `POST /api/v1/auth/login` | ✅ |
| Session refresh | `POST /api/v1/auth/refresh` | 🔶 proposed (§5.1) |
| Home — go online | `POST /api/v1/driver/go-online` | ✅ |
| Home — go offline | `POST /api/v1/driver/go-offline` | ✅ |
| Home — restore toggle state | `GET /api/v1/driver/availability` | ✅ |
| Location upload (while online) | `POST /api/v1/driver/location` | ✅ |
| Rider reads driver location (not this app, noted for completeness) | `GET /api/v1/driver/{driverID}/location` | ✅ |
| Driver profile header | `GET /api/v1/drivers/{driverID}/profile` | ✅ |
| Trip offer polling | `GET /api/v1/driver/current-offer` | ✅ |
| Trip history list | `GET /api/v1/driver/trips` | ✅ |
| Accept offer | `POST /api/v1/rides/{tripID}/accept` | ✅ |
| Reject offer | `POST /api/v1/rides/{tripID}/reject` | ✅ |
| Arrive at pickup | `POST /api/v1/rides/{tripID}/arrive` | ✅ |
| Start trip | `POST /api/v1/rides/{tripID}/start` | ✅ |
| Finish trip | `POST /api/v1/rides/{tripID}/finish` | ✅ |
| Trip payment status (polled) | `GET /api/v1/rides/{tripID}` | ✅ |
| Rate rider | `POST /api/v1/rides/{tripID}/rate` | ✅ |
| Get a specific rating | `GET /api/v1/rides/{tripID}/rating` | ✅ |
| Aggregate ratings list (Profile) | `GET /api/v1/driver/ratings` | 🔶 (§12.2) |
| Wallet summary | `GET /api/v1/driver/wallet` | 🔶 (§10.3) |
| Wallet transaction history | `GET /api/v1/driver/wallet/transactions` | 🔶 (§10.3) |
| Withdraw funds | `POST /api/v1/driver/wallet/withdraw` | 🔶 (§10.3) |
| Get/set linked payout account | `GET`/`POST /api/v1/driver/wallet/payout-account` | 🔶 (§10.3) |
| Earnings summary (day/week/month) | derive client-side from `GET /api/v1/driver/trips` | ✅ derivable, no new endpoint needed (§11.3) |
| Earnings detail (fee/tax breakdown) | `GET /api/v1/driver/earnings/{tripId}` | 🔶 (§11.3) |
| Demand/heatmap zones | `GET /api/v1/dispatch/demand-zones` | 🔶 (§7.3) |
| Notifications list | `GET /api/v1/driver/notifications` | 🔶 (§13.5) |
| Mark notification read | `POST /api/v1/driver/notifications/{id}/read` | 🔶 (§13.5) |
| Push token registration | `POST /api/v1/driver/push-token` | 🔶 (§13.3) |
| Activity rate metrics | `GET /api/v1/driver/activity-rate` | 🔶 (§12.2) |

**Explicitly not inventing:** dispatch's internal gRPC contracts, the pricing engine's fare calculation logic, the identity service's JWT internals — these already exist and work (confirmed by the completed end-to-end backend flow); this spec only names the *HTTP-gateway-facing* surface a driver client would call.

---

# 15. Data Models

Canonical Dart model shapes for every domain concept this spec touches — either already implemented (✅, referencing the real file) or proposed (🔶, new file path given).

## 15.1 Driver — 🔶 proposed (currently only a `driverId` string is threaded around; no first-class `Driver` model exists client-side)

```dart
// lib/features/profile/domain/models/driver.dart
class Driver {
  final String id;
  final String fullName;
  final String phoneNumber;
  final double rating;
  final int totalCompletedTrips;
  final String? avatarUrl;
}
```

## 15.2 Vehicle — 🔶 proposed

```dart
// lib/features/profile/domain/models/vehicle.dart
class Vehicle {
  final String category;      // matches rider app's VehicleCategory: car | motorcycle | van
  final String brand;
  final String model;
  final String plateNumber;
  final String? colorHex;
}
```

## 15.3 Wallet — 🔶 proposed
See §10.2 (`WalletSummary`, `LinkedPayoutAccount`, `WalletTransaction`).

## 15.4 Income/Earnings — 🔶 proposed
See §11.2 and §11.5 (`EarningsSummary`, `EarningsEntry`, `EarningsDetail`).

## 15.5 Trip — ✅ implemented, split across two live models

```dart
// lib/features/trip/data/trip_offer_repository.dart (TripOffer) — ✅
class TripOffer {
  final String tripId;
  final String pickupAddress;
  final String dropoffAddress;
  final DateTime offerExpiresAt;
  // NOTE (§9.1.1 gap): fare estimate / distance-to-pickup fields should be
  // added here once the backend offer payload includes them.
}

// lib/features/trip/data/active_trip_repository.dart (ActiveTrip) — ✅
class ActiveTrip {
  final String tripId;
  final String pickupAddress;
  final String dropoffAddress;
  final String status;   // driver_assigned | driver_arrived | in_progress | settled | ...
  final int finalFare;
  final String fareCurrency;
}
```

There is also a **richer, unused** `TripOffer`/`RiderInfo`/`TripOfferState` set of models in `features/trips/domain/models/` (plural, preview module) that include a `RiderInfo` (name + rating) the live `TripOffer` above lacks. **Recommendation:** when adding a real fare estimate to the live offer model (per §9.1.1's gap), also promote `RiderInfo` into the live model — showing the rider's name and rating on the offer card is standard ride-hailing practice and the data model already exists, just in the wrong (preview-only) module.

## 15.6 Rating — ✅ implemented (submission), 🔶 proposed (retrieval/list)

Submission shape (inferred from `TripPage._submitDriverRating`): `{ stars: int, comment: String? }`. A first-class `Rating` read-model for §12.2's "Đánh giá từ khách hàng" list:

```dart
// lib/features/profile/domain/models/rating.dart — 🔶 proposed
class Rating {
  final String tripId;
  final int stars;
  final String? comment;
  final DateTime submittedAt;
  final String riderDisplayName;   // likely first-name-only or initials for privacy
}
```

## 15.7 Notification — 🔶 proposed
See §13.2 (`DriverNotification`).

## 15.8 Support Ticket — 🔶 proposed, minimal for v1

```dart
// lib/features/profile/domain/models/support_ticket.dart
class SupportTicketRequest {
  final String category;    // e.g. "payment", "trip_dispute", "app_bug", "other"
  final String message;
  final String? relatedTripId;
}
```

v1 can be as simple as a form that `POST`s this to a backend ticket-creation endpoint (🔶 proposed, likely delegates to an external helpdesk tool rather than a bespoke backend — flag to product/ops before building backend storage for this, a third-party helpdesk integration (Zendesk, Freshdesk, etc.) may be more appropriate than a homegrown ticket system).

---

# 16. Offline Strategy

## 16.1 What must survive an app kill / connectivity loss

| Data | Current handling | Target |
|---|---|---|
| Auth token | ✅ `TokenStorage` (shared_preferences) — survives kill | keep as-is |
| Active trip ID | ✅ `TripStorage`-equivalent for driver (verify `ActiveTripRepository.getStoredTripId()`/`saveActiveTripId()` persistence mechanism — appears to be local storage-backed, confirmed by `_initialize()`'s restore-on-launch logic) | keep as-is |
| Online/offline toggle state | ✅ restored via `GET /api/v1/driver/availability` on `MapPage.initState` — server is the source of truth, not local cache, which is the *correct* choice (a driver's online status must reflect what dispatch actually believes, not a stale local flag) | keep as-is |
| Wallet balance | 🔶 should be cached (last-known value shown immediately, with a background refresh) rather than blank-loading every time — money screens should never feel slow | new |
| Trip history | 🔶 currently re-fetches every time (`RefreshIndicator` + no cache) — acceptable for v1, but consider a simple in-memory cache (not persisted to disk) to avoid a loading flash on tab re-visit within the same session | new, low priority |

## 16.2 Retry / reconnect behavior

- `ApiClient` has a 15s timeout (`_timeout`), throws `ApiException` with the Vietnamese message "Hết thời gian chờ. Kiểm tra kết nối và thử lại." (localized this session) — ✅ good baseline.
- **Trip-critical polling (`_poll`, `_paymentPoll`) silently swallows `ApiException`** and just retries on the next tick — ✅ correct for a background poll (don't show an error toast every 5 seconds during a network hiccup), but there's **no user-visible "connection lost" indicator** if polling fails for an extended period (e.g. 30+ seconds). 🔶 Recommend: if N consecutive poll failures occur, surface a small non-blocking banner ("Mất kết nối — đang thử lại…") so the driver isn't left wondering if the app is just quiet or actually broken.

## 16.3 Lost internet mid-trip

Current behavior: polling/action calls fail silently or surface `_PageState.error` (for direct actions like accept/arrive/start/finish — these are **not** silently retried, correctly, since retrying a mutating action blindly risks double-submission; the existing idempotency work — per project history "idempotency, and dispatch engine lifecycle" — should make server-side retries safe, but the client still correctly shows an explicit retry button rather than auto-retrying a mutation).

🔶 Gap: if connectivity drops *during* an active trip (not while polling for offers), the driver has no way to know if their `finish` request actually went through before connectivity dropped. Recommend: on reconnect, always re-fetch trip status via `GET /api/v1/rides/{tripID}` before assuming a failed mutation needs to be retried (this pattern already exists for app-relaunch recovery in `_initialize()` — extend the same "always check server truth before retrying" principle to a live reconnect event, not just a cold start).

## 16.4 GPS unavailable mid-trip

If GPS drops mid-trip, `TripMetricsEngine`'s distance/duration accumulation stalls silently. 🔶 No current fallback — if the finish-trip metrics end up wildly wrong due to a GPS gap, there's no server-side sanity check visible in the client flow. This is primarily a backend concern (pricing/booking service should sanity-bound client-reported distance/duration against straight-line pickup-dropoff distance) rather than something to fix in Flutter, but the *client* should at minimum detect a prolonged GPS gap (`LocationEngine` stream silence) and show a small "Tín hiệu GPS yếu" indicator so the driver isn't blindsided by a low/wrong fare later.

---

# 17. Error Handling

| Error class | Current handling | Target/gap |
|---|---|---|
| Network unreachable | `ApiException` with localized timeout message | ✅ adequate |
| Timeout (15s) | Same as above | ✅ adequate |
| GPS disabled | Full-screen blocked state (§6.1) | ✅ good |
| Location permission denied (soft) | Full-screen blocked state, retry action | ✅ good |
| Location permission denied (permanent) | Full-screen blocked state, deep-link to Settings | ✅ good |
| Server unavailable (5xx) | Generic `ApiException` message from backend, or fallback string | 🔶 should distinguish "server is down" from "your request was invalid" with different copy — currently both surface similarly |
| Trip expired (offer countdown hit 0 server-side before client) | Handled implicitly — accept would just fail with a server error | 🔶 map this specific case to a friendly "Chuyến đi đã hết hạn, vui lòng chờ chuyến mới" instead of a raw error string (§9.1.2) |
| Driver logged in elsewhere | 🚫 not currently detectable client-side (§8.3 backend gap) | 🔶 once backend enforces single-session, client needs a 401-with-specific-reason handler distinct from plain token expiry, with copy like "Tài khoản đã đăng nhập trên thiết bị khác" |
| JWT expired | 🔶 not handled — no 401 interceptor exists (§5.2 gap) | 🔶 build the interceptor: any 401 → clear session → redirect to login with a toast |
| Trip already accepted by another driver | Generic error today | 🔶 friendly-map per above |
| Withdrawal exceeds daily limit / balance | n/a (feature doesn't exist yet) | 🔶 client-side pre-validation + server-side authoritative check (§10.4.2) |

**General principle established by the existing codebase and to be maintained:** every error path must end in either (a) an automatic, safe retry (polling), or (b) a visible, tappable retry/next-action (mutating requests). Never leave a driver looking at a spinner or a dead screen with no way forward — this was literally the bug that started this session's driver-app investigation (the map stuck loading forever due to a missing permission declaration), and it's the single failure mode this entire document is written to prevent from recurring in every other corner of the app.

---

# 18. Loading UX

## 18.1 Current patterns (✅, keep using these — do not introduce a new loading paradigm)

- `AsyncStateView<T>` — the one true loading/success/empty/error wrapper, already shared with the rider app's pattern (though driver app currently has its own copy at `shared/widgets/async_state_view.dart` rather than a shared package — acceptable duplication for now given the apps are separate Flutter projects with no shared package today; 🔶 consider extracting a shared Dart package (`packages/panda_ui`) if a third app or significant further duplication emerges, but not worth the refactor for two files today).
- In-button spinners for single-action buttons (login, availability toggle, offer accept/reject) — ✅ correct pattern, avoids a jarring full-screen loading state for a fast, localized action.
- `CircularProgressIndicator` centered, with a label, for slower/page-level loads (`_LabelledSpinner` in `TripPage`) — ✅ good, the label ("Đang khởi động…", "Đang chấp nhận chuyến…") keeps the driver informed of *what* is loading, not just *that* something is.

## 18.2 What's missing

- **No skeleton loaders anywhere** (only spinners). For content-heavy new screens (Wallet balance, Earnings list, Notification list), a skeleton shimmer feels meaningfully faster than a spinner for the same actual load time — this is worth the small added complexity for the money-related screens specifically (§10.4.1, §11). Not worth retrofitting onto every existing spinner-based screen — apply selectively to new, high-value screens only.
- **No optimistic updates outside the trip-accept flow.** The existing optimistic accept (§9.1.2) is the right model — apply the same principle to Withdraw (show the pending withdrawal in the transaction list immediately, reconcile on next fetch) once Wallet is built.

## 18.3 Retry affordance consistency

Every error state's retry button should say **"Thử lại"** — confirmed this is already the consistent copy across the rider app and the driver app's existing error views. Maintain this exact string for every new error view in this spec; do not introduce synonyms ("Tải lại", "Làm mới lại") that would fragment the app's voice.

---

# 19. Design System

**Constraint:** PandaDriver is Material 3, per this project's existing `AppTheme` (confirmed in both apps). This section defines a driver-specific application of the shared design language, not a new design system — do not fork the visual identity between rider and driver apps beyond what a driver's context genuinely requires (e.g. larger tap targets for glance-and-tap-while-driving use, which the rider app doesn't need to optimize for as aggressively).

## 19.1 Colors

Reuse the existing `ColorScheme` from `AppTheme` (Panda's established primary/secondary/error/tertiary palette — do not introduce Be's yellow/navy branding, that was reference material only). Semantic additions specific to driver context:

| Token | Usage |
|---|---|
| `colorScheme.primary` | Online status, primary CTAs (Accept, Start, Finish) |
| `colorScheme.error` | Offline status accents, Reject button outline, urgent countdown (≤10s) |
| `colorScheme.tertiary` | "In progress" status banner (already used this way in `TripPage._StatusBanner`) |
| `colorScheme.secondary` | "Awaiting/pending" states (already used this way for `driver_assigned`-not-yet-arrived) |
| Green (earnings positive) | Wallet credit amounts, payout totals — currently `Colors.green`/primary interchangeably across the codebase; 🔶 standardize on a single semantic "positive money" color token when Wallet is built, don't reuse raw `Colors.green` ad hoc |
| Red/error (earnings negative) | Fee/tax deduction line items in Earnings Detail (§11.4) |

## 19.2 Typography

Material 3 default type scale via `Theme.of(context).textTheme`, as already used throughout — `titleLarge`/`titleMedium` for headers, `bodyMedium`/`bodySmall` for content, `labelSmall` for captions (e.g. address-row labels). No custom font has been introduced in either app — keep the platform default (Roboto/system) rather than adding a custom font asset, which is unjustified weight for no evidenced brand requirement.

## 19.3 Spacing

Observed convention across the codebase: 4px grid, most common paddings are `8`, `12`, `16`, `20`, `24`. Maintain this — do not introduce a different spacing scale for new screens (e.g. an 8pt-strict Material spec) that would visually clash with existing padding choices.

## 19.4 Corner radius

Cards/containers consistently use `BorderRadius.circular(12)`; bottom sheets and page-top corners use `BorderRadius.vertical(top: Radius.circular(20))` (confirmed in `BookingBottomSheet` on the rider side, apply the same 20px sheet-radius convention to any new driver bottom sheet, e.g. Quick Actions §3.1).

## 19.5 Buttons

- `FilledButton` — primary action (Accept, Start Trip, Confirm Withdraw).
- `OutlinedButton` — secondary/destructive-adjacent action (Reject, in the error-red variant already used in `TripPage`).
- `TextButton` — tertiary/low-emphasis (Skip rating, Sửa/Edit).
- Minimum height `52` for primary actions in trip-critical flows (`Size.fromHeight(52)`, already the established convention in `TripPage`) — apply the same 52px minimum to Withdraw's confirm button and any other high-stakes action, smaller default Material button heights are not appropriate for a driver glancing quickly while stationary/idling.

## 19.6 Cards

`Card` with `EdgeInsets.all(20)` internal padding for content-rich cards (offer card, trip-execution card) — maintain for Wallet balance card and Earnings detail card, these are the same "important information block" pattern.

## 19.7 Icons

Material Icons throughout, no custom icon set. Status-specific icon choices already established and should be reused for consistency:
- `Icons.power_settings_new` — offline/toggle-off
- `Icons.bolt` — online/active
- `Icons.local_taxi` — trip/dispatch context
- `Icons.location_on` (primary color) — pickup
- `Icons.flag` (error color) — dropoff
- `Icons.directions_car` / `Icons.directions_car_outlined` — vehicle/idle map state

New icons needed: `Icons.account_balance_wallet` (Wallet — note: already used in the *bottom nav* for the rider-side "Earnings"-equivalent concept per the driver nav; reuse the same icon for the new Wallet entry point for visual consistency rather than picking a different wallet glyph), `Icons.receipt_long` (Earnings Detail), `Icons.support_agent` (Support), `Icons.description_outlined` (Documents).

## 19.8 Map controls

Existing: default Google Maps zoom controls (`zoomControlsEnabled: true`), compass (`compassEnabled: true`), no custom map toolbar (`mapToolbarEnabled: false` — correct, avoids Google's default "open in Maps app" toolbar cluttering a driver-app map). New controls needed per §7.1: a custom recenter FAB (white circular button, `Icons.my_location`, bottom-right, matching the reference app's floating-button visual style but using PandaDriver's own elevation/shadow tokens, not copied pixel values).

## 19.9 Bottom sheets

`showModalBottomSheet` + `DraggableScrollableSheet`, exactly as already implemented for `BookingBottomSheet` on the rider side — reuse this exact pattern for the new Quick Actions sheet (§3.1) rather than a fixed-height sheet, since the content (icon grid) may grow over time.

## 19.10 Dialogs

`AlertDialog`, Material 3 default styling, used sparingly (per this spec's recommendation to *avoid* modal interruption for time-sensitive trip states, §3.1) — reserved for genuinely blocking confirmations (logout, end-trip-early, leave-active-trip-screen).

---

# 20. Animations

## 20.1 Existing (✅)

- `AnimatedSwitcher` (300ms) for Loading→Success/Empty/Error cross-fades in `AsyncStateView` — ✅ keep as the standard for all new async content.
- Implicit `setState`-driven rebuilds for countdown timers, status banners — no explicit animation curve specified today (instant snap on state change for banners/labels) — acceptable, over-animating status text would feel sluggish for a driver who needs information fast, not delightful.
- `GoogleMap.animateCamera` for map recentering (used in the rider app's place-search feature built this session; the driver map does not yet use programmatic camera animation since it has no search/selection flow — n/a here today).

## 20.2 Screen transitions

Default `go_router`/Material page transitions (platform default slide/fade) — no custom transition curve implemented. 🔶 No change recommended — custom page transitions are a low-value polish item for a utility-focused driver app; do not spend engineering time here before the functional gaps (§9, §10, §13) are closed.

## 20.3 Trip request animation

🔶 Proposed: the offer card (§9.1.1) should animate in with a brief scale/slide-up (200-300ms) rather than an instant `setState` pop-in, paired with the audible/haptic alert (§9.1.1's biggest gap) — a driver's attention needs to be drawn to this specific moment more than any other in the app, so it's the one place worth deliberate motion design investment.

## 20.4 Button press feedback

Standard Material ripple (`InkWell`/button defaults) throughout — ✅ no gaps, this is free from the Material widget set.

## 20.5 Loading

Standard `CircularProgressIndicator` (indeterminate) — no custom Lottie/Rive animation anywhere in either app today. 🔶 Not recommended to introduce a custom loading animation asset — adds a dependency and asset-maintenance burden disproportionate to the value over the platform default, for a utility app where drivers want speed and clarity over delight.

## 20.6 Bottom sheet

Default `DraggableScrollableSheet` physics (already used, feels native, matches platform conventions) — ✅ no change.

## 20.7 Hero animation

None implemented anywhere in either app. 🔶 Not recommended for this app's screen set — Hero transitions shine for image-to-detail flows (e.g. e-commerce product galleries); there's no analogous "shared visual element between two screens" moment in this driver app's flows that would benefit from one. Skip.

---

# 21. Future Features

Explicitly out of scope for the v1 roadmap (§22) but worth naming so they aren't rediscovered from scratch later:

| Feature | Notes |
|---|---|
| In-app chat (driver ↔ rider) | Requires new realtime messaging infra; most ride-hailing apps route this through phone call/SMS instead for v1 — recommend the same (a simple `tel:` deep-link "Gọi khách hàng" button, cheaper than building chat) |
| Voice guidance / in-app navigation | Explicitly rejected in §9.1.4 in favor of deep-linking to Google Maps — revisit only if deep-linking proves insufficient in practice |
| Fuel price / fuel station finder | Vehicle-type-specific (ICE only), low differentiation value |
| EV charging station finder | Only relevant once/if PandaDriver has meaningful EV-driver share; the reference app's "Trạm sạc gần đây" is 🚫 not applicable today |
| Leaderboard / gamification / quests | Reference app's "Thành tích hoạt động" territory — genuinely effective for engagement in mature markets, but needs a dedicated incentives/economics design (not just a UI), defer to a growth-team-led initiative |
| Referral program | Straightforward to build (§12.2) once there's a referral-tracking backend; not urgent pre-launch |
| Driver ranking / tiers | Similar to rider app's `MemberLevel` (Standard/Silver/Gold/Platinum) concept already modeled on the rider side (`DOC-0002 §6.7` commission tiers referenced in rider code comments) — a driver-side tier system tied to commission rates is a natural future pairing, but is a pricing/business-model decision, not an engineering one |
| Biometric login | Deferred per §5.4 — low value while sessions already persist indefinitely and there's no password to replace |
| OTP pickup verification | Deferred per §9.1.6 — only build if trip-hijacking/wrong-pickup becomes a measured fraud vector |
| Background location / foreground service | Not "future" so much as "next" — flagged as high-priority in §6.2, listed here only for completeness of the full feature inventory |

---

# 22. Flutter Implementation Priority

Phased roadmap. Each phase should ship as a usable increment, not a big-bang release.

## Phase 1 — Fix and harden what exists (do this first, before any new screen)

- Fix the 401/expired-session gap (§5.2, §17) — an interceptor in `ApiClient`.
- Fix the rider-cancellation-during-payment-wait gap (§8.3, §9.1.9) — verify and patch `_paymentPoll`.
- Add the audible/haptic new-offer alert (§9.1.1) — highest driver-value, lowest-effort item in the entire spec.
- Wire the already-built waiting-timer widgets into the live `TripPage` (§9.1.5) — nearly free, widgets exist.
- Resolve the `HomePage` vs. `MapPage` duplication (§7.2) — pick one, delete or repurpose the other.
- Add a Google Maps navigation deep-link button (§9.1.4).

## Phase 2 — Home screen polish

- Persistent online/offline status banner (§7.4), wired to `DriverActivityStatus`.
- Wallet-balance shortcut pill (stub-linkable even before Wallet ships — can point to a "Sắp ra mắt" placeholder initially if Wallet backend isn't ready).
- Recenter FAB.
- Quick Actions bottom sheet (start with just "Hỗ trợ" and "Sổ tay tài xế" as the only two real entries; the rest are 🚫/low-priority per §7.1).

## Phase 3 — Trip experience depth

- Fare estimate + distance-to-pickup on the offer card (§9.1.1) — needs backend payload extension.
- Rider name/rating on the offer card (§15.5) — promote existing `RiderInfo` model from preview module.
- Payment-method-aware "awaiting payment" copy (§9.1.9).
- Friendly error-message mapping for known failure cases (§9.1.2, §17).

## Phase 4 — Earnings (buildable now, no backend blocker for the summary)

- `EarningsPage` rewrite: period tabs + summary, aggregating `GET /api/v1/driver/trips` client-side (§11.3).
- Earnings Detail page — **blocked** until backend exposes fee/tax breakdown per trip; build the UI against a mock repository first (matches this project's established pattern for building ahead of backend, e.g. rider app's mock notification repository), swap to real data when ready.

## Phase 5 — Wallet (backend-blocked, coordinate with backend team before starting)

- Confirm the §10.3 API contract with backend.
- Build Wallet summary, Withdraw, Wallet History pages against the confirmed contract (or a mock repository if backend timeline lags Flutter capacity).

## Phase 6 — Notifications

- Build the list UI against a mock repository immediately (§13.6) — no blocker.
- Layer in real backend + push (FCM) once available (§13.3).

## Phase 7 — Profile expansion

- Vehicle Info, Documents (read-only), Support form, Settings (notification toggle only) — in that priority order, per §12.3's "highest driver-value, lowest-backend-dependency" ranking.

## Phase 8 — Permission hardening

- Background location + foreground service (§6.2).
- Notification permission flow (§6.3), once push is being built (Phase 6).
- Battery optimization exemption explainer (§6.4).

## Phase 9 — Deferred / opportunistic

- Everything in §21, plus heatmap/demand-zones (§7.3) once the dispatch team scopes `GET /api/v1/dispatch/demand-zones` — this is intentionally last because it's the most backend-heavy, most cosmetic-relative-to-effort item in the whole spec.

---

# 23. Components Library

Reusable widgets this spec requires, organized by whether they already exist (✅, reuse) or need building (🔶, new). Naming follows the existing codebase's convention (`PascalCase`, feature-agnostic widgets live in `shared/widgets/`, feature-specific ones stay in their feature's `presentation/widgets/`).

| Component | Status | Notes |
|---|---|---|
| `AsyncStateView<T>` | ✅ | `shared/widgets/` — the loading/success/empty/error workhorse |
| `AvailabilityToggle` | ✅ | `features/home/presentation/widgets/` |
| `DriverSummaryHeader` | ✅ (unwired) | harvest from legacy `HomePage`, see §7.2 |
| `HomeStatsRow` | ✅ (unwired) | same |
| `HomeStatusCard` | ✅ (unwired) | same |
| `QuickActionCard` | ✅ (unwired) | same, reuse for §3.1's Quick Actions sheet |
| `WalletBalanceCard` | 🔶 new | §10.4.1 — large amount + CTA, model after rider's `FareSummaryCard` visual weight |
| `WalletTransactionTile` | 🔶 new | §10.4.3 |
| `IncomeCard` (Earnings summary card) | 🔶 new | §11.1 — big total + period navigator |
| `EarningsBreakdownRow` | 🔶 new | §11.4 — label/value row with optional info-icon tooltip, negative-amount styling |
| `EarningsPeriodTabs` | 🔶 new | §11.1 — Ngày/Tuần/Tháng segmented control |
| `MapRecenterFab` | 🔶 new | §7.1/§19.8 |
| `OnlineStatusBanner` | 🔶 new | §7.4 — persistent top banner, replaces/extends current implicit toggle-only status |
| `TripOfferCard` (live) | ✅ (`_OfferCard` in `TripPage`) | consider promoting from private `_OfferCard` to a named, testable public widget as it grows (fare estimate, rider info additions in Phase 3) |
| `WaitingTimerBadge` | ✅ (exists in preview module, unwired) | promote from `features/trips/presentation/widgets/waiting_timer.dart` |
| `WaitingFeeCard` | ✅ (exists in preview module, unwired) | promote from `features/trips/presentation/widgets/waiting_fee_card.dart` |
| `RatingBadge` | 🔶 new | small star+number chip, reusable across Profile header, offer card (rider rating), earnings detail |
| `NotificationTile` | 🔶 new | model after rider app's `notification_tile.dart` |
| `UnreadBadge` | 🔶 new (port) | direct port of rider app's `unread_badge.dart` |
| `VehicleInfoCard` | 🔶 new | §12.3 — could adapt the rider app's existing `vehicle_info_card.dart` (built for showing a *driver's* vehicle to a *rider*) for the reverse direction |
| `SupportTicketForm` | 🔶 new | §15.8 |
| `HeatmapLegend` | 🔶 new, **blocked** | §7.3 — do not build before demand-zones backend exists |
| `DemandZoneOverlay` | 🔶 new, **blocked** | same |

---

# 24. State/Controller Inventory

*(Retained heading name per spec convention; content lists the actual pattern in use — plain `State<T>` classes paired with `Repository` objects, not Bloc/Cubit. See §2.1 for why.)*

| Screen | Controller (State class) | Backing Repository | Status |
|---|---|---|---|
| Login | `_LoginPageState` | `AuthRepository` | ✅ |
| Home/Map | `_MapPageState` | `AvailabilityRepository`, `LocationUploadService` | ✅ |
| Trip | `_TripPageState` | `TripOfferRepository`, `ActiveTripRepository` | ✅ |
| Trip History | `_DriverTripHistoryPageState` (or equivalent) | inline `ApiClient.get` (no dedicated repository class today — 🔶 consider extracting a `DriverTripHistoryRepository` for consistency with every other screen's pattern) | ✅ (screen), 🔶 (refactor) |
| Earnings (rewrite) | `_EarningsPageState` | `EarningsRepository` (new, §11) | 🔶 |
| Earnings Detail | `_EarningsDetailPageState` | `EarningsRepository.fetchDetail()` | 🔶 |
| Wallet | `_WalletPageState` | `WalletRepository` (new, §10) | 🔶 |
| Withdraw | `_WithdrawPageState` | `WalletRepository.withdraw()` | 🔶 |
| Wallet History | `_WalletHistoryPageState` | `WalletRepository.fetchTransactions()` | 🔶 |
| Notifications | `_NotificationsPageState` | `NotificationsRepository` (new, §13) | 🔶 |
| Profile | `_ProfilePageState` | `DriverProfileRepository` (new — currently profile data appears embedded/mocked rather than a dedicated repository; verify and extract) | ✅ (partial), 🔶 (repository extraction) |
| Vehicle Info | `_VehicleInfoPageState` | `DriverProfileRepository.fetchVehicle()` | 🔶 |
| Support | `_SupportPageState` | `SupportRepository` (new) | 🔶 |

**Cross-cutting "controllers" (not page-scoped):**

| Name | Role | Status |
|---|---|---|
| `AuthState` | app-wide login/logout, drives router redirect | ✅ |
| `TokenStorage` | JWT persistence | ✅ |
| `LocationUploadService` | periodic location push while online, exposes `locationStream` also consumed by `TripMetricsEngine` | ✅ |
| `TripMetricsEngine` | client-side distance/duration accumulation during an active trip | ✅ |

---

# 25. Final TODO

Implementation checklist. Grouped by the phases in §22. Check items off as they ship.

## Phase 1 — Fix & harden

- [ ] Add 401-expired-session interceptor to `ApiClient` (driver app) — clear token, redirect to login, toast
- [ ] Verify and fix rider-cancellation handling in `TripPage._paymentPoll`
- [ ] Add audible + haptic alert on new trip offer (`vibration` + a short notification sound asset)
- [ ] Promote `WaitingTimerBadge`/`WaitingFeeCard` from `features/trips` (preview) into `features/trip` (live), wire into `_TripExecutionCard`
- [ ] Resolve `HomePage` vs `MapPage` duplication — decide, then delete or repurpose `HomePage` and its unwired widgets
- [ ] Add `url_launcher` dependency; add "Chỉ đường" button deep-linking to Google Maps navigation from the active-trip card
- [ ] Confirm/patch backend: reject `go-offline` while driver has an active trip (409)

## Phase 2 — Home polish

- [ ] Build `OnlineStatusBanner`, wire to `DriverActivityStatus`
- [ ] Build `MapRecenterFab`
- [ ] Add Wallet-balance shortcut pill to Home (placeholder-linkable if Wallet isn't ready yet)
- [ ] Build Quick Actions bottom sheet (Hỗ trợ, Sổ tay tài xế entries only for v1)

## Phase 3 — Trip depth

- [ ] Backend: extend offer payload with fare estimate, distance-to-pickup, rider name+rating
- [ ] Client: wire real values into `_OfferCard`/`_InfoChip`s (replace `—` placeholders)
- [ ] Promote `RiderInfo` model from preview module into live `TripOffer`
- [ ] Add payment-method field to `ActiveTrip`, branch "awaiting payment" copy on cash vs. digital
- [ ] Map known error codes (offer taken, offer expired) to friendly Vietnamese copy

## Phase 4 — Earnings

- [ ] Rewrite `EarningsPage`: `EarningsPeriodTabs`, aggregate `GET /api/v1/driver/trips` client-side into `EarningsSummary`
- [ ] Build `IncomeCard` component
- [ ] Build Earnings Detail page UI against a mock `EarningsRepository`
- [ ] Backend: expose fee/tax breakdown per trip (`GET /api/v1/driver/earnings/{tripId}`)
- [ ] Swap Earnings Detail to real backend once available

## Phase 5 — Wallet

- [ ] Confirm `WalletSummary`/`WalletTransaction`/withdrawal API contract with backend team (§10.3)
- [ ] Build `WalletPage` + `WalletBalanceCard` against mock or real repository
- [ ] Build `WithdrawPage` with client + server validation
- [ ] Build `WalletHistoryPage` with pagination
- [ ] Backend: `wallet` service/extension, minimum-reserve and daily-limit as server-configured values

## Phase 6 — Notifications

- [ ] Build `NotificationsPage` list UI against a mock `NotificationsRepository` (port rider app's pattern)
- [ ] Build `NotificationTile`, port `UnreadBadge`
- [ ] Backend: notification persistence + list/read endpoints (shared with rider app if feasible)
- [ ] Add `firebase_messaging`, device-token registration endpoint, backend push-send capability
- [ ] Add `POST_NOTIFICATIONS` permission flow (Android 13+)

## Phase 7 — Profile expansion

- [ ] Extract `DriverProfileRepository` (if not already separated from inline fetch logic)
- [ ] Build Vehicle Info page (verify if vehicle data already exists server-side before assuming new backend work)
- [ ] Build Documents page (read-only view)
- [ ] Build Support ticket form + backend/helpdesk integration decision
- [ ] Build Settings page (notification toggle, minimal v1)
- [ ] Add logout confirmation dialog
- [ ] Build aggregate ratings list page + backend endpoint

## Phase 8 — Permissions hardening

- [ ] Background location: choose `flutter_foreground_task` vs. native approach, add `ACCESS_BACKGROUND_LOCATION` with proper Android 10+/13+ prompt sequencing
- [ ] Add foreground service persistent notification while online
- [ ] Add battery-optimization-exemption explainer screen + Settings deep-link
- [ ] Add Play Store data-safety disclosure updates for background location (compliance, not engineering, but blocks release)

## Phase 9 — Deferred / opportunistic

- [ ] Backend: scope `GET /api/v1/dispatch/demand-zones`
- [ ] Build `HeatmapLegend` + `DemandZoneOverlay` once backend above ships
- [ ] Revisit §21 items as product priorities warrant (referral, achievements, driver tiers, OTP pickup verification, biometric login, chat)

## Ongoing / cross-cutting

- [ ] Every new screen uses `AsyncStateView<T>` — no hand-rolled `FutureBuilder`
- [ ] Every new mutating action shows a visible retry, never a silent failure or an infinite spinner
- [ ] Every new user-facing string is written in Vietnamese from the start (no English-then-translate-later pattern, per this project's established localization discipline)
- [ ] Every new permission request has its manifest declaration added in the same commit (see §6.1's cautionary tale)
- [ ] `flutter analyze` clean before every merge

---

*End of specification. This document should be treated as living — update the ✅/🔶/🚫 status markers as work ships, rather than maintaining a separate changelog.*
