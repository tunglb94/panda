/// The UI-facing state machine for an incoming trip offer, covering both
/// the original offer lifecycle (Phase D-03) and the post-Accept dispatch
/// session (Phase D-04):
///
/// ```
/// newOffer -> accepting -> assigned            (accept succeeds)
///                       -> failed   -> newOffer (accept fails, Retry)
///                       -> timeout  -> newOffer (accept times out, Retry)
/// newOffer -> rejected                          (Reject)
/// newOffer -> expired                           (countdown reaches zero)
/// assigned -> navigatingToPickup                 (Start Navigation, Phase D-05)
/// navigatingToPickup -> arrivedAtPickup           (route progress reaches 0%, Phase D-06)
/// ```
///
/// `arrivedAtPickup` is as far as Phase D-06 goes — `passengerBoarding` does
/// not exist yet (see `docs/project/MVP_DEVELOPMENT_PLAN.md` Driver App
/// Roadmap stage D5/D6).
///
/// This is the *only* state machine for the offer's lifecycle — it is
/// deliberately separate from `AsyncStateView`'s Loading/Success/Empty/Error
/// machine (which only governs whether an offer was fetched at all). See
/// `TripsPage` for how the two compose without merging.
enum TripOfferState {
  newOffer,
  accepting,
  assigned,
  rejected,
  expired,
  failed,
  timeout,
  navigatingToPickup,
  arrivedAtPickup,
}

extension TripOfferStateX on TripOfferState {
  String get label => switch (this) {
        TripOfferState.newOffer => 'New Offer',
        TripOfferState.accepting => 'Accepting',
        TripOfferState.assigned => 'Assigned',
        TripOfferState.rejected => 'Rejected',
        TripOfferState.expired => 'Expired',
        TripOfferState.failed => 'Failed',
        TripOfferState.timeout => 'Timeout',
        TripOfferState.navigatingToPickup => 'Navigating to Pickup',
        TripOfferState.arrivedAtPickup => 'Arrived at Pickup',
      };
}
