# FAIRRIDE EOS — Project Memory
Last updated: 2026-07-07 by Principal Engineer AI

## Current Phase
Phase D-06 — Driver App: Arrived at Pickup (COMPLETE)
Previous: Phase D-05 — Driver App: Driver Assigned & Navigation Ready (COMPLETE)

## Phase D-06 — Driver App: Arrived at Pickup (COMPLETE — 2026-07-07)

Reference: `docs/project/MVP_DEVELOPMENT_PLAN.md` Section 6 (Driver App
Roadmap, stage D5 — UI portion, completes the pickup leg started in D-05)
and Section 12 (Phase Registry). UI only — no backend, gRPC, proto, map/GPS.
No state management library, no new repository (per the task: "Không tạo
repository mới"). Scope strictly `apps/driver/lib/features/trips/` — Rider
app, Home dashboard, auth were not touched. Flow covered:
`NavigatingToPickup → ArrivedAtPickup` only; `passengerBoarding` does **not**
exist yet (explicitly out of scope).

### State machine — extended, and the *trigger* for the new state is progress, not a timer

`TripOfferState` (D-05, 8 values) extended to **9 values**: added
`arrivedAtPickup`. Critically, per the task ("Không chuyển bằng timer cứng.
State phải đổi khi progress hoàn thành"), this transition is **not** driven
by an independent timer — it's driven by the *same* route-progress ticker
from D-05 finally reaching its end:

- `RouteProgressModel.stepDown()`'s floor changed from **20 → 0** (D-05
  explicitly floored at 20 because `arrived` didn't exist yet; now it does).
  `RouteProgressIndicator`'s trailing label reads "Arrived" instead of
  "0% remaining" once `progress <= 0`.
- `_RouteProgressTicker` (private, in `trip_offer_view.dart`) gained an
  `onArrived` callback. Its `_scheduleTick()` now checks, after each
  `stepDown()`, whether the new progress is `<= 0`; if so it calls
  `onArrived()` **exactly once** and returns without scheduling another
  tick (previously it just returned once `<= 20`, since 20 was the end).
  `TripsPage._handleArrived()` (guarded like every other transition:
  `if (_state != navigatingToPickup) return;`) flips the state.
- A defensive edge case: if a ticker were ever seeded already at 0 (never
  happens from the live repository, which always starts at 100), calling
  `onArrived()` synchronously from `initState` would risk a "setState during
  build" error one level up (in `TripsPage`) — guarded with a
  `WidgetsBinding.instance.addPostFrameCallback`.

### New widgets under `apps/driver/lib/features/trips/presentation/widgets/`

| File | Purpose |
|------|---------|
| `waiting_timer.dart` | `WaitingTimer` — self-ticking mm:ss counter. **No package** — a recursive `Future.delayed(1s)` chain, the exact same self-scheduling shape as `_RouteProgressTicker` (D-05), with its own `_stopped` dispose-guard. `onMinutePassed(int minute)` fires once per whole minute (mock only). `initialSeconds` (default 0) lets a caller seed a starting point — used by the Arrival Preview to show "Waiting 03:00"/"08:00" instantly instead of waiting 3/8 real minutes; the timer keeps ticking live from that seed rather than freezing, same philosophy as every other preview page in this app reusing real, interactive widgets instead of inventing frozen mock renderings |
| `waiting_fee_card.dart` | `WaitingFeeCard` — pure `StatelessWidget`, computed entirely from `elapsedMinutes`: free for the first 5 minutes, then 2.000đ per additional minute (own tiny thousands-separator formatter — no `intl`/package, since "Không thêm package" applies here too). "Chỉ UI" — no backend billing |
| `passenger_action_panel.dart` | `PassengerActionPanel` — Passenger On Board / Contact Rider / Cancel Trip. All three are **plain callbacks only** — no phone dialer, no popup/dialog. "Passenger On Board" is a placeholder like the other two since `passengerBoarding` isn't a state yet |
| `driver_arrival_card.dart` | `DriverArrivalCard` — the Arrived screen. A `StatefulWidget` (holds `_elapsedMinutes`, fed by `WaitingTimer.onMinutePassed`, relayed into `WaitingFeeCard`) composing: the "Arrived at Pickup" status via the **existing, unmodified** `DriverStatusBanner` (task: "Không tạo widget mới" for the banner — reuse as-is with new icon/title/subtitle), pickup address via `TripAddressRow`, passenger name via `RouteStatTile`, estimated fare via `FareEstimateCard` (reused whole, not re-derived), then `WaitingTimer`/`WaitingFeeCard`/`PassengerActionPanel` |
| `presentation/pages/arrival_preview_page.dart` | `ArrivalPreviewPage` — "Arrival Preview": `ChoiceChip` row steps through Arrived/Waiting 00:00/03:00/08:00 by passing different `initialWaitingSeconds` into `DriverArrivalCard` directly — **no repository call**, cross-fades with `AnimatedSwitcher` |

**Modified:**
- `route_progress_model.dart` (`stepDown` floor default 20→0; doc comments updated)
- `route_progress_indicator.dart` (trailing label "Arrived" at progress 0)
- `trip_offer_view.dart` (+`arrivedAtPickup` case → `DriverArrivalCard`; `_RouteProgressTicker`/`_NavigatingToPickupContent` +`onArrived`; +`onArrived`/`onPassengerOnBoard` params on `TripOfferView`)
- `trips_page.dart` (+`_handleArrived` — guarded state transition, not a callback the UI invokes directly; +`_handlePassengerOnBoard` — plain SnackBar placeholder; wired both into the `TripOfferView(...)` call)
- `trip_offer_preview_menu_page.dart` (+"Arrival Preview" entry)

### Reuse discipline (explicit task requirement: "Reuse tối đa... Không duplicate")

`TripAddressRow`, `RouteStatTile`, `DriverStatusBanner`, `FareEstimateCard`
are all reused **verbatim** inside `DriverArrivalCard` — none were copied or
forked. `DriverStatusBanner` needed zero source changes (it was already
generic enough from D-05 to accept "Arrived at Pickup" as just another
icon/title/subtitle).

### Test-writing lesson learned this phase (new — a second twist on the D-04/D-05 timer lessons)

D-05 already established that `pumpAndSettle()` is unsafe around a bare
`Future.delayed` chain (it falsely reports "settled" without the next tick
firing, since nothing schedules a frame while just waiting) and that D-05's
route ticker could be **fully drained** with one big `pump(9s)` because it
had a floor (20%) where it permanently stopped rescheduling.

D-06 broke that assumption twice over:

1. Changing the floor to 0 meant the *old* D-05 tests' `pump(seconds: 9)`
   drain (sized for "4 ticks to the 20% floor") was now **one tick short**
   of the new 5-tick chain to the 0% floor — those tests started failing
   with "A Timer is still pending" the moment the floor changed, even
   though their own assertions never touched Arrived at all.
2. `WaitingTimer` has **no floor at all** — it ticks for as long as the
   Arrived screen is mounted. A test that reaches `arrivedAtPickup` can
   never "pump long enough" to drain it to completion, because there is no
   completion.

The fix, applied uniformly to both problems: stop trying to pump a
`Future.delayed` chain to its natural end when the test doesn't care about
reaching that end. Instead, **unmount the widget tree** (`pumpWidget` a
throwaway `SizedBox()`) — this runs `dispose()`, which flips each ticker's
`_stopped` guard — then pump past whatever *single* tick was already
in-flight before disposal, so it fires and returns immediately (via the
guard) instead of staying "pending." This works identically regardless of
how many ticks remain or whether the chain has a floor at all, which is why
it replaced the old fixed-duration drains for every D-05 navigation test too
(see `_disposeAndDrainPendingTimer` in `test/widget_test.dart` — formerly
`_drainWaitingTimer`, generalized once it became clear both tickers needed
the identical fix). Tests that *do* want to reach Arrived on purpose
(`_reachArrivedContent`) still use one big `pump(seconds: 11)` to cross all
5 route-ticker ticks (10s) plus buffer, since that chain **does** have a
natural end and cascading through it is the whole point of that helper.

**Verified this session:** `flutter analyze` — 0 issues (no warnings).
`flutter test` — 40/40 pass (32 from D-01–D-05, five of which needed their
drain updated for the floor change, behavior otherwise unaffected; + 8 new:
route progress completing transitions Navigating→Arrived, `WaitingTimer`
counts up in mm:ss and fires `onMinutePassed` once per minute, `WaitingTimer`
does not tick after being disposed, `WaitingFeeCard` free-then-charged
thresholds, `PassengerActionPanel`'s three buttons fire independently,
`DriverArrivalCard` renders its banner/address/passenger correctly, the
preview menu lists "Arrival Preview", `ArrivalPreviewPage` steps through
Arrived/Waiting 00:00–08:00 with zero repository/Future involvement).
`git status` confirms `apps/rider` untouched.

## Phase D-05 — Driver App: Driver Assigned & Navigation Ready (COMPLETE — 2026-07-07)

Reference: `docs/project/MVP_DEVELOPMENT_PLAN.md` Section 6 (Driver App
Roadmap, stage D5 — UI portion, pickup leg only) and Section 12 (Phase
Registry). UI only — no backend, gRPC, proto, map/GPS provider. No state
management library introduced. Scope strictly
`apps/driver/lib/features/trips/` — Rider app, Home dashboard, auth were not
touched. Flow covered: `Assigned → NavigatingToPickup` only; `Arrived` does
**not** exist yet (explicitly out of scope this phase).

### State machine — extended, not renamed this time

`TripOfferState` (D-04, 7 values) extended to **8 values**: added
`navigatingToPickup` at the end. Reached only from `assigned`, via
`TripsPage._handleNavigate()` (guarded: `if (_state != assigned) return;`,
same pattern as every other transition in this file).

### New domain model

`domain/models/route_progress_model.dart`:
- `TrafficLevel` enum (`normal, slow, heavy`) — **not a `String`**, per this
  project's recurring state-machine convention.
- `RouteProgressModel { remainingDistanceKm, remainingDurationMin, progress, trafficLevel }`.
  `progress` is percent **remaining** (100 → 20 this phase, floored — 0/
  "Arrived" doesn't exist yet).
- `RouteProgressModel.mock({progress, trafficLevel})` — pure, static factory:
  bases the pickup leg on the *same* 1.8km/6min numbers
  `TripOffer.distanceToPickupKm` and the Assigned screen's ETA already use
  (18 km/h assumed approach speed), so all three screens agree at 100%.
  Traffic only slows the ETA (normal ×1.0, slow ×1.3, heavy ×1.6) — the
  physical distance remaining is traffic-independent.
- `RouteProgressModel.stepDown({by = 20, floor = 20})` — recomputes the next
  tick via `.mock()` (not iterative scaling) to avoid floating-point drift.

### New repository method

`DriverTripOfferRepository.fetchRouteProgress({Duration delay = 600ms, TrafficLevel traffic = normal})`
→ `Future<RouteProgressModel>` — returns `RouteProgressModel.mock(progress: 100, trafficLevel: traffic)`
after the mock delay. `traffic` is dev-selectable via a new
`PopupMenuButton<TrafficLevel>` ("Traffic (dev)") in `TripsPage`'s AppBar —
directly the `TrafficLevel` enum as the parameter, no extra demo-mode enum
needed.

### New files under `apps/driver/lib/features/trips/`

| File | Purpose |
|------|---------|
| `domain/models/route_progress_model.dart` | `TrafficLevel` + `RouteProgressModel` (see above) |
| `presentation/widgets/route_stat_tile.dart` | Generic icon+value+label tile. Extracted so `TripAssignedCard`'s new Pickup ETA/Distance to Pickup stats and `DriverNavigationCard`'s Distance Remaining/ETA share one widget — also **de-duplicated** `FareEstimateCard`'s old private `_Stat` into this (identical layout, would have been a third copy otherwise) |
| `presentation/widgets/driver_status_banner.dart` | Icon-in-circle + title + optional subtitle ("Driving to Pickup"). Deliberately generic/reusable — the task explicitly asked for this to be reusable by a later phase (e.g. "Arrived", "Trip in Progress") |
| `presentation/widgets/route_progress_indicator.dart` | Mock linear progress bar (no map/GPS) + a color-coded traffic badge (green/orange/red for normal/slow/heavy) |
| `presentation/widgets/driver_navigation_card.dart` | The Navigation screen's main card: `DriverStatusBanner` + pickup address (reuses `TripAddressRow`) + Distance Remaining/ETA (`RouteStatTile` ×2) + `RouteProgressIndicator` + Cancel Trip/Contact Rider buttons. New widget, does **not** duplicate `TripAssignedCard` |
| `presentation/pages/navigation_preview_page.dart` | "Navigation Preview" — `ChoiceChip` row steps through Assigned/100%/80%/60%/40%/20% via `RouteProgressModel.mock(...)` directly, **no repository call**, cross-fades with `AnimatedSwitcher` |

**Modified:**
- `trip_offer_state.dart` (+`navigatingToPickup`, +label)
- `trip_assigned_card.dart` (+Pickup ETA/Distance to Pickup via `RouteStatTile`, +"Current status: Ready to navigate" line; `onNavigate`'s *meaning* changed upstream — the widget itself still just reports the tap)
- `driver_trip_offer_repository.dart` (+`fetchRouteProgress`)
- `fare_estimate_card.dart` (its private `_Stat` replaced with the new shared `RouteStatTile` — no behavior change, pure de-duplication)
- `trip_offer_view.dart` (+`navigatingToPickup` case → `_NavigatingToPickupContent`, a new private widget wrapping a **second, nested** `AsyncStateView<RouteProgressModel>` + `_RouteProgressTicker` — see below; +`routeFuture`/`onContactRider`/`onCancelTrip` params)
- `trips_page.dart` (`_handleNavigate` no longer just shows a SnackBar — it now transitions `assigned → navigatingToPickup` and calls `fetchRouteProgress()`; new `_handleContactRider`/`_handleCancelTrip`, both **plain placeholder callbacks only** — no phone dialer, no popup/dialog, matching the task's explicit "đều chỉ callback. Không mở điện thoại. Không popup." for both actions, same as `onNavigate` used to be before this phase; new "Traffic (dev)" `PopupMenuButton<TrafficLevel>`)
- `trip_offer_preview_menu_page.dart` (+"Navigation Preview" entry, linking to the new page)

### Architecture: nested `AsyncStateView`, kept separate from the business state machine

The `navigatingToPickup` case needed its *own* async fetch
(`fetchRouteProgress()`), independent of the outer offer fetch. Rather than
threading a raw `Future` through more state, `TripOfferView`'s new
`_NavigatingToPickupContent` wraps a **second** `AsyncStateView<RouteProgressModel>`
nested inside the outer `AsyncStateView<TripOffer?>` — a direct, deliberate
demonstration that this pattern composes at any depth without merging
Loading/Success/Empty/Error into the `TripOfferState` enum.

Once that inner fetch resolves, `successBuilder` returns a `_RouteProgressTicker`
(`StatefulWidget`) seeded with the fetched `RouteProgressModel`. It self-schedules
`stepDown()` every 2 seconds via a recursive `Future.delayed` chain — **not**
a repeated repository call, purely local computation — and stops scheduling
once `progress <= 20` (the floor). Because `successBuilder` is only
re-invoked when the `FutureBuilder`'s snapshot changes, and the ticker is a
genuine `StatefulWidget` at a stable tree position, its internal ticking
state survives parent rebuilds correctly.

### Test-writing lesson learned this phase (new — distinct from D-04's countdown lesson)

D-04 already established that `pumpAndSettle()` is unsafe while a
**continuously-ticking `AnimationController`** is in the tree (it never
settles). This phase surfaced the *opposite* failure mode: `pumpAndSettle()`
is *also* unsafe around a bare **`Future.delayed` chain with gaps between
ticks** — but for the opposite reason. `pumpAndSettle()` repeatedly pumps in
short bursts (default 100ms) and stops as soon as a burst produces no newly
scheduled frame. A `Future.delayed(2s)` schedules **no frames at all** while
it's waiting — nothing animates in the meantime — so `pumpAndSettle()`
falsely concludes the tree has "settled" almost immediately, without the 2s
timer ever firing. The fix used throughout the new tests: advance time with
explicit `pump(Duration(...))` jumps sized to actually cross each tick
boundary (e.g. `pump(2100ms)` to force exactly one `stepDown()`), and —
because the ticker leaves a pending `Future.delayed` running until it hits
the 20% floor — every test that enters `navigatingToPickup` finishes with a
large enough `pump()` (9s, comfortably past all 4 remaining ticks) to drain
it fully before the test ends, avoiding a "pending Timer" failure exactly
like the one already documented in D-04's Accept-delay tests.

Also hit, fixed immediately: `RouteStatTile`'s icon+`Column(value,label)` Row
had no `Expanded` around the column, so longer labels ("Distance to Pickup",
"Distance remaining") overflowed once two tiles shared a row at the app's
~400px test width — same "unconstrained value Text in a Row" class of bug
documented in earlier phases (`ReceiptRow`, `InfoRow`, `QuickActionCard`).
Fixed by wrapping the `Column` in `Expanded` with `maxLines: 1` +
`TextOverflow.ellipsis` on both `Text`s; the same fix was applied to
`TripAssignedCard`'s new one-line "Current status" row.

**Verified this session:** `flutter analyze` — 0 issues (no warnings).
`flutter test` — 32/32 pass (24 from D-01–D-04, unaffected, + 8 new: Start
Navigation transitions Assigned→Navigating, the Navigation screen renders
initial distance/ETA/traffic, route progress ticks 100%→80% over simulated
time, the traffic badge reflects the dev "Traffic" menu, Cancel Trip and
Contact Rider are plain callbacks with no `AlertDialog`/`Dialog`, the preview
menu lists "Navigation Preview", `NavigationPreviewPage` steps through
Assigned→100%→20% with zero repository/Future involvement). `git status`
confirms `apps/rider` untouched.

## Phase D-04 — Driver App: Accept Flow & Dispatch Session (COMPLETE — 2026-07-07)

Reference: `docs/project/MVP_DEVELOPMENT_PLAN.md` Section 6 (Driver App
Roadmap, stage D4 — UI portion, completing what D-03 started) and Section 12
(Phase Registry). UI only — no backend, API, or protobuf changes. No state
management library introduced. Scope strictly `apps/driver/lib/features/trips/`
— Rider app, Home dashboard, and auth were not touched.

### State machine — renamed and extended

`TripOfferStatus` (D-03, 4 values) → **renamed** to `TripOfferState` (task's
own naming convention) and extended to 7 values:
`newOffer, accepting, assigned, rejected, expired, failed, timeout`.
`accepted` was renamed to `assigned` — it's now the true terminal success
state after dispatch confirms (Dispatch Accepted + Trip Assigned + Ready to
Navigate collapsed into one screen, per the task). File renamed
`trip_offer_status.dart` → `trip_offer_state.dart`; all references (view,
pages, tests) updated accordingly.

**Two state machines, still not merged** (explicit requirement): `AsyncStateView`
(Loading/Success/Empty/Error — untouched, zero edits this phase) governs
whether an offer was fetched at all; `TripOfferState` governs the offer's own
lifecycle once fetched, including the new post-Accept sub-flow. `TripsPage`
composes them (`AsyncStateView.successBuilder` renders `TripOfferView` keyed
by `TripOfferState`) without collapsing them into one enum.

### New repository contract

`DriverTripOfferRepository.acceptOffer({Duration delay = 1200ms, DispatchAcceptStatus outcome = success})`
→ `Future<DispatchAcceptResult>`. `DispatchAcceptStatus` (success/failed/timeout)
and `DispatchAcceptResult` are new types in
`domain/models/dispatch_accept_result.dart` — a dedicated result type, not a
`bool`, so a third outcome (timeout) fits naturally.

### New files under `apps/driver/lib/features/trips/`

| File | Purpose |
|------|---------|
| `domain/models/trip_offer_state.dart` | Replaces `trip_offer_status.dart` — see above |
| `domain/models/dispatch_accept_result.dart` | `DispatchAcceptStatus` + `DispatchAcceptResult` |
| `presentation/widgets/trip_address_row.dart` | Extracted from `TripOfferCard`'s private `_AddressRow` so `TripAssignedCard` could reuse it instead of duplicating (explicit task requirement: "Không duplicate UI đã có") |
| `presentation/widgets/accept_loading_button.dart` | Full-width disabled button + spinner; replaces the Accept/Reject row entirely during `accepting` (so Reject is unavailable by omission, not just disabled) |
| `presentation/widgets/dispatch_status_banner.dart` | Failed/Timeout banner with a required `onRetry` — the spec always ends error handling in an explicit Retry, never a silent auto-revert |
| `presentation/widgets/trip_assigned_card.dart` | ✓ Trip Assigned, pickup/destination (via `TripAddressRow`), estimated fare, "Start Navigation" (`onNavigate` mock callback only) |
| `presentation/pages/dispatch_session_preview_page.dart` | New "Dispatch Session Preview" list (Accepting/Assigned/Failed/Timeout), linked from `TripOfferPreviewMenuPage`; renders via the same `TripOfferStatePreviewPage`, no repository call |

**Modified:** `driver_trip_offer_repository.dart` (+`acceptOffer`), `trip_offer_view.dart`
(+`_AcceptingContent`, `assigned`/`failed`/`timeout` cases, +`onNavigate`/`onRetry`
params), `trips_page.dart` (accept/retry/navigate handlers + a second dev
PopupMenuButton — "Accept outcome (dev)": Success/Failed/Timeout), `trip_offer_card.dart`
(uses the extracted `TripAddressRow`), `trip_offer_preview_menu_page.dart`
(now lists only the 3 base offer states + one "Dispatch Session Preview" tile,
rather than looping over every enum value).

### Race-condition guard (explicit requirement: "không được race condition")

`AnimatedSwitcher` keeps the outgoing widget mounted for ~350ms during its
crossfade — so the in-flight `newOffer` content (with its ticking
`CountdownIndicator`) is technically still alive for a moment after Accept is
pressed and the state has already moved to `accepting`. Rather than relying
solely on widget disposal timing, `TripsPage._handleExpired` guards with
`if (_state != TripOfferState.newOffer) return;` before transitioning to
`expired` — correct regardless of exactly when the old countdown widget gets
torn down.

### Test-writing lesson learned this phase (documented so it isn't relearned)

The offer's 15-second countdown runs a **continuously-ticking**
`AnimationController` for the entire time `TripOfferState.newOffer` is
showing. That makes `tester.pumpAndSettle()` **unsafe** during that window —
it keeps fast-forwarding simulated time trying to reach "nothing left
animating," and since the countdown never stops on its own before 15s, it
will run right through the whole thing and expire the offer out from under
the test. A fixed-duration `pump()` avoids that, but is itself a timing
gamble against a `PopupMenuButton`'s own open/close transition (a few
attempts at 300–400ms still produced a flaky hit-test miss on the popup
item). The reliable fix: for the new "Accept outcome (dev)" menu, tests grab
the `PopupMenuButton<DispatchAcceptStatus>` widget and call its `onSelected`
callback **directly**, bypassing the popup UI's own timing entirely — see
`_selectAcceptOutcome()` in `test/widget_test.dart`. Also fixed: one test
left a pending 1.2s `Future.delayed` timer running past the test's end
(flutter_test fails a test for that on its own), drained by pumping past the
delay before the test finishes.

**Verified this session:** `flutter analyze` — 0 issues (no warnings).
`flutter test` — 24/24 pass (16 from D-01–D-03, unaffected apart from the
`TripOfferStatus`→`TripOfferState` rename and the preview-menu test update
for the new 3-state + Dispatch-Session-tile layout, + 8 new: Accept→Assigned,
loading button disabled while accepting, countdown cannot fire after Accept,
Accept→Failed, Accept→Timeout, Retry→fresh New Offer, Assigned shows correct
data, onNavigate fires exactly once).

## Phase D-03 — Driver App: Incoming Trip Module (COMPLETE — 2026-07-07)

Reference: `docs/project/MVP_DEVELOPMENT_PLAN.md` Section 6 (Driver App
Roadmap, stage D4 — UI portion) and Section 12 (Phase Registry). UI only —
no backend, API, or protobuf changes. No state management library
introduced.

**New files under `apps/driver/lib/features/trips/`:**
| File | Purpose |
|------|---------|
| `domain/models/trip_offer_status.dart` | `TripOfferStatus` enum (newOffer/accepted/rejected/expired) |
| `domain/models/rider_info.dart` | `RiderInfo` — name, rating, avatar initial |
| `domain/models/trip_offer.dart` | `TripOffer` — rider, pickup/destination, distance-to-pickup, trip distance/duration/fare, nullable `surgeMultiplier` (`hasSurge` only true above 1.0x) |
| `domain/models/mock_trip_offer_catalog.dart` | Single shared sample offer, used by both the repository's "normal" mode and the preview pages (one source of truth for demo data) |
| `data/driver_trip_offer_repository.dart` | `DriverTripOfferDemoMode` enum (normal/empty/error) + `DriverTripOfferRepository.fetchOffer({mode, delay})` — `delay` is a real parameter (default 700ms), satisfying "configurable delay" literally |
| `presentation/widgets/rider_info_card.dart` | Avatar (initials) + name + rating |
| `presentation/widgets/trip_offer_card.dart` | Composes `RiderInfoCard` + pickup/destination rows + distance-to-pickup + surge badge (only rendered when `offer.hasSurge`) |
| `presentation/widgets/fare_estimate_card.dart` | Estimated fare (headline) + trip distance/duration |
| `presentation/widgets/countdown_indicator.dart` | `CountdownIndicator` — 15s `AnimationController`-driven circular countdown; turns red under 5s remaining; fires `onExpired` exactly once via a status listener when it completes |
| `presentation/widgets/trip_action_buttons.dart` | Accept/Reject row |
| `presentation/widgets/trip_offer_view.dart` | `TripOfferView` — content-only widget composing everything per `TripOfferStatus`, wrapped in an `AnimatedSwitcher` for cross-fade; reused by both the live `TripsPage` and the dev preview pages |
| `presentation/pages/trips_page.dart` | Rewritten (was the D-01 placeholder) — live flow: `AsyncStateView<TripOffer?>` for the fetch, then a local New Offer→Accepted/Rejected/Expired state machine; two AppBar actions: data dev-preview menu (Normal/Empty/Error) + a link to the trip-offer state preview menu |
| `presentation/pages/trip_offer_preview_menu_page.dart` + `trip_offer_state_preview_page.dart` | Same menu→single-state-preview pattern as `apps/rider`'s `TripPreviewMenuPage`/`TripStatePreviewPage` (R-02) and this app's own Home dev menu (D-02) |

**Design note:** `AsyncStateView`'s Empty state here means "no incoming offer right now" (`TripOffer?` is `null`) — distinct from the New Offer/Accepted/Rejected/Expired state machine, which only exists once an offer has actually been fetched successfully. These are two independent concerns (data availability vs. offer lifecycle), same pattern as Home's dashboard-fetch vs. availability-toggle split (D-02).

**Bug found and fixed during this phase (test-only, not app code):** the countdown-expiry test asserted `find.text('Offer Expired')` immediately after `tester.pump(const Duration(seconds: 16))`, but the `AnimatedSwitcher` cross-fade was still mid-transition at that exact frame (old "New Offer" content and the new "Offer Expired" banner briefly coexist). Fixed by adding `await tester.pumpAndSettle();` after the pump, before asserting — the same class of timing issue as R-04's overflow bugs, but this one was in the test, not the widget.

**Verified this session:** `flutter analyze` — 0 issues. `flutter test` —
16/16 pass (8 from D-01/D-02 + 8 new: Trips tab loads offer with countdown,
Accept → "Trip Accepted", Reject → "Trip Rejected", countdown reaching zero
auto-shows "Offer Expired", Empty/Error via the dev preview menu, the
preview menu lists all 4 states, one state preview page renders correctly).

## Phase D-02 — Driver App: Home Dashboard Module (COMPLETE — 2026-07-06)

Reference: `docs/project/MVP_DEVELOPMENT_PLAN.md` Section 6 (Driver App
Roadmap, stage D3 — UI portion) and Section 12 (Phase Registry). UI only —
no backend, API, or protobuf changes. No state management library
introduced.

**New files under `apps/driver/lib/`:**
| File | Purpose |
|------|---------|
| `shared/widgets/async_state_view.dart` | Hand-mirrored copy of `apps/rider`'s `AsyncStateView<T>` (Profile module, R-03) — no shared package between the two apps (D-01 decision), so re-implemented verbatim rather than imported |
| `features/home/domain/models/driver_availability_status.dart` | `DriverAvailabilityStatus` enum (offline/goingOnline/online/goingOffline) — drives the toggle's own transient states |
| `features/home/domain/models/driver_activity_status.dart` | `DriverActivityStatus` enum (offline/waitingForTrips/searchingNearby/busy) — drives the Home Status Card; an independent, finer-grained axis from availability (see design note below) |
| `features/home/domain/models/driver_home_summary.dart` | `DriverHomeSummary` — driver identity, nullable vehicle (drives the Empty state), today's stats |
| `features/home/data/driver_home_repository.dart` | `DriverHomeDemoMode` enum (normal/empty/error) + `DriverHomeRepository.fetchSummary()` — same dev-preview convention as Rider's Notification Center/Trip History |
| `features/home/presentation/widgets/driver_summary_header.dart` | Avatar/name/rating/vehicle (or "No vehicle assigned yet") |
| `features/home/presentation/widgets/home_stats_row.dart` | 3-column stat row: trips today, earnings today, online duration |
| `features/home/presentation/widgets/availability_toggle.dart` | The large primary switch — owns its own mock transition timing (~1.2s to go online, ~0.9s to go offline), reports every state change (incl. transient) to the parent page |
| `features/home/presentation/widgets/home_status_card.dart` | 4-message status card, cross-fades on change |
| `features/home/presentation/widgets/quick_action_card.dart` + `quick_actions_section.dart` | Reusable action card + the 2×2 grid of Earnings/Trip History/Support/Vehicle, all placeholder-nav (SnackBar) only |
| `features/home/presentation/pages/home_page.dart` | Rewritten (was the D-01 `PlaceholderTabContent` placeholder) — full dashboard, dev "Preview state" menu (Normal/Empty/Error data + Simulate/Clear busy) |

**Design note — two related but distinct state axes:** the task listed 4
toggle states (Offline/Going Online/Online/Going Offline) and 4 *different*
status-card messages (Offline/Waiting for trips/Searching nearby/Busy). These
aren't the same axis: going online settles into "Waiting for trips", then
auto-advances to "Searching nearby" after ~3s (simulating idle offer
polling) with no further user action — a real, if mock, animated state
progression. "Busy" has no natural trigger yet (needs an assigned trip —
Roadmap D4/D6, not built), so it's only reachable via the dev menu's
"Simulate busy" / "Clear busy" — consistent with how Rider's R-03/R-04 make
otherwise-unreachable states demoable.

**Bugs found and fixed during this phase:**
- `QuickActionCard`'s content (icon + label) overflowed its grid cell by
  ~12px — `childAspectRatio: 2.4` made cells too short for the content at
  the tested viewport width. This is a genuine responsive-layout bug (not
  test-only): fixed by lowering the ratio to `1.6` (taller cells), which
  also fixed a secondary hit-test-miss warning on the overflowing tile's
  text (the overflow was corrupting layout enough to throw off tap
  targeting too).
- One D-01 test (`DriverApp shows Home tab and all 5 bottom nav
  destinations`) started failing once this phase added an "Earnings" Quick
  Action label, which duplicated the existing "Earnings" nav destination
  label. Not a bug — tightened that assertion (and the others alongside it)
  to `find.widgetWithText(NavigationDestination, ...)` so it specifically
  checks the nav bar, unaffected by future same-named labels elsewhere on
  the page.

**Verified this session:** `flutter analyze` — 0 issues. `flutter test` —
8/8 pass (3 from D-01, tightened as above + 5 new: dashboard loads
summary/stats, toggle walks through all 4 states with the status card
advancing in sync (Offline → Going Online → Online/Waiting → Searching →
Going Offline → Offline), Empty state (no vehicle) via dev menu, Error state
via dev menu, Quick Action tap shows its placeholder message).

## Phase D-01 — Driver App: Initial Project Scaffold (COMPLETE — 2026-07-06)

Reference: `docs/project/MVP_DEVELOPMENT_PLAN.md` Section 6 (Driver App
Roadmap, stage D1) and Section 12 (Phase Registry). UI only — no backend,
API, or protobuf changes. No state management library introduced. **First
Driver App phase — `apps/driver` did not exist before this session.**

**Scaffolding method:** ran `flutter create --org com.fairride --project-name
driver apps/driver` (same approach `apps/rider` used in Phase 12) rather than
hand-writing `android/`/`ios/` platform files. Bundle ID: `com.fairride.driver`.
Bumped `flutter_lints` to `^6.0.0` (generator default was `^5.0.0`) to match
`apps/rider`; SDK constraint (`^3.9.2`) already matched the installed
toolchain, so — unlike R-01 — no version-mismatch fix was needed here.

**Architecture decision (Roadmap D1 explicitly calls for one):** `apps/driver`
is a fully separate Flutter project with its own `pubspec.yaml` — **no shared
Dart package with `apps/rider`.** The design system (typography scale,
button/card/input shapes) is hand-mirrored into Driver's own `AppTheme`
rather than imported, with a distinct deep-orange accent (`#EF6C00`) replacing
Rider's green (`#1A8C4E`) per the branding requirement. This means any future
Rider design-system change must be manually ported to Driver's `AppTheme` too
— a real cost, but this task's explicit ask was "decide," so this is
recorded as a deliberate choice, not an oversight.

**New files under `apps/driver/lib/`:**
| File | Purpose |
|------|---------|
| `main.dart` / `app.dart` | Entry point — `DriverApp`, `MaterialApp.router`, same shape as `apps/rider`'s `main.dart`/`app.dart` |
| `core/theme/app_theme.dart` | Driver's `AppTheme` — same structure as Rider's (incl. `CardThemeData`, not `CardTheme`, learned from the R-01 fix), distinct accent color |
| `core/router/app_router.dart` | `go_router` with named routes (`name: 'home'` etc.) + `StatefulShellRoute.indexedStack`, 5 branches |
| `shared/widgets/scaffold_with_nav.dart` | 5-destination bottom nav shell (Home/Trips/Earnings/Notifications/Profile) |
| `shared/widgets/placeholder_tab_content.dart` | Reusable placeholder body (icon/title/subtitle), used by 4 of the 5 tabs |
| `features/home\|trips\|earnings\|notifications/presentation/pages/*_page.dart` | Simple placeholder pages — no `domain`/`data` yet since there is no data to layer for a placeholder |
| `features/profile/domain/models/app_info.dart` | `AppInfo` model + `BuildModeKind` enum — app version is mock, build mode is **real** (`kReleaseMode`/`kProfileMode`/`kDebugMode`), Flutter version/environment are explicit placeholders |
| `features/profile/data/mock_app_info_repository.dart` | `MockAppInfoRepository` — the "data" layer requested by this phase's folder-structure requirement |
| `features/profile/presentation/widgets/info_row.dart` | Reusable label/value row |
| `features/profile/presentation/pages/profile_page.dart` | Placeholder + one real "Developer" entry point |
| `features/profile/presentation/pages/developer_page.dart` | Development Utilities screen — the one fully "real" (if mock-data-backed) screen in this phase |

**Folder structure note:** only the Profile feature has the full
`domain/data/presentation/widgets` layering this phase's requirements
describe — it's the one feature with actual data (`AppInfo`). The 4 simple
placeholder tabs only have `presentation/pages/`, matching how `apps/rider`
itself started (Phase 12: placeholder-only pages, domain layers added only
when real data arrived in later phases). Profile's layering is the pattern
future Driver phases (D2+) should follow for Home/Trips/Earnings/Notifications
once they have real data.

**Bug found and fixed during this phase:** `InfoRow`'s `Row` had no
`Expanded` around the value `Text`, so the long placeholder strings (e.g.
"3.35.4 (placeholder — not read at runtime)") overflowed on the Developer
page — same class of bug as `ReceiptRow` in R-04, fixed the same way
(`Expanded` + right-aligned wrapping text).

**Verified this session:** `flutter pub get` — succeeds. `flutter analyze` —
0 issues. `flutter test` — 3/3 pass (tab shell + all 5 destinations render,
tab switching works, Developer page opens with all 4 diagnostic fields).
All three commands run independently from `apps/rider`'s — separate Flutter
projects, separate toolchain runs.

## Phase R-04 — Rider App: Ride History Module (COMPLETE — 2026-07-06)

Reference: `docs/project/MVP_DEVELOPMENT_PLAN.md` Section 5 (Rider App Roadmap,
new stage R9 — not in the original roadmap, added this phase) and Section 12
(Phase Registry). UI only — no backend, API, or protobuf changes. No state
management library introduced.

**New files under `apps/rider/lib/features/history/`:**
| File | Purpose |
|------|---------|
| `domain/models/trip_history_status.dart` | `TripHistoryStatus` enum (completed/cancelled) + label/color/icon — distinct from `RiderTripStatus` (live trip phases in `features/trip`) |
| `domain/models/trip_history_filters.dart` | `TripHistoryStatusFilter` (All/Completed/Cancelled) + `TripHistoryDateFilter` (All time/Today/This Week/This Month) with `.matches(DateTime)` |
| `domain/models/trip_timeline_event.dart` | `TripTimelineEvent` (label/time/icon) for the Trip Detail timeline |
| `domain/models/trip_history_entry.dart` | `TripHistoryEntry` — composed almost entirely from earlier phases' types: `TripSelection` (map/R-01), `MockDriver` (trip/R-02), `MockFareBreakdown` + `PaymentMethod` + `VehicleCategory` (booking/R-01). Plus `matchesQuery()` extension for the search box. |
| `domain/models/mock_trip_history_catalog.dart` | 5 sample trips spanning today/this week/this month/45 days ago (so every date filter has both a match and a non-match), mixed completed/cancelled |
| `domain/models/mock_trip_history_repository.dart` | `TripHistoryDemoMode` enum (normal/empty/error) + `MockTripHistoryRepository.fetchHistory()` — same dev-preview convention as Notification Center (R-03) |
| `presentation/widgets/status_chip.dart` | Reusable Completed/Cancelled chip |
| `presentation/widgets/trip_history_tile.dart` | Compact list row (driver+vehicle, status, pickup/destination, time, rating, fare) — deliberately its own compact widget rather than `PickupCard`/`DestinationCard` (those are sized for detail screens, not a dense list); still built on the same `TripSelection` data |
| `presentation/widgets/trip_history_section_header.dart` | Date-group header + shared `labelFor()` day-bucketing logic ("Today"/"Yesterday"/"MMM d, yyyy") |
| `presentation/widgets/history_filter_bar.dart` | Search field + status/date `ChoiceChip` rows, all client-side |
| `presentation/widgets/vehicle_info_card.dart` | Dedicated "Vehicle information" block (icon/category/model/plate) — kept separate from `DriverInfoCard` per the phase's explicit item list |
| `presentation/widgets/distance_duration_card.dart` | Distance/duration stat row, styled like `EtaArrivalCard`/`ProfileStatsRow` |
| `presentation/widgets/trip_timeline.dart` | Vertical milestone timeline (Requested → Assigned → Started → Completed, or → Cancelled) |
| `presentation/widgets/payment_method_row.dart` | Read-only payment display — reuses the `PaymentMethod` *model* but not `PaymentMethodCard` (that widget opens a picker, wrong affordance for a past trip) |
| `presentation/widgets/receipt_row.dart` | `ReceiptRow` (label/value, monospace) + `ReceiptDivider` (dashed rule) — Receipt is deliberately styled distinct from `FareSummaryCard`, not reusing that widget |
| `presentation/pages/trip_history_page.dart` | Trip History screen — `AsyncStateView<List<TripHistoryEntry>>`, dev "Preview state" menu, grouped-by-day list, in-place "no trips match these filters" message distinct from the repository-level Empty state |
| `presentation/pages/trip_detail_page.dart` | Trip Detail — reuses `PickupCard`/`DestinationCard`/`FareSummaryCard` (booking) + `DriverInfoCard` (trip) directly |
| `presentation/pages/receipt_page.dart` | Receipt — Trip ID/rider (`MockRiderProfileCatalog.sample.fullName`, reused from Profile)/driver/vehicle/date/payment/fare breakdown/mock 8% tax/total |
| `history.dart` | Barrel export |

**Integration point:** added a "Trip History" tile to `ProfilePage`, directly below the existing "Settings" tile (same style) — the original Phase-12 settings list had a "Trip History" entry that R-03 dropped when it replaced the list with the required 9-item set; this restores that access point on the module that now actually implements it.

**Bugs found and fixed during this phase (not just pre-existing toolchain noise, like R-01/R-03):**
- `TripHistoryPage._load()` called `setState(() => _future = ...)` — an arrow body, so the closure evaluated to (returned) the assigned `Future`, which throws `setState() callback argument returned a Future`. Fixed by switching to a block body `setState(() { _future = ...; })`, matching the pattern already used in `NotificationCenterPage`.
- `ReceiptRow` had no `Expanded`/`Flexible` around its value `Text`, so a long combined value (e.g. "Toyota Vios (51G-123.45)") overflowed the row on narrow widths — a real responsive-layout bug, not just a test artifact. Fixed by wrapping the value in `Expanded` with right-aligned, wrapping text.

**Verified this session:** `flutter analyze` — 0 issues. `flutter test` — 15/15 pass (9 from R-01/R-02/R-03 + 6 new: Trip History loads/groups by day, Completed filter hides cancelled trips, Empty state via dev menu, Error state via dev menu, Trip Detail shows route/driver/timeline/fare, Receipt shows trip ID/rider/total).

## Phase R-03 — Rider App: Profile Module (COMPLETE — 2026-07-06)

Reference: `docs/project/MVP_DEVELOPMENT_PLAN.md` Section 5 (Rider App Roadmap,
stage R8 — UI portion only) and Section 12 (Phase Registry). UI only — no
backend, API, or protobuf changes. No state management library introduced.

**New files under `apps/rider/lib/features/profile/`:**
| File | Purpose |
|------|---------|
| `domain/models/rider_profile.dart` | `MemberLevel` enum (standard/silver/gold/platinum, mock only) + `RiderProfile` model |
| `domain/models/mock_profile_repository.dart` | `MockProfileRepository.fetchProfile({simulateError})` — simulated delay, sample data |
| `domain/models/settings_entry.dart` | `SettingsAction` enum (9 required + `developerPreview`) + `SettingsEntry` model + `MockSettingsCatalog` (grouped: account/preferences/support/logout) |
| `domain/models/notification_item.dart` | `NotificationType` enum + `NotificationItem` model (`copyWith`, `relativeTimeLabel`) |
| `domain/models/mock_notification_repository.dart` | `NotificationDemoMode` enum (normal/empty/error) + `MockNotificationRepository` + `MockNotificationCatalog.sample()` |
| `presentation/widgets/async_state_view.dart` | Generic `AsyncStateView<T>` — Loading/Success/Empty/Error wrapper around a `Future`, animated cross-fade between states. Reused by both Profile and Notification Center. |
| `presentation/widgets/profile_header.dart` | Avatar (initials) + name + phone + member level badge |
| `presentation/widgets/profile_stats_row.dart` | Rating + total completed trips, side by side |
| `presentation/widgets/settings_tile.dart` | Reusable settings row (icon/label/chevron, destructive variant for Logout) |
| `presentation/widgets/settings_section.dart` | Labelled (or unlabelled) group of `SettingsTile`s in a card |
| `presentation/widgets/developer_preview_section.dart` | Carries the R-02 "Trip UI Preview (dev)" entry point into the new Settings screen, in its own Developer section |
| `presentation/widgets/unread_badge.dart` | Reusable small count badge, animated in/out |
| `presentation/widgets/notification_tile.dart` | Single notification row — animated read/unread background, unread dot |
| `presentation/pages/profile_page.dart` | Rewritten Profile Screen — `AsyncStateView<RiderProfile>`, bell icon + `UnreadBadge`, "Settings" entry point |
| `presentation/pages/settings_page.dart` | Settings Screen — 3 grouped sections (Account/Preferences/Support) + Developer section + standalone Logout; every non-Notifications/Logout tap shows an explicit placeholder message |
| `presentation/pages/notification_center_page.dart` | Notification Center — `AsyncStateView<List<NotificationItem>>`, mark-one/mark-all read, dev "Preview state" menu (Normal/Empty/Error) |
| `profile.dart` | Barrel export |

**Integration points:**
- `ProfilePage` is a `StatefulWidget` now (was `StatelessWidget`) — needed for the mock profile fetch and unread-count state. Router usage unchanged (`const ProfilePage()`, no params), so `app_router.dart` needed no edit this phase.
- Notification unread count flows back to `ProfilePage` via `Navigator.push<int>` / `Navigator.pop(context, unreadCount)` — plain Navigator result passing, not a new state layer.
- The old private `_SettingsTile`/inline settings list from Phase 12 is fully replaced by the new reusable `SettingsTile`/`SettingsSection` components.

**Scoping decision (documented, not a shortcut):** Empty/Error states are demonstrated concretely in the Notification Center (where they're meaningful, via the dev preview menu) rather than forced onto the Profile screen, which only has a genuine Loading→Success path. `AsyncStateView` supports all four generically either way.

**Verified this session:** `flutter analyze` — 0 issues. `flutter test` — 9/9 pass (4 from R-01/R-02 + 5 new: Profile loading→success, Settings lists all 9 entries, Notifications shows mock list, Notifications empty state via dev menu, Notifications error state via dev menu). One test-only fix needed: `ListView` virtualises off-screen children even with a plain `children:` list, so the Settings test grows the test viewport (`tester.view.physicalSize`) rather than asserting on real scroll gestures.

## Phase R-02 — Rider App: Trip Lifecycle UI (COMPLETE — 2026-07-06)

Reference: `docs/project/MVP_DEVELOPMENT_PLAN.md` Section 5 (Rider App Roadmap,
stage R5 — UI portion only) and Section 12 (Phase Registry). UI only — no
backend, API, or protobuf changes. No state management library introduced.

**New files under `apps/rider/lib/features/trip/`:**
| File | Purpose |
|------|---------|
| `domain/models/rider_trip_status.dart` | `RiderTripStatus` enum (5 states) + extension: label, short label, status message, icon, progress value, `isCancellable`/`hasDriver`/`showsEta`/`showsSafetyActions` flags |
| `domain/models/mock_driver.dart` | `MockDriver` model (name, vehicle model, plate, rating, initials for avatar) |
| `domain/models/mock_trip_catalog.dart` | Sample driver + mock ETA/arrival-time helpers per status |
| `domain/models/mock_trip_repository.dart` | `MockTripRepository.watchLifecycle()` — async-generator `Stream<RiderTripStatus>`, the "mock repository" driving the timed demo; no HTTP |
| `presentation/widgets/driver_info_card.dart` | Avatar (initials) + name + rating + vehicle/plate |
| `presentation/widgets/trip_progress_indicator.dart` | 5-stage animated progress bar + labels |
| `presentation/widgets/trip_status_banner.dart` | Icon + animated status message |
| `presentation/widgets/eta_arrival_card.dart` | ETA + estimated arrival time |
| `presentation/widgets/cancel_ride_button.dart` | Confirm dialog + mock cancel (no API call) |
| `presentation/widgets/emergency_button.dart` | Placeholder — snackbar explains it's not implemented |
| `presentation/widgets/contact_driver_button.dart` | Placeholder — snackbar explains it's not wired |
| `presentation/pages/searching_driver_view.dart` | Searching Driver — pulsing icon animation, Cancel Ride |
| `presentation/pages/driver_assigned_view.dart` | Driver Assigned — driver info, ETA, Contact/Emergency, Cancel Ride |
| `presentation/pages/driver_arriving_view.dart` | Driver Arriving — same layout, shorter ETA |
| `presentation/pages/trip_in_progress_view.dart` | Trip In Progress — no Cancel Ride (trip already started) |
| `presentation/pages/trip_completed_view.dart` | Trip Completed — reuses Booking's `FareSummaryCard`; Done button; no Cancel/Contact/Emergency |
| `presentation/pages/trip_lifecycle_page.dart` | `TripLifecyclePage` — persistent Scaffold, `AnimatedSwitcher` (fade+slide) across the 5 views, driven by `MockTripRepository` |
| `presentation/pages/trip_state_preview_page.dart` | Renders one state in isolation with its own Scaffold — the "independently previewable" entry point |
| `presentation/pages/trip_preview_menu_page.dart` | Lists all 5 states, navigates to `trip_state_preview_page.dart` for each |
| `trip.dart` | Barrel export |

**Integration points (minimal, additive, reusing the Booking UI as instructed):**
- `features/booking/presentation/widgets/booking_form_body.dart` — `_handleBookRide` now pushes `TripLifecyclePage` (with the trip selection + fare captured at booking time) after the mock delay/snackbar, instead of just showing a snackbar
- `features/profile/presentation/pages/profile_page.dart` — added a "Trip UI Preview (dev)" settings tile (required adding an `onTap` param to the existing private `_SettingsTile`) navigating to `TripPreviewMenuPage`
- Every trip view reuses `PickupCard`/`DestinationCard` from the Booking module; `TripCompletedView` also reuses `FareSummaryCard` + `MockFareBreakdown`

**Explicitly NOT done (by design — matches Roadmap R5's "real wiring" split):** no live driver location, no real Dispatch/Trip status polling, `MockTripRepository` runs on a fixed local timer only. Cancel/Contact Driver/Emergency are all mock placeholders that clearly tell the user so (snackbar/dialog text), never silently pretending to work.

**Verified this session:** `flutter analyze` — 0 issues. `flutter test` — 4/4 pass (1 from R-01 + 3 new: preview menu lists all 5 states, Driver Assigned preview shows driver info/buttons, Trip Completed preview shows fare + hides Cancel Ride).

## Phase R-01 — Rider App: Booking UI Module (COMPLETE — 2026-07-06)

Reference: `docs/project/MVP_DEVELOPMENT_PLAN.md` Section 5 (Rider App Roadmap,
stages R1 + R1b) and Section 12 (Phase Registry). UI only — no backend, API, or
protobuf changes. No state management library introduced (plain
`StatefulWidget`/`setState`, consistent with the rest of `apps/rider`).

**New files under `apps/rider/lib/features/booking/`:**
| File | Purpose |
|------|---------|
| `domain/models/vehicle_option.dart` | `VehicleCategory` enum + `VehicleOption` model (rates mirror `backend/services/pricing` `DefaultFareConfig` shape — mock only, no import) |
| `domain/models/payment_method.dart` | `PaymentMethodType` enum + `PaymentMethod` model |
| `domain/models/promo_result.dart` | `PromoResult` + `MockPromoValidator` (hardcoded codes `FAIRRIDE10`, `WELCOME20`) |
| `domain/models/mock_trip_metrics.dart` | Haversine distance + rough duration estimate between `TripSelection` pickup/destination — client-side only, no routing API |
| `domain/models/mock_fare_calculator.dart` | `MockFareBreakdown.calculate()` — mirrors the backend Pricing formula shape (`max(base+distance+time, minimum) + bookingFee`) |
| `domain/models/mock_booking_catalog.dart` | Central mock data: vehicle list, payment methods, `sampleTripSelection` fallback |
| `presentation/widgets/trip_point_cards.dart` | `PickupCard` + `DestinationCard` |
| `presentation/widgets/vehicle_selector.dart` | Horizontally scrollable ride-option selector, animated selection |
| `presentation/widgets/fare_summary_card.dart` | Itemised fare breakdown, animated total (`TweenAnimationBuilder`) |
| `presentation/widgets/payment_method_card.dart` | Selected method + modal picker |
| `presentation/widgets/promo_code_entry.dart` | Promo text entry, shake-on-invalid / check-on-valid animations |
| `presentation/widgets/book_ride_button.dart` | CTA with idle → loading → success animated states (mock `Future.delayed`, no API call) |
| `presentation/widgets/booking_form_body.dart` | Composes all of the above; shared by `BookingPage` and `BookingBottomSheet` |
| `presentation/widgets/booking_bottom_sheet.dart` | Modal `DraggableScrollableSheet` wrapper around `BookingFormBody` |
| `presentation/pages/booking_page.dart` | Rewritten (was Phase-12 static placeholder) — full-page host for `BookingFormBody` |
| `booking.dart` | Barrel export (page + bottom sheet), matching the `core/location/location.dart` convention |

**Integration points (minimal, additive):**
- `core/router/app_router.dart` — Booking route now reads an optional `TripSelection` via `state.extra`
- `features/map/presentation/pages/map_page.dart` — added a "Book this ride" button to the confirmed pickup/destination panel, opening `BookingBottomSheet` with the real `TripSelection`; no change to the existing selection state machine

**Toolchain verification (Roadmap R1, previously blocked):**
- This machine has a working Flutter 3.35.4 / Dart 3.9.2 install. `pubspec.yaml`'s `environment.sdk` constraint (`^3.12.2`, set on a machine with a newer/different toolchain) was lowered to `^3.9.2` to match — otherwise `flutter pub get` fails outright with a version-solving error.
- Fixed to get `flutter analyze` clean: `AppTheme` — `ThemeData.cardTheme` needed `CardThemeData` (not `CardTheme`) on this SDK; `vehicle_selector.dart` — `unnecessary_underscores` lint (`(_, __)` → `(_, _)`); `test/widget_test.dart` — was still the unmodified `flutter create` counter-app template referencing a nonexistent `MyApp` class (real root widget is `RiderApp`), replaced with a real smoke test rendering `BookingPage`.
- **Verified this session:** `flutter pub get` — succeeds (14 packages have newer versions available, not upgraded). `flutter analyze` — 0 issues. `flutter test` — 1/1 pass.

**Explicitly NOT done (by design — see MVP_DEVELOPMENT_PLAN.md Section 9):** no API client, no real fare/booking/payment/promo backend calls, no auth. All data is mock. Real wiring is Roadmap stage R4 (booking submission) / R6 (payment), tracked separately.

## MVP Development Plan (NEW — 2026-07-06, revised v1.1 same day after team review)
`docs/project/MVP_DEVELOPMENT_PLAN.md` created — planning-only document, no code/API/proto changes.
This is now the SSOT for MVP implementation ordering and supersedes ad-hoc phase sequencing
until DOC-0004 (Implementation Master Plan) is generated. Key conclusions from the repo audit
that produced it:
- Identity has no register/login use case or API surface — nothing device-facing can authenticate
  yet. **v1.1 correction:** this blocks Integration Testing / End-to-End (Milestone M4) only, NOT
  Rider/Driver UI construction. Both UI tracks build against a MockAuth stub in parallel with real
  Identity work (Section 9 Parallel Development Matrix, Section 10 tree: Foundation → {Backend
  Core, Rider UI, Driver UI, Identity Integration} → Integration → Beta). MockAuth's interface
  must mirror the real auth contract exactly (Section 14 Rule 5) so the swap at M4 is a config
  change, not a rewrite.
- CI only runs tests for `backend/shared`; per-service test suites (~460 tests across identity,
  user, driver, trip, dispatch, pricing, booking, gateway) do not run in CI. CI build matrix is
  also missing `booking` and `gateway`.
- `apps/driver` and any admin web project do not exist yet.
- Section 12 of the doc ("Phase Registry") is the forward-looking phase tracker going forward;
  historical phases (1–17, H2–H4) remain recorded here and in CHANGELOG.md, not duplicated there.
- Section 9 ("Parallel Development Matrix") is the fast day-to-day reference: what can be worked
  on right now vs. what's actually blocked, per module.
See the document itself for full detail: Current Project Status, Dependency Graph, Parallel
Development Matrix, Recommended Development Order, per-app roadmaps, Milestones, and Risks.

## Documentation Strategy Change
ORIGINAL: 72-document comprehensive roadmap
CURRENT: Lean Documentation — 5 permanent core documents + on-demand as needed
DECISION DATE: 2026-06-30
ADR REQUIRED: ADR-0008 (pending)
The 72-document plan from PHASE-0-EOS-BLUEPRINT.md is superseded as the primary delivery plan.
Documents from that plan remain available as on-demand resources, generated when a specific
implementation phase requires them.

## EOS Core Documents Progress
| Doc ID | Title | Status | Path |
|--------|-------|--------|------|
| DOC-0001 | Project Constitution | Draft — Awaiting CTO Approval | docs/business/mission/project-constitution-v0.1.md |
| DOC-0001A | AI Development Governance | Draft — Awaiting CTO Approval | .ai/governance/ai-development-governance-v0.1.md |
| DOC-0002 | Product Vision | Draft — Awaiting CTO+CPO Approval | docs/product/requirements/product-vision-v0.1.md |
| DOC-0003 | System Architecture | Not yet generated — BLOCKED | docs/architecture/system/ |
| DOC-0004 | Implementation Master Plan | Not yet generated — BLOCKED | docs/product/roadmap/ |

## Open Questions — Priority View

### P0 Critical (must resolve before DOC-0003)
| OQ ID | Question |
|-------|----------|
| OQ-001 | Primary launch city and country |
| OQ-006 | Driver classification model (contractor vs employee) |
| OQ-B01 | Cash collection at MVP? (market may require Day 1) |

### P1 High (must resolve before DOC-0004)
| OQ ID | Question |
|-------|----------|
| OQ-002 | Funding stage and runway at MVP |
| OQ-007 | Offline-capable dispatch required at MVP? |
| OQ-B02 | First-ride Rider incentive amount |
| OQ-B03 | Driver 0% commission early adopter programme scope |
| OQ-B04 | MVP ride categories (motorcycle taxi Day 1?) |
| OQ-B05 | Booking fee amount |
| OQ-B06 | Instant withdrawal fee and mechanism |
| OQ-B07 | Map provider requirement at MVP |

### P2 Medium
OQ-003, OQ-005, OQ-A001, OQ-A002, OQ-A003, OQ-A004, OQ-A005, OQ-B08

### P3 Low
OQ-004, OQ-008, OQ-009

## Pending ADRs (16 total)
ADR-0001 (Go), ADR-0002 (Flutter), ADR-0003 (Next.js), ADR-0004 (doc-first), ADR-0005 (microservices),
ADR-0006 (event-driven), ADR-0007 (ai-first), ADR-0008 (lean-docs — P0), ADR-0009 (launch-market — P0),
ADR-0010 (commission-model — P0), ADR-0011 (surge-cap — P0), ADR-0012 (surge-passthrough — P1),
ADR-0013 (driver-tier-calc — P1), ADR-0014 (mvp-payments — P1), ADR-0015 (north-star — P1),
ADR-0016 (driver-classification — P0)

## Key Product Decisions Documented in DOC-0002
| Decision | Value | Section |
|----------|-------|---------|
| MVP Commission Rate | 15% flat | 6.7 |
| Commission Tiers | Standard 15% → Platinum 10% | 6.7 |
| Surge Cap | 2.0x hard cap | 6.9 |
| Surge Revenue to Driver | 100% of surge to Driver | 6.7 |
| North Star Metric | Weekly Fair Matches (WFM) | 6.23 |
| MVP Supply Target | 1,000 active Drivers by Month 3 | 6.19 |
| MVP Trip Target | 5,000 completed trips/day by Month 6 | 6.19 |
| Rider RSAT Target | ≥ 80% | 6.22 |
| Platform Availability | ≥ 99.9% | 6.19 |
| Dispute Resolution SLA | 95% within 24 hours | 6.22 |

## Phase 1 Implementation — Skeleton (COMPLETE)

| Deliverable | Path | Status |
|------------|------|--------|
| Go workspace | `backend/go.work` | ✅ |
| Shared package (config/logger/errors/db/redis/kafka/grpc) | `backend/shared/` | ✅ tested |
| 14 service skeletons | `backend/services/*/` | ✅ all compile |
| Docker Compose (postgres/redis/kafka) | `infra/docker/` | ✅ |
| Makefile | `Makefile` | ✅ |
| CI workflow | `.github/workflows/ci.yml` | ✅ |
| Service Dockerfile | `backend/Dockerfile` | ✅ |

## Phase 1.5 — Architecture Audit Fixes (COMPLETE)

All Critical and High findings from the audit have been resolved.

| Finding | Severity | Fix Applied | File |
|---------|----------|-------------|------|
| C-001: gRPC reflection always on in production | Critical | `Options.EnableReflection` + gated in `bootstrap.Run()` | `shared/grpc/server.go`, `shared/server/bootstrap.go` |
| C-002: CI test command `./...` broken | Critical | `working-directory: backend/shared` + `go test ./...` | `.github/workflows/ci.yml` |
| C-003: readiness probe always returns 200 | Critical | Real `ReadinessTracker` with 503 on not-ready | `shared/server/bootstrap.go` |
| H-001: `max()` shadow of Go 1.21+ builtin | High | Removed function, inlined logic | `shared/redis/client.go` |
| H-002: kafka-ui pinned to `latest` | High | Pinned to `v0.7.2` | `infra/docker/docker-compose.yml` |
| H-003: DB pool exhaustion (25 × 14 = 350 > 100) | High | MaxConns default 25→5, MinConns 5→2 | `shared/config/config.go` |
| H-004: gRPC msg limits not wired from config | High | `MaxRecvMsgSizeMB`/`MaxSendMsgSizeMB` passed via `Options` | `shared/grpc/server.go`, `shared/server/bootstrap.go` |
| H-005: `kafka.Message` type leak across package boundary | High | Unexported `raw` field; `FetchMessage`/`Commit` use `*Message` | `shared/kafka/consumer.go` |
| H-006: CI build command `./services/...` broken | High | Changed to `go build github.com/fairride/$svc/...` | `.github/workflows/ci.yml` |
| M-002: Docker init/postgres directory missing | Medium | Added `.gitkeep` | `infra/docker/init/postgres/.gitkeep` |
| M-004: Dockerfile buildVersion uses `$(date)` (breaks cache) | Medium | Changed to `ARG GIT_COMMIT`; passed at build time | `backend/Dockerfile` |
| M-005: No linter configuration | Medium | Added `.golangci.yml` | `.golangci.yml` |
| M-007: 14 identical main.go files (90 lines each) | Medium | `server.Run()` bootstrap; all 14 now 7 lines each | `shared/server/bootstrap.go` + all service `main.go` |
| L-008: Kafka ACL — no RequireAllAcks for financial | Low | Added `RequireAllAcks bool` to `ProducerConfig` | `shared/kafka/producer.go` |

**Deferred to later phases (medium/low, no immediate risk):**
- M-001: Type aliases → interfaces for DB/Redis (Phase 2+)
- M-003: Correlation/trace ID in logging interceptor (Phase 2+)
- M-006: Database migration tooling (Phase 2.1)

**Build verification:** all 14 services compile. `go test -race ./...` passes (3 packages tested, 0 failures).

## Audit Score (Phase 1.5 exit)
| Category | Score Before | Score After |
|----------|-------------|-------------|
| Security | 65/100 | 95/100 |
| Reliability | 55/100 | 90/100 |
| Code Quality | 75/100 | 92/100 |
| DevOps | 60/100 | 90/100 |
| Overall | 64/100 | **92/100** |

**Verdict: GO for Phase 2**

## Phase 2.1 — Identity Foundation (COMPLETE)

Clean Architecture domain layer for the Identity service. No DB, no endpoints, no auth.

| File | Purpose |
|------|---------|
| `services/identity/domain/entity/permission.go` | Permission entity + 23 named permission constants + `NewPermission` / `ReconstitutePermission` |
| `services/identity/domain/entity/role.go` | Role entity + 6 system role constants + `NewRole` / `ReconstituteRole`; `AddPermission`, `RemovePermission`, `HasPermission`, `CanDelete` |
| `services/identity/domain/repository/permission_repository.go` | `PermissionRepository` interface (FindByID, FindByName, FindByResource, FindAll, Save, Delete) |
| `services/identity/domain/repository/role_repository.go` | `RoleRepository` interface (FindByID, FindByName, FindAll, Save, Delete) |
| `services/identity/app/container.go` | Composition root; holds `RoleRepository` + `PermissionRepository` interfaces |
| `services/identity/domain/entity/permission_test.go` | 8 unit tests for Permission |
| `services/identity/domain/entity/role_test.go` | 17 unit tests for Role |

**System roles (from DOC-0002 §6.12):** rider, driver, fleet_operator, city_manager, support_agent, super_admin

**Permission format:** `"resource:action"` — validated at construction. 23 well-known constants (trips, drivers, riders, wallet, payments, dispatch, reviews, reports, support, admin).

**Build verification:** all 14 services compile. 25 identity entity tests pass. 3 shared packages pass.

## Phase 2.2 — Identity Persistence (COMPLETE)

PostgreSQL repository implementations for Role and Permission.

| File | Purpose |
|------|---------|
| `services/identity/infrastructure/postgres/permission_repository.go` | `PermissionRepository`: FindByID, FindByName, FindByResource, FindAll, Save (upsert), Delete |
| `services/identity/infrastructure/postgres/role_repository.go` | `RoleRepository`: FindByID, FindByName, FindAll, Save (upsert + tx permission replace), Delete |
| `services/identity/infrastructure/postgres/testmain_test.go` | TestMain: skip when DATABASE_URL unset; createSchema/dropSchema/setupTest helpers |
| `services/identity/infrastructure/postgres/permission_repository_test.go` | 10 integration tests |
| `services/identity/infrastructure/postgres/role_repository_test.go` | 13 integration tests |

**Schema (test scaffolding — NOT a migration):**
- `identity_permissions` (id PK, name UNIQUE, resource, action, description, created_at)
- `identity_roles` (id PK, name UNIQUE, description, is_system, created_at, updated_at)
- `identity_role_permissions` (role_id FK, permission_id FK, PK composite)

**go.mod changes (identity service):**
- Added `require github.com/jackc/pgx/v5 v5.6.0`
- Added `replace github.com/fairride/shared => ../../shared` (for `GOWORK=off go mod tidy`; no effect in workspace mode)

**Verification:**
- `go build -o /dev/null github.com/fairride/identity/...` — ✅ clean
- `go vet github.com/fairride/identity/...` — ✅ clean
- `go test github.com/fairride/identity/infrastructure/postgres/...` (no DATABASE_URL) — ✅ skips gracefully
- Integration tests (need DB): run with `DATABASE_URL=... go test -v github.com/fairride/identity/infrastructure/postgres/...`
- All 14 service builds still pass

## Phase 2.3 — User Domain Model (COMPLETE)

Pure domain layer for the User entity. No DB, no auth, no infrastructure.

| File | Purpose |
|------|---------|
| `services/identity/domain/entity/user.go` | `UserType` enum (Rider/Driver/FleetOperator/Admin), `UserStatus` enum (PendingVerification/Active/Suspended/Deactivated), `User` struct, `NewUser` (validated), `ReconstituteUser` (no validation), `Activate`/`Suspend`/`Deactivate` status transition methods |
| `services/identity/domain/entity/user_test.go` | 27 unit tests covering construction, email validation, all status transitions, error cases, full lifecycle |
| `services/identity/domain/repository/user_repository.go` | `UserRepository` interface (FindByID, FindByPhone, FindAll, Save, Delete) |

**Business rules enforced:**
- Phone and name: non-empty (whitespace-only rejected)
- UserType: one of TypeRider, TypeDriver, TypeFleetOperator, TypeAdmin
- Email: optional (empty allowed); if present must be `local@domain.tld` format
- New users always start as `StatusPendingVerification`
- `Activate`: PendingVerification or Suspended → Active; others fail CodePreconditionFailed
- `Suspend`: Active → Suspended; others fail CodePreconditionFailed
- `Deactivate`: Active or Suspended → Deactivated (terminal); PendingVerification and already-Deactivated fail
- `ReconstituteUser`: no validation, accepts any status (DB rehydration)

**Verification:**
- `go test -race github.com/fairride/identity/domain/entity/...` — ✅ 52 tests pass (27 new + 25 existing)
- `go build -o /dev/null github.com/fairride/identity/...` — ✅ clean
- No new dependencies: stdlib `strings`/`time` + `github.com/fairride/shared/errors` only

## Phase 2.4 — User Persistence (COMPLETE)

PostgreSQL repository implementation for User entity.

| File | Purpose |
|------|---------|
| `services/identity/infrastructure/postgres/user_repository.go` | `UserRepository`: FindByID, FindByPhone, FindAll, Save (upsert), Delete |
| `services/identity/infrastructure/postgres/user_repository_test.go` | 14 integration tests covering Save/Create, Save/Update (status transitions), DuplicatePhone, FindByID, FindByPhone, FindAll, Delete |
| `services/identity/infrastructure/postgres/testmain_test.go` | Updated: added `identity_users` table to createSchema/dropSchema/setupTest |

**Schema (test scaffolding — NOT a migration):**
- `identity_users` (id PK, phone_number TEXT UNIQUE NOT NULL, name TEXT, email TEXT DEFAULT '', type TEXT, status TEXT, role_id TEXT, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
- `email` stored as empty string when not provided (consistent with entity `string` type — no pointer/NULL needed)

**Key implementation details:**
- `Save` is an upsert (ON CONFLICT id): creates or updates in one call
- `created_at` is NOT in the UPDATE SET — immutable after first insert
- `updated_at` is taken from `user.UpdatedAt` (set by domain entity during status transitions), not from repository-side `time.Now()`
- Phone uniqueness enforced at DB level; duplicate returns `CodeAlreadyExists`
- `FindByPhone` lookup is exact-match (no normalisation — callers own that concern)

**Verification:**
- `go build -o /dev/null github.com/fairride/identity/...` — ✅ clean
- `go vet github.com/fairride/identity/...` — ✅ clean
- `go test github.com/fairride/identity/infrastructure/postgres/...` (no DATABASE_URL) — ✅ skips gracefully
- `go test -race github.com/fairride/identity/domain/entity/...` — ✅ 52 tests pass

## Phase 2.5 — JWT Foundation (COMPLETE)

HS256 token infrastructure using only stdlib. Zero new dependencies.

| File | Purpose |
|------|---------|
| `services/identity/infrastructure/jwt/config.go` | `Config` struct (AccessSecret, RefreshSecret, AccessTokenTTL, RefreshTokenTTL), `DefaultConfig()` (15 min / 7 days), `Validate()` — enforces ≥32-byte secrets, distinct secrets, positive TTLs, refresh > access |
| `services/identity/infrastructure/jwt/service.go` | `TokenService` (HS256 sign/verify), `AccessClaims`, `RefreshClaims`, `RefreshToken` value object; `GenerateAccessToken`, `GenerateRefreshToken`, `ValidateAccessToken`, `ValidateRefreshToken` |
| `services/identity/infrastructure/jwt/service_test.go` | 25 unit tests — Config validation, generate/validate round-trips, claim field correctness, expiry, wrong secret, malformed input, cross-token kind rejection |

**Key implementation decisions:**
- Pure stdlib: `crypto/hmac`, `crypto/sha256`, `crypto/rand`, `encoding/base64`, `encoding/json` — zero external JWT dependency
- Separate HMAC secrets for access vs refresh (distinct secret = leaked access key cannot forge refresh tokens)
- JTI: 16-byte `crypto/rand` → hex (32-char, one per token call — ensures uniqueness)
- Refresh token carries `Family` ID for future token rotation (generated now, enforced in Phase 2.6+)
- `ValidateX` returns `CodeUnauthenticated` for all failure modes — no distinction between expired and tampered (information leak prevention)
- `RefreshToken` value object carries `TokenID`, `Family`, `ExpiresAt` for server-side storage by the application layer
- `encodeToken`/`verifyToken` are package-private — not part of public API

**Payload field names (compact, to keep token size small):**
- `sub` userID, `exp` expiry, `iat` issued-at, `jti` token ID, `tkt` kind (access/refresh), `utype` user type, `rid` role ID, `fam` family

**Verification:**
- `go test -race github.com/fairride/identity/infrastructure/jwt/...` — ✅ 25/25 pass
- `go build -o /dev/null github.com/fairride/identity/...` — ✅ clean
- `go vet github.com/fairride/identity/...` — ✅ clean

## Phase 3 — User Profile Module (COMPLETE)

Standalone `services/user` microservice. gRPC transport, PostgreSQL persistence, Clean Architecture.

| File | Purpose |
|------|---------|
| `proto/user/v1/user_profile.proto` | Service definition: `UserProfileService` with `GetProfile` + `UpdateProfile` RPCs; `UserProfileProto` message with all 10 fields |
| `services/user/grpc/userpb/` | Generated pb.go + grpc.pb.go (protoc 35.1 + protoc-gen-go v1.34.2 + protoc-gen-go-grpc v1.3.0) |
| `services/user/domain/entity/profile.go` | `UserProfile` entity; `Gender` enum (male/female/other/unspecified); `ProfileStatus` enum (active/suspended/deleted); `NewUserProfile` (validated); `ReconstituteUserProfile` (no validation); `Update()` method |
| `services/user/domain/entity/profile_test.go` | 22 unit tests — construction, all field validations, date-of-birth rules, update, phone/status immutability |
| `services/user/domain/repository/profile_repository.go` | `ProfileRepository` interface (FindByID, Save) |
| `services/user/app/get_profile.go` | `GetProfileUseCase.Execute(ctx, userID)` |
| `services/user/app/update_profile.go` | `UpdateProfileUseCase.Execute(ctx, UpdateProfileInput)` — fetch → domain.Update → Save |
| `services/user/app/app_test.go` | 11 use-case unit tests with in-memory stub repository |
| `services/user/infrastructure/postgres/profile_repository.go` | `ProfileRepository`: FindByID, Save (upsert — created_at immutable, date_of_birth nullable) |
| `services/user/infrastructure/postgres/testmain_test.go` | TestMain: skip without DATABASE_URL; createSchema/dropSchema/setupTest |
| `services/user/infrastructure/postgres/profile_repository_test.go` | 7 integration tests (skipped without DATABASE_URL) |
| `services/user/grpc/handler.go` | `Handler` implements `UserProfileServiceServer`; `toProto()` + `toGRPCError()` helpers |
| `services/user/grpc/handler_test.go` | 10 gRPC handler unit tests with stub repo; covers success, empty ID, NotFound, validation error, error code mapping |
| `services/user/cmd/server/main.go` | Wires pool → ProfileRepository → UseCases → Handler → gRPC registration |

**Domain business rules:**
- `full_name`, `phone`: required (whitespace-only rejected)
- `email`: optional; if non-empty must pass `local@domain.tld` structural check
- `avatar`: optional; any non-empty string accepted (URL validation is caller's concern)
- `date_of_birth`: optional (zero = not provided); if set must be in the past, ≤150 years ago
- `gender`: one of male/female/other/unspecified
- `Update()` does NOT change `phone` or `status` — those are owned by Identity/admin flows
- `ProfileStatus` starts as `active` on creation

**DB schema (test scaffolding):**
- `user_profiles` (id PK, full_name, phone, email DEFAULT '', avatar DEFAULT '', date_of_birth TIMESTAMPTZ NULL, gender DEFAULT 'unspecified', status DEFAULT 'active', created_at, updated_at)
- `date_of_birth` is NULL in DB when not set (zero `time.Time` in entity)

**Proto toolchain (installed this phase):**
- `protoc` 35.1 via brew
- `protoc-gen-go` v1.34.2 and `protoc-gen-go-grpc` v1.3.0 via `go install`
- Generated code committed to `services/user/grpc/userpb/` (not gitignored)

**go.mod changes (user service):**
- Added: `google.golang.org/grpc v1.64.0`, `google.golang.org/protobuf v1.34.2`, `github.com/jackc/pgx/v5 v5.6.0`
- Added `replace github.com/fairride/shared => ../../shared` (for GOWORK=off tidy)

**Verification:**
- `go test -race github.com/fairride/user/domain/entity/... github.com/fairride/user/app/... github.com/fairride/user/grpc/...` — ✅ 43/43 pass
- `go test github.com/fairride/user/infrastructure/postgres/...` (no DATABASE_URL) — ✅ skips gracefully
- `go build -o /dev/null github.com/fairride/user/...` — ✅ clean
- `go vet github.com/fairride/user/...` — ✅ clean
- `go build -o /dev/null github.com/fairride/identity/...` — ✅ still clean

## Phase 4 — Driver Profile Module (COMPLETE)

Standalone `services/driver` microservice. gRPC transport, PostgreSQL persistence, Clean Architecture.

| File | Purpose |
|------|---------|
| `proto/driver/v1/driver_profile.proto` | Service definition: `DriverProfileService` with 5 RPCs; `DriverProfileProto` with all 12 fields |
| `services/driver/grpc/driverpb/` | Generated pb.go + grpc.pb.go (protoc 35.1) |
| `services/driver/domain/entity/driver.go` | `DriverProfile` entity; `VehicleType` enum (car/motorcycle/van); `OnlineStatus` enum (offline/online); `VerificationStatus` enum (pending/verified/rejected/suspended); `NewDriverProfile` (validated); `ReconstituteDriverProfile`; `Update`, `GoOnline`, `GoOffline`, `Verify`, `Reject`, `Suspend`, `Reinstate` |
| `services/driver/domain/entity/driver_test.go` | 30 unit tests — construction, all state transitions, error cases |
| `services/driver/domain/repository/driver_repository.go` | `DriverRepository` interface (FindByID, FindByUserID, Save) |
| `services/driver/app/get_driver.go` | `GetDriverProfileUseCase`, `GetDriverProfileByUserIDUseCase` |
| `services/driver/app/update_driver.go` | `UpdateDriverProfileUseCase`, `UpdateOnlineStatusUseCase`, `UpdateVerificationStatusUseCase`; `VerificationAction` type with 4 named constants |
| `services/driver/app/app_test.go` | 16 use-case unit tests with in-memory stub |
| `services/driver/infrastructure/postgres/driver_repository.go` | `DriverRepository`: FindByID, FindByUserID, Save (upsert — user_id immutable after first insert) |
| `services/driver/infrastructure/postgres/testmain_test.go` | TestMain: skip without DATABASE_URL; createSchema/dropSchema/setupTest |
| `services/driver/infrastructure/postgres/driver_repository_test.go` | 7 integration tests (skipped without DATABASE_URL) |
| `services/driver/grpc/handler.go` | `Handler` implements `DriverProfileServiceServer`; 5 methods + `toProto()` + `toGRPCError()` |
| `services/driver/grpc/handler_test.go` | 14 gRPC handler unit tests |
| `services/driver/cmd/server/main.go` | Wires pool → DriverRepository → 5 UseCases → Handler → gRPC registration |

**Domain business rules:**
- Required fields at creation: `driverID`, `userID`, `licenseNumber`, `plateNumber`, `vehicleType`
- Optional: `vehicleBrand`, `vehicleModel`, `vehicleColor` (empty string allowed)
- New drivers always start as `OnlineStatusOffline` + `VerificationStatusPending`
- `GoOnline`: requires `VerificationStatusVerified`; fails CodePreconditionFailed otherwise
- `GoOffline`: fails if already offline
- `Verify`: pending → verified only
- `Reject`: pending → rejected only
- `Suspend`: verified → suspended; forces offline as side-effect
- `Reinstate`: suspended → verified only
- `Update()`: validates licenseNumber, vehicleType, plateNumber; brand/model/color optional

**VerificationAction string constants** (used in proto `verification_status` field):
- `"verified"`, `"rejected"`, `"suspended"`, `"reinstated"`

**DB schema (test scaffolding):**
- `driver_profiles` (driver_id PK, user_id TEXT UNIQUE NOT NULL, license_number, vehicle_type, vehicle_brand DEFAULT '', vehicle_model DEFAULT '', vehicle_color DEFAULT '', plate_number, online_status DEFAULT 'offline', verification_status DEFAULT 'pending', created_at, updated_at TIMESTAMPTZ)
- `user_id` is NOT in the ON CONFLICT UPDATE SET — identity-linked, immutable

**go.mod (driver service):** `pgx/v5 v5.6.0`, `grpc v1.64.0`, `protobuf v1.34.2`, `replace shared => ../../shared`

**Verification:**
- `go test github.com/fairride/driver/domain/entity/... github.com/fairride/driver/app/... github.com/fairride/driver/grpc/...` — ✅ 60/60 pass
- `go test github.com/fairride/driver/infrastructure/postgres/...` (no DATABASE_URL) — ✅ skips gracefully
- `go build -o /dev/null github.com/fairride/driver/...` — ✅ clean

## Phase 5 — Vehicle Module (COMPLETE)

Added to `services/driver` (driver bounded context). Reuses `VehicleType` enum from Phase 4.

| File | Purpose |
|------|---------|
| `proto/driver/v1/vehicle.proto` | `VehicleService` with 4 RPCs (Create/Update/Delete/List); `VehicleProto` with 10 fields |
| `services/driver/grpc/driverpb/vehicle.pb.go` + `vehicle_grpc.pb.go` | Generated (same driverpb package) |
| `services/driver/domain/entity/vehicle.go` | `Vehicle` entity; reuses `VehicleType` enum; `NewVehicle` (validated); `ReconstituteVehicle`; `Update`; `year` field (0=not provided) |
| `services/driver/domain/entity/vehicle_test.go` | 16 entity unit tests — construction, year bounds, all vehicle types, update, reconstitution |
| `services/driver/domain/repository/vehicle_repository.go` | `VehicleRepository` interface (FindByID, FindByDriverID, Save, Delete) |
| `services/driver/app/vehicle.go` | 4 use cases: `CreateVehicleUseCase` (generates ID), `UpdateVehicleUseCase`, `DeleteVehicleUseCase`, `ListVehiclesUseCase` |
| `services/driver/app/vehicle_test.go` | 13 use-case unit tests with in-memory stub |
| `services/driver/infrastructure/postgres/vehicle_repository.go` | PostgreSQL impl: FindByID, FindByDriverID, Save (upsert), Delete (hard delete with RowsAffected check) |
| `services/driver/infrastructure/postgres/testmain_test.go` | Updated: added `vehicles` table + index to createSchema/dropSchema; TRUNCATE both tables in setupTest |
| `services/driver/infrastructure/postgres/vehicle_repository_test.go` | 8 integration tests (skipped without DATABASE_URL) |
| `services/driver/grpc/vehicle_handler.go` | `VehicleHandler` implements `VehicleServiceServer`; 4 RPC methods + `vehicleToProto()` |
| `services/driver/grpc/vehicle_handler_test.go` | 14 gRPC handler unit tests |
| `services/driver/cmd/server/main.go` | Updated: adds VehicleRepository → 4 UseCases → VehicleHandler → `RegisterVehicleServiceServer` |

**Domain business rules:**
- Required at creation: `vehicleID`, `driverID`, `vehicleType`, `plateNumber`
- Optional: `brand`, `model`, `color`, `year` (0 = not provided)
- Year constraints: if > 0, must be ≥ 1900 and ≤ current_year + 1 (next model year allowed)
- `Update()` does NOT change `vehicleID`, `driverID`, or `createdAt`
- `Delete()` is a hard delete; returns CodeNotFound if the vehicle does not exist
- `ListVehicles` returns empty slice (not error) when driver has no vehicles

**ID generation:** `CreateVehicleUseCase` generates a 32-char hex random ID via `crypto/rand` (16 bytes).

**DB schema (test scaffolding):**
- `vehicles` (vehicle_id PK, driver_id NOT NULL, type, brand DEFAULT '', model DEFAULT '', color DEFAULT '', plate_number, year INT DEFAULT 0, created_at, updated_at TIMESTAMPTZ)
- Index: `vehicles_driver_id_idx` on `(driver_id)` for `FindByDriverID`
- `driver_id` is NOT in the ON CONFLICT UPDATE SET — the vehicle's owning driver is immutable

**Verification:**
- `go test github.com/fairride/driver/domain/entity/... github.com/fairride/driver/app/... github.com/fairride/driver/grpc/...` — ✅ 103/103 pass
- `go test github.com/fairride/driver/infrastructure/postgres/...` (no DATABASE_URL) — ✅ skips gracefully
- `go build -o /dev/null github.com/fairride/driver/...` — ✅ clean

## Phase 6 — Driver Availability Module (COMPLETE)

Added to `services/driver`. Pure Redis layer — no PostgreSQL, no GPS, no WebSocket.

| File | Purpose |
|------|---------|
| `proto/driver/v1/availability.proto` | `DriverAvailabilityService` (GoOnline, GoOffline, Heartbeat, GetAvailability); `AvailabilityResponse` |
| `services/driver/grpc/driverpb/availability.pb.go` + `_grpc.pb.go` | Generated (same driverpb package) |
| `services/driver/domain/entity/availability.go` | `AvailabilityState` value object: `DriverID`, `IsOnline`, `LastSeen` |
| `services/driver/domain/repository/availability_repository.go` | `AvailabilityRepository` interface (SetOnline, SetOffline, RefreshHeartbeat, GetAvailability) |
| `services/driver/app/availability.go` | 4 use cases: `GoOnlineUseCase`, `GoOfflineUseCase`, `HeartbeatUseCase`, `GetAvailabilityUseCase` |
| `services/driver/app/availability_test.go` | 14 use-case unit tests + full lifecycle test |
| `services/driver/infrastructure/redis/availability_repository.go` | Redis impl with pipeline; `NewAvailabilityRepositoryWithTTL` for test TTL injection |
| `services/driver/infrastructure/redis/testmain_test.go` | Skip if REDIS_ADDR not set |
| `services/driver/infrastructure/redis/availability_repository_test.go` | 9 integration tests incl. TTL expiry |
| `services/driver/grpc/availability_handler.go` | `AvailabilityHandler` implements `DriverAvailabilityServiceServer`; `availabilityToProto` |
| `services/driver/grpc/availability_handler_test.go` | 12 gRPC handler unit tests |
| `services/driver/cmd/server/main.go` | Updated: Redis connect → AvailabilityRepository → 4 UseCases → AvailabilityHandler → register |

**Redis key schema:**
- `fairride:drv:online:{driverID}` — TTL=5min; presence = driver is online
- `fairride:drv:lastseen:{driverID}` — TTL=30d; value = RFC3339Nano timestamp

**Behaviour contracts:**
- `SetOnline`: idempotent — resets TTL if already online
- `SetOffline`: idempotent — DEL online key, update last_seen
- `RefreshHeartbeat`: uses `EXPIRE`; returns `CodePreconditionFailed` if key not present (TTL expired or never set)
- `GetAvailability`: never returns CodeNotFound; zero `LastSeen` = never seen
- `last_seen` survives the online key's TTL (independent key with longer TTL)

**Architecture note:**
Two separate "online" concepts in the driver service:
- PostgreSQL `driver_profiles.online_status` (Phase 4) — persistent authorization state set by admin/domain logic; requires VerificationStatus=Verified
- Redis `fairride:drv:online:*` (Phase 6) — real-time heartbeat presence; set by driver app. These are independent — a driver must be authorized at DB level AND heartbeating at Redis level to receive trips.

**go.mod changes (driver service):** Added `github.com/redis/go-redis/v9 v9.5.1` + transitive deps `cespare/xxhash/v2`, `dgryski/go-rendezvous`

**Verification:**
- `go test github.com/fairride/driver/domain/entity/... github.com/fairride/driver/app/... github.com/fairride/driver/grpc/...` — ✅ 130/130 pass
- `go test github.com/fairride/driver/infrastructure/redis/...` (no REDIS_ADDR) — ✅ skips gracefully
- `go build -o /dev/null github.com/fairride/driver/...` — ✅ clean

## Phase 7 — Trip Foundation (COMPLETE)

Standalone `services/trip` microservice. gRPC transport, PostgreSQL persistence, Clean Architecture.

| File | Purpose |
|------|---------|
| `proto/trip/v1/trip.proto` | `TripService` with 3 RPCs (CreateTrip, CancelTrip, GetTrip); `TripProto` with 9 fields |
| `services/trip/grpc/trippb/trip.pb.go` + `trip_grpc.pb.go` | Generated (protoc 35.1) |
| `services/trip/domain/entity/trip.go` | `TripStatus` enum (7 values); `Trip` aggregate root; `NewTrip` (validated); `ReconstituteTrip`; `Cancel(reason, now)` — PreconditionFailed from InProgress/Completed/Cancelled; `IsCancellable()` |
| `services/trip/domain/entity/trip_test.go` | 13 unit tests — construction, all Cancel status paths, reconstitution |
| `services/trip/domain/repository/trip_repository.go` | `TripRepository` interface (Save, FindByID, FindByRiderID) |
| `services/trip/app/create_trip.go` | `CreateTripUseCase` — generates tripID via `crypto/rand`, calls NewTrip, repo.Save |
| `services/trip/app/cancel_trip.go` | `CancelTripUseCase` — FindByID → Cancel → Save |
| `services/trip/app/get_trip.go` | `GetTripUseCase` — repo.FindByID |
| `services/trip/app/app_test.go` | 8 use-case tests with in-memory stub repo |
| `services/trip/infrastructure/postgres/trip_repository.go` | PostgreSQL impl: Save (upsert), FindByID, FindByRiderID (ORDER BY created_at DESC) |
| `services/trip/infrastructure/postgres/testmain_test.go` | TestMain: skip without DATABASE_URL; trips table schema |
| `services/trip/infrastructure/postgres/trip_repository_test.go` | 6 integration tests (skipped without DATABASE_URL) |
| `services/trip/grpc/handler.go` | `Handler` implements `TripServiceServer`; 3 RPC methods + `toProto()` + `toGRPCError()` |
| `services/trip/grpc/handler_test.go` | 10 gRPC handler unit tests |
| `services/trip/cmd/server/main.go` | Wires pool → TripRepository → 3 UseCases → Handler → RegisterTripServiceServer |

**Trip status enum (7 values):**
- `pending`, `searching`, `driver_assigned`, `driver_arrived`, `in_progress`, `completed`, `cancelled`

**Cancellable statuses:** pending, searching, driver_assigned, driver_arrived (NOT in_progress, completed, or already cancelled)

**DB schema (test scaffolding):**
- `trips` (trip_id PK, rider_id NOT NULL, driver_id DEFAULT '', status DEFAULT 'pending', pickup_address NOT NULL, dropoff_address NOT NULL, cancellation_reason DEFAULT '', created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
- Index: `trips_rider_id_idx` on `(rider_id)` for FindByRiderID
- On upsert conflict: updates driver_id, status, cancellation_reason, updated_at; rider_id/pickup_address/dropoff_address are immutable after insert

**go.mod (trip service):** `pgx/v5 v5.6.0`, `grpc v1.64.0`, `protobuf v1.34.2`, `replace shared => ../../shared`

**Verification:**
- `go test github.com/fairride/trip/domain/entity/... github.com/fairride/trip/app/... github.com/fairride/trip/grpc/...` — ✅ 31/31 pass
- `go test github.com/fairride/trip/infrastructure/postgres/...` (no DATABASE_URL) — ✅ skips gracefully
- `go build -o /dev/null github.com/fairride/trip/...` — ✅ clean

## Phase 8 — Dispatch MVP (COMPLETE)

Standalone `services/dispatch` microservice. gRPC, Redis GEO, PostgreSQL. Clean Architecture.
Algorithm: nearest available driver wins. No AI, scoring, surge, or heat maps.

| File | Purpose |
|------|---------|
| `proto/dispatch/v1/dispatch.proto` | 5 RPCs: RequestDispatch, AcceptTrip, RejectTrip, UpdateDriverLocation, GetDispatchStatus |
| `services/dispatch/grpc/dispatchpb/` | Generated pb.go + grpc.pb.go |
| `services/dispatch/domain/entity/dispatch_job.go` | `JobStatus` enum (5 values); `DispatchJob` aggregate; `NearbyDriver` value; `OfferToDriver`, `Accept`, `Reject`, `TimeoutOffer`, `MarkFailed`, `Cancel`; `HasBeenOffered`, `IsOfferExpired`, `OfferedDriverIDsCSV` |
| `services/dispatch/domain/entity/dispatch_job_test.go` | 24 entity unit tests |
| `services/dispatch/domain/repository/repository.go` | 3 interfaces: `DispatchJobRepository`, `DriverLocationRepository`, `TripUpdater` |
| `services/dispatch/app/offer_next_driver.go` | Shared `offerNextDriver()` helper — finds next eligible driver or fails the job |
| `services/dispatch/app/request_dispatch.go` | `RequestDispatchUseCase` — creates job, sets trip→searching, offers nearest driver |
| `services/dispatch/app/accept_trip.go` | `AcceptTripUseCase` — validates offer, sets trip→driver_assigned |
| `services/dispatch/app/reject_trip.go` | `RejectTripUseCase` — clears offer, retries with next nearest |
| `services/dispatch/app/update_location.go` | `UpdateDriverLocationUseCase` |
| `services/dispatch/app/get_dispatch_status.go` | `GetDispatchStatusUseCase` |
| `services/dispatch/app/engine.go` | `DispatchEngine` — background goroutine (5s tick) that auto-retries expired offers |
| `services/dispatch/app/app_test.go` | 16 use-case unit tests with in-memory stubs |
| `services/dispatch/infrastructure/redis/driver_location_repository.go` | Redis GEO: `UpdateLocation` (GEOADD + TTL key), `FindNearby` (GEOSEARCH ASC), `IsActive` (EXISTS), `RemoveLocation` (ZREM + DEL) |
| `services/dispatch/infrastructure/redis/testmain_test.go` | Skip if REDIS_ADDR unset |
| `services/dispatch/infrastructure/redis/driver_location_repository_test.go` | 4 Redis integration tests |
| `services/dispatch/infrastructure/postgres/dispatch_repository.go` | PostgreSQL impl: Save (upsert), FindByID, FindByTripID, FindExpiredOffers |
| `services/dispatch/infrastructure/postgres/trip_updater.go` | `TripUpdater`: SetSearching, AssignDriver — updates shared `trips` table |
| `services/dispatch/infrastructure/postgres/testmain_test.go` | Creates both `trips` + `dispatch_jobs` tables |
| `services/dispatch/infrastructure/postgres/dispatch_repository_test.go` | 7 Postgres integration tests |
| `services/dispatch/grpc/handler.go` | `Handler` embeds `UnimplementedDispatchServiceServer`; all 5 RPCs + `toProto` + `toGRPCError` |
| `services/dispatch/grpc/handler_test.go` | 16 gRPC handler unit tests |
| `services/dispatch/cmd/server/main.go` | Wires pool → Redis → 5 UseCases → Handler → register; starts engine when both DB+Redis ready |

**Dispatch job status (5 values):** `pending → searching → assigned`, or `failed` / `cancelled`

**Dispatch algorithm:**
1. `RequestDispatch` creates job, sets trip to `searching`, calls `offerNextDriver`
2. `offerNextDriver`: `GEOSEARCH` (nearest first) → filter `HasBeenOffered` + `IsActive` → `OfferToDriver` → save; if none found or max attempts reached → `MarkFailed`
3. `AcceptTrip`: validates job searching, driverID matches current offer, offer not expired → `Assigned`; updates trip to `driver_assigned`
4. `RejectTrip`: validates driver match → clears offer → `offerNextDriver` retries
5. `DispatchEngine`: polls DB every 5s for `status='searching' AND offer_expires_at < NOW()` → `TimeoutOffer` → `offerNextDriver`

**Redis key schema (dispatch service — independent of Phase 6 availability):**
- `fairride:dispatch:drv:loc` — GEO sorted set; GEOADD/GEOSEARCH/ZREM
- `fairride:dispatch:drv:active:{driverID}` — TTL=35s; SET on each location update; EXISTS to check active

**PostgreSQL schema (dispatch_jobs):**
- `job_id` PK, `trip_id` UNIQUE NOT NULL, `rider_id`, `pickup_lat/lon` DOUBLE PRECISION, `status`, `current_driver_id` DEFAULT '', `assigned_driver_id` DEFAULT '', `offered_driver_ids` TEXT DEFAULT '' (comma-separated), `offer_expires_at` TIMESTAMPTZ (NULL when no active offer), `offer_timeout_sec` INT DEFAULT 30, `max_attempts` INT DEFAULT 5, `attempt_count` INT DEFAULT 0, `created_at`, `updated_at`
- Partial index on `(offer_expires_at) WHERE status='searching'` for efficient expired-offer queries
- Dispatch service also directly updates `trips.status` and `trips.driver_id` (shared DB, MVP trade-off)

**go.mod (dispatch service):** `pgx/v5 v5.6.0`, `go-redis/v9 v9.5.1`, `grpc v1.64.0`, `protobuf v1.34.2`, `replace shared => ../../shared`

**Verification:**
- `go test github.com/fairride/dispatch/domain/entity/... github.com/fairride/dispatch/app/... github.com/fairride/dispatch/grpc/...` — ✅ 56/56 pass
- `go test github.com/fairride/dispatch/infrastructure/postgres/...` (no DATABASE_URL) — ✅ skips gracefully
- `go test github.com/fairride/dispatch/infrastructure/redis/...` (no REDIS_ADDR) — ✅ skips gracefully
- `go build -o /dev/null github.com/fairride/dispatch/...` — ✅ clean

## Phase 9 — Pricing MVP (COMPLETE)

Pure compute `services/pricing` microservice. No DB, no Redis. Clean Architecture.
Formula: `ride_fare = max(base + distance + time, MinimumFare)`, `total = ride_fare + BookingFee`. All amounts int64 in smallest currency unit.

| File | Purpose |
|------|---------|
| `proto/pricing/v1/pricing.proto` | 2 RPCs: EstimateFare, CalculateFinalFare; FareBreakdown message (11 fields) |
| `services/pricing/grpc/pricingpb/` | Generated pb.go + grpc.pb.go |
| `services/pricing/domain/entity/fare.go` | `VehicleType` enum (car/motorcycle/van); `VehicleRates` struct; `FareConfig` + `DefaultFareConfig()`; `FareBreakdown` struct |
| `services/pricing/domain/entity/fare_test.go` | 4 entity unit tests — config completeness, positive rates, minimum≥base, constants |
| `services/pricing/app/fare_calculator.go` | `FareCalculator`: `Estimate` (IsFinal=false), `CalculateFinal` (IsFinal=true), shared `calculate`; `roundToUnit` with math.Round |
| `services/pricing/app/fare_calculator_test.go` | 20 unit tests — all 3 vehicle types, minimum fare enforcement, distance/time rounding, booking fee invariant, IsFinal flag, upfront pricing guarantee, all error cases, zero-zero inputs |
| `services/pricing/grpc/handler.go` | `Handler` embeds `UnimplementedPricingServiceServer`; 2 RPCs + `toProto` + `toGRPCError` |
| `services/pricing/grpc/handler_test.go` | 9 gRPC handler unit tests — valid car/motorcycle/van, missing vehicle type, unknown vehicle type, negative distance, minimum fare, IsFinal flag, proto breakdown invariant |
| `services/pricing/cmd/server/main.go` | Wires `NewFareCalculator(DefaultFareConfig())` → Handler → `RegisterPricingServiceServer`; no DB/Redis needed |

**DefaultFareConfig (USD cents):**
- Car: BaseFare=50, PerKmRate=30, PerMinuteRate=5, MinimumFare=200, BookingFee=50
- Motorcycle: BaseFare=30, PerKmRate=20, PerMinuteRate=3, MinimumFare=150, BookingFee=30
- Van: BaseFare=100, PerKmRate=50, PerMinuteRate=8, MinimumFare=300, BookingFee=75

**VehicleType duplicated from driver service** — no cross-service import dependency.
**Upfront pricing guarantee:** Estimate and CalculateFinal use identical formula; only `IsFinal` flag differs.

**go.mod (pricing service):** `grpc v1.64.0`, `protobuf v1.34.2`, `replace shared => ../../shared`. No pgx, no go-redis.

**Verification:**
- `go test github.com/fairride/pricing/...` — ✅ 33/33 pass (4 entity + 20 app + 9 grpc)
- `go build -o /dev/null github.com/fairride/pricing/...` — ✅ clean

## Phase 10 — Booking Flow MVP (COMPLETE)

End-to-end booking orchestration connecting Trip, Dispatch, and Pricing services.

### Trip service extensions (services/trip)
- `Trip.Start(now)` — DriverAssigned/DriverArrived → InProgress (CodePreconditionFailed otherwise)
- `Trip.Complete(finalFareTotal, fareCurrency, now)` — InProgress → Completed, stores fare
- `Trip.FinalFareTotal int64`, `Trip.FareCurrency string` — new fields persisted in DB
- `ReconstituteTrip` signature updated: added `finalFareTotal int64, fareCurrency string` params
- Proto: added `StartTrip`, `CompleteTrip` RPCs; added `final_fare_total`, `fare_currency` to `TripProto`
- New use cases: `StartTripUseCase`, `CompleteTripUseCase`
- DB schema: added `final_fare_total BIGINT DEFAULT 0`, `fare_currency TEXT DEFAULT ''` columns
- All Save/Select queries updated to include new columns
- Tests: +18 new tests (6 entity + 8 app + 10 handler) → trip total: 57 tests

### Booking service (services/booking) — NEW
Pure orchestration layer. No DB, no Redis. Composes Trip + Dispatch + Pricing via gRPC.

| File | Purpose |
|------|---------|
| `proto/booking/v1/booking.proto` | 6 RPCs: BookRide, AcceptDispatchOffer, RejectDispatchOffer, StartTrip, FinishTrip, GetBookingDetails |
| `grpc/bookingpb/` | Generated proto files |
| `app/clients.go` | `TripClient`, `DispatchClient`, `PricingClient` interfaces + `TripInfo`, `DispatchInfo`, `FareInfo` DTOs |
| `app/book_ride.go` | `BookRideUseCase`: CreateTrip → RequestDispatch |
| `app/accept_reject.go` | `AcceptDispatchOfferUseCase`, `RejectDispatchOfferUseCase` |
| `app/start_trip.go` | `StartTripUseCase`: delegates to TripClient.StartTrip |
| `app/finish_trip.go` | `FinishTripUseCase`: CalculateFinalFare → CompleteTrip |
| `app/get_booking.go` | `GetBookingDetailsUseCase`: GetTrip + GetDispatchStatus (dispatch 404 → "unknown") |
| `app/app_test.go` | 16 use case unit tests + `TestFullBookingFlow` (all 5 steps in sequence) |
| `grpc/handler.go` | `Handler` embeds `UnimplementedBookingServiceServer`; all 6 RPCs |
| `grpc/handler_test.go` | 14 gRPC handler unit tests |
| `grpc/adapters/trip_adapter.go` | `TripAdapter` wrapping `trippb.TripServiceClient` |
| `grpc/adapters/dispatch_adapter.go` | `DispatchAdapter` wrapping `dispatchpb.DispatchServiceClient` |
| `grpc/adapters/pricing_adapter.go` | `PricingAdapter` wrapping `pricingpb.PricingServiceClient` |
| `cmd/server/main.go` | Wires gRPC client connections → adapters → use cases → handler → register |

**Complete booking flow implemented:**
1. `BookRide` → trip.CreateTrip + dispatch.RequestDispatch → status: searching
2. `AcceptDispatchOffer` → dispatch.AcceptTrip → status: driver_assigned
3. `RejectDispatchOffer` → dispatch.RejectTrip → dispatch retries next driver
4. `StartTrip` → trip.StartTrip → status: in_progress
5. `FinishTrip` → pricing.CalculateFinalFare + trip.CompleteTrip → status: completed
6. `GetBookingDetails` → trip.GetTrip + dispatch.GetDispatchStatus (graceful on dispatch 404)

**NOT implemented:** Payment, Wallet, Notifications, GPS navigation, Chat

**go.mod (booking service):** depends on dispatch, pricing, shared, trip + grpc + protobuf; no pgx, no go-redis
**go.work:** updated to include `./services/booking`

**Verification:**
- `go test github.com/fairride/trip/domain/entity/... .../app/... .../grpc/...` — ✅ 57 pass (up from 31)
- `go test github.com/fairride/booking/...` — ✅ 30 pass (16 app + 14 grpc)
- `go build -o /dev/null github.com/fairride/{trip,dispatch,pricing,booking}/...` — ✅ all clean
- Total across all phases: 173 unit tests pass

## Phase 11 — API Gateway MVP (COMPLETE)

HTTP-only gateway service exposing REST endpoints to Rider App and Driver App. Calls the Booking service via gRPC internally.

### Architecture
- **Pure HTTP** — no gRPC server. Uses `net/http` ServeMux (Go 1.22 with method+path routing).
- **JWT auth middleware** — validates Bearer tokens via `identity/infrastructure/jwt.TokenService`. RiderID/DriverID extracted from claims — never from the request body.
- **gRPC client** — single `bookingpb.BookingServiceClient` connection to the Booking service.
- **Error mapping** — gRPC status codes → HTTP status codes at the edge.
- **Custom HTTP server** — does NOT use `shared/server/bootstrap.go` (which also starts gRPC).

### Files

| File | Purpose |
|------|---------|
| `services/gateway/go.mod` | Depends on booking, identity, shared; replace directives for all local modules (incl. transitive: dispatch, pricing, trip) |
| `services/gateway/http/middleware/auth.go` | `Auth(svc)` middleware — extracts Bearer, validates via `jwt.TokenService.ValidateAccessToken`, injects `*AccessClaims` into context via `ClaimsKey`; `ClaimsFromContext` helper |
| `services/gateway/http/middleware/logging.go` | `Logging(log)` middleware — wraps ResponseWriter to capture status code; logs method/path/status/duration via zerolog |
| `services/gateway/http/handlers/errors.go` | `writeGRPCError`, `writeBadRequest`, `writeJSON`; `grpcToHTTP` mapping (NotFound→404, InvalidArgument→400, Unauthenticated→401, PermissionDenied→403, FailedPrecondition→422, AlreadyExists→409, else→500) |
| `services/gateway/http/handlers/booking_handler.go` | `BookingClient` interface (mockable); `BookingHandler` with 6 HTTP handler methods |
| `services/gateway/http/router.go` | `NewRouter(bh, authMiddleware, log)` — builds ServeMux with `/health` (no auth) + all `/api/v1/rides/*` routes (auth-wrapped) + logging outer wrapper |
| `services/gateway/cmd/server/main.go` | Reads JWT secrets (JWT_ACCESS_SECRET, JWT_REFRESH_SECRET required), connects to booking service (BOOKING_ADDR), builds handler chain, starts `http.Server` |
| `services/gateway/http/middleware/auth_test.go` | 5 unit tests: no header, invalid format, invalid token, valid token+claims-in-context, case-insensitive Bearer |
| `services/gateway/http/handlers/booking_handler_test.go` | 14 unit tests: all 6 handlers (success + error paths) + gRPC error code mapping table (7 codes) |

### REST endpoints

| Method | Path | Handler | Auth source |
|--------|------|---------|------------|
| `POST` | `/api/v1/rides` | BookRide | RiderID from JWT `UserID` |
| `GET` | `/api/v1/rides/{tripID}` | GetBooking | — |
| `POST` | `/api/v1/rides/{tripID}/accept` | AcceptDispatchOffer | DriverID from JWT `UserID` |
| `POST` | `/api/v1/rides/{tripID}/reject` | RejectDispatchOffer | DriverID from JWT `UserID` |
| `POST` | `/api/v1/rides/{tripID}/start` | StartTrip | — |
| `POST` | `/api/v1/rides/{tripID}/finish` | FinishTrip | body: `{vehicle_type, distance_km, duration_min}` |
| `GET` | `/health` | inline | no auth |

### Key design decisions
- `BookingClient` interface in handlers package (not imported from bookingpb) — enables unit testing without gRPC.
- Auth middleware and router are decoupled — router accepts `func(http.Handler) http.Handler`; the JWT service is only imported in `cmd/server/main.go`.
- `ClaimsKey` is an unexported `contextKey` type — prevents accidental string collision with other context values.
- `statusRecorder` wrapper captures response status for logging after the handler writes it.
- Go 1.22 `r.PathValue("tripID")` used for path params — no external router.

**go.work:** updated to include `./services/gateway`

**Verification:**
- `go test github.com/fairride/gateway/...` — ✅ 18 pass (5 middleware + 13 handler)
- `go build -o /dev/null github.com/fairride/gateway/...` — ✅ clean
- `go build -o /dev/null github.com/fairride/{trip,dispatch,pricing,booking,gateway}/...` — ✅ all clean
- Total across all phases: 191 unit tests pass

## Phase 15 — Location Engine (COMPLETE — pub get + analyze pending)

Reusable, stream-based GPS service layer for the Rider App. No UI dependency.

### New files
```
apps/rider/lib/core/location/
  location_engine_config.dart   — LocationEngineConfig value class
  location_engine.dart          — LocationEngine class + public types
  location.dart                 — barrel export
```

### Public API

**Value types:**
- `LocationUpdate` — immutable GPS fix: `latitude`, `longitude`, `accuracyMeters`, `timestamp`, `altitude`, `speed`, `heading`
- `GpsStatus` enum — `enabled` / `disabled`
- `LocationPermissionStatus` enum — `granted` / `denied` / `permanentlyDenied`
- `LocationEngineState` enum — `stopped` / `running` / `paused`

**`LocationEngineConfig`:**
- `accuracy: LocationAccuracy` — default `high`
- `distanceFilter: double` — metres; default `5.0`
- `updateIntervalMs: int` — ms (Android only); default `5000`
- `copyWith()` for immutable updates

**`LocationEngine`:**
| Member | Description |
|---|---|
| `locationStream` | `Stream<LocationUpdate>` — broadcast; continuous GPS fixes |
| `gpsStatusStream` | `Stream<GpsStatus>` — GPS on/off events while running |
| `permissionStream` | `Stream<LocationPermissionStatus>` — permission changes while running |
| `state` | Current `LocationEngineState` |
| `config` | Current `LocationEngineConfig` |
| `start()` | Check permission → start GPS status listener → start position stream |
| `stop()` | Cancel all subscriptions → back to `stopped` |
| `pause()` | Suspend position event delivery (GPS status still flows) |
| `resume()` | Resume position events from `paused` state |
| `updateConfig(config)` | Apply new config; restarts position stream if running |
| `dispose()` | `stop()` + close all StreamControllers; engine unusable after this |

### Platform-specific LocationSettings
- Android: `AndroidSettings(accuracy, distanceFilter, intervalDuration)` — honours `updateIntervalMs`
- iOS/macOS: `AppleSettings(accuracy, distanceFilter, activityType: other, pauseLocationUpdatesAutomatically: false)`
- Fallback: `LocationSettings(accuracy, distanceFilter)` for non-mobile

### Behaviour on GPS/permission events while running
- **GPS disabled:** `_gpsStatusCtrl.add(GpsStatus.disabled)` → cancel position sub → keep `state=running` → when GPS re-enables: auto-restart position stream
- **Permission revoked:** `PermissionDeniedException` caught from position stream → `_permissionCtrl.add(denied)` → cancel position sub
- **GPS re-enabled:** GPS status stream fires → `_startPositionStream()` called automatically

### Action required on home machine
```bash
cd apps/rider
flutter pub get
flutter analyze
```

IDE shows phantom "package not found" errors until `pub get` runs — not real code errors.

**NOT implemented (by design):** booking, driver tracking, routes, Google Directions, geocoding, API calls, any UI.

## Phase 17 — Pickup & Destination Selection (COMPLETE — pub get + analyze pending)

Pickup and destination selection UI on top of the Phase 14 map foundation.
`map_page.dart` completely rewritten; no new dependencies.

### New file
`lib/features/map/domain/models/trip_selection.dart`
- `TripSelection` value class: `pickup LatLng`, `destination LatLng`, `pickupAddress String?`, `destinationAddress String?`
- `pickupAddress` / `destinationAddress` are null until geocoding is added (Phase 18+)

### Modified file
`lib/features/map/presentation/pages/map_page.dart`

**New state machine (`_SelectionMode`):**
| Mode | User action | Center pin | Panel content |
|---|---|---|---|
| `pickupPending` | Drag map to set pickup | ✅ visible | "Set pickup" row (live coords) + Confirm Pickup button |
| `destinationPending` | Drag map to set destination | ✅ visible | Pickup row (confirmed, Edit) + "Set destination" row (live coords) + Confirm Destination button |
| `confirmed` | Both set | ❌ hidden | Pickup row (Edit) + Destination row (Edit) |

**`_CenterPin` widget:**
- Absolute overlay centred in the Stack
- `padding: EdgeInsets.only(bottom: 48)` shifts the icon upward so the pin tip aligns with the geometric map centre
- Hidden when `_selectionMode == confirmed`

**`_SelectionPanel` widget (bottom sheet):**
- `Material(elevation: 12, borderRadius: vertical top 20)`
- `SafeArea(top: false)` handles home indicator
- `_PointRow` shows: icon + label + subtitle (optional) + coordinate (formatted to 5dp) + trailing widget (Edit button)

**Key interactions:**
- `GoogleMap.onCameraMove` → updates `_cameraCenter` live; skipped when `confirmed`
- `GoogleMap.padding: EdgeInsets.only(bottom: 240)` → Google Maps controls sit above panel
- Confirm Pickup: `_pickupPoint = _cameraCenter`; if destination already set → go to `confirmed` (edit-pickup flow), else → `destinationPending`
- Confirm Destination: `_destinationPoint = _cameraCenter` → `confirmed`
- Edit Pickup: clear `_pickupPoint`, keep `_destinationPoint`; animate camera back to last pickup; → `pickupPending`
- Edit Destination: clear `_destinationPoint`, keep `_pickupPoint`; animate camera back to last destination; → `destinationPending`
- Green marker = pickup (confirmed), Red marker = destination (confirmed)

**`_tripSelection` getter:**
Returns `TripSelection(pickup, destination)` when both are confirmed; `null` otherwise. Prepared for booking phase.

**NOT implemented (by design):** route calculation, fare estimation, Booking API, driver search, trip creation, geocoding, address lookup.

### Coordinate display
Formatted as `lat, lng` to 5 decimal places (≈ 1 m precision). Address field always null until geocoding phase.

### Action required on home machine
```bash
cd apps/rider
flutter pub get
flutter analyze
```

---

## Backlog — Map Abstraction Layer (not yet implemented)

**Priority:** Before Route Engine / Directions API phase.
**Motivation:** MapPage currently uses `google_maps_flutter` directly. Switching to Mapbox or HERE Maps later would require touching all map code. An abstraction layer isolates that risk.

**Proposed interface (future phase):**
```dart
abstract class MapProvider {
  Future<List<Place>> search(String keyword);
  Future<RouteResult> getRoute(LatLng origin, LatLng destination);
  Future<String> reverseGeocode(LatLng location);
  Future<LatLng> currentLocation();
}
```
`GoogleMapProvider` would implement this. All map-using pages consume `MapProvider` via DI.
**Implement before:** Route Engine, Directions API, Polyline, ETA, Distance phases.

---

## Human Checkpoints Pending
| HC ID | Scope | Action Required |
|-------|-------|----------------|
| HC-001 | DOC-0001 | CTO approval → rename to v1.0 |
| HC-002 | DOC-0001A | CTO approval → rename to v1.0 |
| HC-003 | DOC-0002 | CTO + CPO approval → rename to v1.0 |
| HC-P7 | Phase 7 Trip Foundation | CTO approval to proceed to next phase |
| HC-P8 | Phase 8 Dispatch MVP | CTO approval to proceed to next phase |
| HC-P9 | Phase 9 Pricing MVP | CTO approval to proceed to next phase |
| HC-P10 | Phase 10 Booking Flow MVP | CTO approval to proceed to next phase |
| HC-P11 | Phase 11 API Gateway MVP | CTO approval to proceed to next phase |
| HC-P12 | Phase 12 Rider App Skeleton | CTO approval to proceed to next phase |
| HC-P14 | Phase 14 Map Foundation | CTO approval to proceed to next phase |
| HC-P15 | Phase 15 Location Engine | CTO approval to proceed to next phase |
| HC-P17 | Phase 17 Pickup & Destination Selection | CTO approval to proceed to next phase |

## Phase 14 — Map Foundation (COMPLETE — pub get + analyze pending)

Google Maps integration for the Rider App. Home tab now shows a full-screen interactive map centred on the user's GPS location.

### New dependency (pubspec.yaml)
- `google_maps_flutter: ^2.10.0` — Google Maps Flutter SDK
- `geolocator: ^13.0.0` — GPS + location permission handling

### New file
`lib/features/map/presentation/pages/map_page.dart`

**State machine (`_LocationStatus` enum):**
| State | Trigger | UI |
|---|---|---|
| `loading` | initial / retry | `CircularProgressIndicator` + "Finding your location…" |
| `permissionDenied` | user tapped Deny | error view + "Grant permission" → retries `_resolveLocation` |
| `permissionPermanentlyDenied` | permanently denied | error view + "Open Settings" → `Geolocator.openAppSettings()` |
| `gpsDisabled` | GPS off / timeout | error view + "Open Location Settings" → `Geolocator.openLocationSettings()` |
| `ready` | position obtained | `GoogleMap` widget full-screen |

**Map config when ready:**
- `myLocationEnabled: true` — blue dot on user position
- `myLocationButtonEnabled: true` — "My Location" FAB (Android) / button (iOS)
- `zoomControlsEnabled: true` — +/- buttons (Android)
- `compassEnabled: true` — compass shown when map is rotated
- `mapToolbarEnabled: false` — no marker toolbar (no markers in this phase)
- `mapType: MapType.normal`
- Camera starts at `zoom: 15.0` centred on `_position`

**Permission flow (geolocator):**
1. `Geolocator.isLocationServiceEnabled()` — GPS check first
2. `Geolocator.checkPermission()` → if denied → `requestPermission()`
3. `Geolocator.getCurrentPosition(accuracy: high)` with 10 s timeout
4. Any timeout/error → `gpsDisabled` state

### Platform config
| File | Change |
|---|---|
| `android/app/src/main/AndroidManifest.xml` | Added `ACCESS_FINE_LOCATION` + `ACCESS_COARSE_LOCATION` permissions; added `com.google.android.geo.API_KEY` meta-data placeholder |
| `ios/Runner/Info.plist` | Added `NSLocationWhenInUseUsageDescription` + `NSLocationAlwaysAndWhenInUseUsageDescription` |
| `ios/Runner/AppDelegate.swift` | Added `import GoogleMaps` + `GMSServices.provideAPIKey("YOUR_GOOGLE_MAPS_API_KEY")` |

### Router change
`lib/core/router/app_router.dart` — home branch (`/`) now renders `MapPage` instead of `HomePage`.

### Action required on home machine
```bash
cd apps/rider
flutter pub get
flutter analyze
# Replace API key placeholders before running on device:
#   android/app/src/main/AndroidManifest.xml  → YOUR_GOOGLE_MAPS_API_KEY
#   ios/Runner/AppDelegate.swift               → YOUR_GOOGLE_MAPS_API_KEY
```

**NOT implemented (by design):** booking, destination selection, route drawing, driver markers, reverse geocoding, search, place autocomplete, camera tracking, realtime updates.

## Phase 12 — Rider App Skeleton (COMPLETE — analyze pending)

Flutter Rider App skeleton at `apps/rider/`. Scaffolded with `flutter create`, then replaced `lib/` with custom feature structure.

### flutter analyze status
**NOT RUN** — Flutter installed via Homebrew on work machine but `~/.config` owned by root (permission error). User will run `flutter pub get && flutter analyze` on home machine where Flutter is properly set up.

### pubspec.yaml dependencies
- `flutter` SDK
- `cupertino_icons: ^1.0.8`
- `go_router: ^14.0.0` — declarative routing with `StatefulShellRoute` for bottom nav
- dev: `flutter_test`, `flutter_lints: ^6.0.0`

### File structure

```
apps/rider/
  lib/
    main.dart                                        entry point — runApp(RiderApp)
    app.dart                                         RiderApp — MaterialApp.router wired to AppRouter
    core/
      theme/app_theme.dart                           Material 3 theme; primary #1A8C4E (FAIRRIDE green)
      router/app_router.dart                         GoRouter with StatefulShellRoute (3 branches)
    shared/
      widgets/scaffold_with_nav.dart                 NavigationBar shell — Home / Booking / Profile tabs
    features/
      home/presentation/pages/home_page.dart         "Where to?" search bar, recent places, ride categories
      booking/presentation/pages/booking_page.dart   Vehicle selector (Car/Moto/Van), fare breakdown, confirm sheet
      profile/presentation/pages/profile_page.dart   Profile header + rating badge + settings tiles + sign out
```

### Routing (go_router StatefulShellRoute)
- `/` → `HomePage` (tab 0)
- `/booking` → `BookingPage` (tab 1) — also navigable from Home search bar tap
- `/profile` → `ProfilePage` (tab 2)

### Design tokens (AppTheme)
- Primary: `#1A8C4E` (green — fairness/growth)
- Secondary text: `#6B7280`
- Surface: `#F8F9FA`
- Accent background: `#E8F5ED`
- Material 3 `useMaterial3: true`

### What's NOT implemented (by design)
- No API calls, no authentication, no Google Maps
- No state management library
- No real data — all UI is static placeholder content

### To complete on home machine
```bash
cd apps/rider
flutter pub get
flutter analyze
```
IDE shows expected "package not found" errors until pub get runs.

**go.work:** NOT applicable (Flutter is separate from the Go workspace)

## Next Phase (pending CTO approval)
Phase 2.6 — Register / Login use cases (in `services/identity/app/`)
- `RegisterUser` use case: create User, assign default role, return user ID
- `ActivateUser` use case: verify OTP result → Activate, issue access + refresh tokens
- `LoginWithPhone` use case: find user, issue tokens (post-OTP — OTP delivery deferred)
- Application layer wires `TokenService` + `UserRepository` + `RoleRepository`
- No API, no OTP delivery service yet

## Git Checkpoint — MVP Milestone 01 (2026-07-03)

**Commit:** `feat(mvp): complete backend booking flow and rider map foundation`
**Branch:** `main`

### Phases completed in this checkpoint
| Phase | Description | Tests |
|---|---|---|
| 1 / 1.5 | Infrastructure skeleton + architecture audit | 3 shared pkgs |
| 2.1–2.5 | Identity foundation (roles, users, JWT) | 77 |
| 3 | User Profile Service (gRPC + PostgreSQL) | 43 |
| 4 | Driver Profile Service (gRPC + PostgreSQL) | 60 |
| 5 | Vehicle Module | 43 |
| 6 | Driver Availability (Redis GEO heartbeat) | 130 total driver |
| 7 | Trip Foundation | 57 |
| 8 | Dispatch MVP (nearest-driver algorithm) | 56 |
| 9 | Pricing MVP (pure compute) | 33 |
| 10 | Booking Orchestration (full 5-step flow) | 30 |
| 11 | API Gateway (HTTP + JWT + gRPC→HTTP mapping) | 18 |
| 12 | Rider App Skeleton (Flutter, go_router, Material 3) | — |
| 14 | Google Maps integration + permission lifecycle | — |
| 15 | Location Engine (stream-based GPS service) | — |
| 17 | Pickup & Destination Selection UI | — |

**Total backend unit tests:** 460 (0 failures)
**Flutter analyze:** pending — must run on home machine

### What's NOT yet committed / staged
- `flutter pub get` output (`.flutter-plugins`, `.dart_tool/`) — not generated yet; will be gitignored
- Google Maps API key — placeholder in `AndroidManifest.xml` and `AppDelegate.swift`
- Integration tests (Postgres / Redis infra) — skip without env vars; not blocked

---

## Phase H2 — Hardening: Atomic Transactions (COMPLETE — 2026-07-06)

### Problem fixed
`AcceptTripUseCase` and `RequestDispatchUseCase` each performed two cross-table
writes sequentially with no transaction. A failure between the two writes left
the system in a partial state (e.g. trip = `driver_assigned` but dispatch job
still `searching`).

### Solution
Added `repository.Transactor` interface:
```go
type Transactor interface {
    WithinTx(ctx context.Context, fn func(DispatchJobRepository, TripUpdater) error) error
}
```
Implemented by `infrastructure/postgres.Transactor` using `pgx.Tx` + deferred
`Rollback`. Two tx-scoped adapters (`txDispatchRepository`, `txTripUpdater`)
implement the existing interfaces without changing them.

### Files changed
| File | Change |
|------|--------|
| `domain/repository/repository.go` | Added `Transactor` interface |
| `infrastructure/postgres/dispatch_repository.go` | Extracted `scanDispatchJob` package-level helper |
| `infrastructure/postgres/transactor.go` | NEW — `Transactor`, `txDispatchRepository`, `txTripUpdater` |
| `app/accept_trip.go` | Replaced `tripUpdater` with `transactor`; two writes now atomic |
| `app/request_dispatch.go` | Replaced `tripUpdater` with `transactor`; SetSearching + Save now atomic |
| `app/app_test.go` | Added `stubTransactor`, `failingTripUpdater`, `saveFailingJobRepo`; 4 new rollback tests |
| `grpc/handler_test.go` | Added `stubTransactor`; updated `newHandler` constructor |
| `cmd/server/main.go` | Wired `dispatchpostgres.NewTransactor(pool)` |

### Test count
Backend dispatch: **59 tests** (was 55; +4 rollback tests). All pass.

### Rollback flow
```
pool.Begin(ctx) → tx
    fn(txJobRepo, txTripUpdater)
        trips.AssignDriver(...)   ← UPDATE trips   ┐
        jobs.Save(...)            ← UPSERT dispatch │ same tx
                                                    │
    if fn error → tx.Rollback()  ← both reverted  ┘
    else        → tx.Commit()
```
`defer tx.Rollback(ctx)` is a no-op after a successful `Commit`, so the pattern
is safe whether fn succeeds or panics.

### Architecture constraint respected
- No saga, no outbox, no event sourcing introduced.
- Existing `TripUpdater` and `DispatchJobRepository` interfaces unchanged.
- `RejectTripUseCase` and `DispatchEngine` unchanged (they only write to
  `dispatch_jobs`, no cross-entity atomicity risk).

---

## Phase H3-H4 — Hardening: Saga Reliability & Dispatch Lifecycle (COMPLETE — 2026-07-06)

### Part A — Saga Reliability (booking service)

**Problem 1 — Orphaned trips:**
When `BookRide` creates a trip successfully but `RequestDispatch` then fails,
the trip stays in `pending` state with no dispatch job — an orphaned trip.

**Fix:** `BookRideUseCase.Execute` now calls `trip.CancelTrip` (best-effort) when
`RequestDispatch` fails. Also added `CancelTrip(ctx, tripID, reason)` to the
`TripClient` interface and implemented it in `TripAdapter` (wraps the existing
`trippb.CancelTrip` RPC — the Trip service already had this endpoint).

**Problem 2 — Duplicate requests:**
No protection against duplicate `BookRide`, `AcceptDispatchOffer`, or `FinishTrip`
calls (network retries, double-submit).

**Fix:** Added `IdempotencyStore` interface to `booking/app` with a PostgreSQL
implementation in `shared/idempotency.PostgresStore` (persists keys in
`idempotency_keys` table) and an in-memory implementation
(`MemoryIdempotencyStore`) for tests. Use cases gain `WithIdempotency(store)` builder
methods — nil store means no checking (existing constructor unchanged).

| Use Case | Idempotency key |
|---|---|
| `BookRide` | caller-supplied `BookRideInput.IdempotencyKey` (empty = no check) |
| `AcceptDispatchOffer` | `"accept:" + tripID` (natural — one accept per trip) |
| `FinishTrip` | `"finish:" + tripID` (natural — one completion per trip) |

Duplicates return `domainerrors.AlreadyExists("duplicate ... request")`.

**Files changed (Part A):**
| File | Change |
|---|---|
| `booking/app/clients.go` | Added `CancelTrip` to `TripClient` interface |
| `booking/app/idempotency.go` | NEW — `IdempotencyStore` interface + `MemoryIdempotencyStore` |
| `booking/app/book_ride.go` | Compensation logic + idempotency + `WithIdempotency` method |
| `booking/app/accept_reject.go` | Idempotency for `AcceptDispatchOfferUseCase` + `WithIdempotency` |
| `booking/app/finish_trip.go` | Idempotency for `FinishTripUseCase` + `WithIdempotency` |
| `booking/grpc/adapters/trip_adapter.go` | Added `CancelTrip` implementation |
| `booking/grpc/handler_test.go` | Added `CancelTrip` stub method |
| `booking/app/app_test.go` | Added `CancelTrip` to `stubTrip`; 4 new tests |
| `booking/cmd/server/main.go` | Wires `shared/idempotency.PostgresStore` (graceful — boots without DB) |
| `shared/idempotency/store.go` | NEW — `Store` interface + `PostgresStore` + `NewPostgresStoreFromURL` |

**New tests (booking):** `TestBookRide_DispatchError_CompensatesTrip`, `TestBookRide_DuplicateIdempotentRequest`, `TestAcceptDispatchOffer_DuplicateIdempotentRequest`, `TestFinishTrip_DuplicateIdempotentRequest`

**Architecture note:** `shared/idempotency.PostgresStore` satisfies `booking/app.IdempotencyStore`
via Go structural typing — no circular imports. `booking/go.mod` does not need a direct pgx dependency
(the store constructor lives in `shared` which already has pgx).

---

### Part B — Dispatch Engine Lifecycle

**Problems fixed:**
1. `Start()` called twice → two background goroutines (doubled processing rate, double lock contention)
2. `Stop()` returned immediately before the goroutine finished (goroutine leak)
3. A job already being processed could start a second goroutine on the next tick (concurrent duplicate processing)
4. `FindExpiredOffers` error silently swallowed (`return`)
5. `offerNextDriver` error silently discarded (`_ = err`)

**Fixes in `dispatch/app/engine.go`:**
| Mechanism | What it guards |
|---|---|
| `sync.Once` (startOnce) | `Start()` idempotent — only first call creates goroutine |
| `sync.Once` (stopOnce) | `Stop()` idempotent — only first call closes `done` channel |
| `sync.WaitGroup` | `Stop()` waits for the main goroutine AND all in-flight job goroutines |
| `sync.Map` (inFlight) | Skips job if its `JobID` is already being processed |
| Per-job goroutine + `wg.Add(1)` | Each expired job processed concurrently; all jobs waited on by `Stop()` |
| `log.Error()` / `log.Warn()` (zerolog) | All silenced errors now produce structured log lines with `job_id` field |

`processJob` extracted as separate method for clarity. Uses `now` captured at start of `processExpiredOffers` tick (not re-sampled per-job).

**New tests (dispatch engine):**
- `TestEngine_StartCalledTwiceCreatesOneWorker` — verifies `FindExpiredOffers` rate ≤14 over 40 ms with 5 ms tick (would be ~16 with two goroutines)
- `TestEngine_GracefulStop` — verifies `Stop()` blocks while a job goroutine is blocked at `Save`, returns promptly after unblock
- `TestEngine_ConcurrentJobsDeduplication` — verifies only 1 `Save` call while first goroutine is in-flight + engine stopped before unblock (prevents new goroutines from starting)

**Test helpers added:**
`countingJobRepo`, `blockOnSaveJobRepo`, `alwaysExpiredJobRepo`, `composedJobRepo`

### Combined test counts after H3-H4
- `dispatch/app`: **22 tests** (was 19; +3 engine lifecycle)
- `booking/app`: **21 tests** (was 17; +4 saga/idempotency)
- `booking/grpc`: **14 tests** (unchanged — stub updated only)
- All other packages: unchanged

**All modules build and test clean:**
`go test ./services/dispatch/... ./services/booking/... ./shared/...` → 0 failures

---

## Notes
- Implementation mode began 2026-07-01. Documentation phase paused.
- DOC-0003 and DOC-0004 remain pending but are not blocking implementation.
- Architecture pattern: 14 microservice skeletons retained; CTO approved continuing with controlled inter-service dependencies.
- 83 canonical terms defined in DOC-0001, DOC-0001A, DOC-0002.
- PHASE-0-EOS-BLUEPRINT.md folder structure (209 folders) remains valid.
