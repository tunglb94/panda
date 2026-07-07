/// Outcome of `DriverTripOfferRepository.acceptOffer()`. A dedicated result
/// type (not a `bool`) so a third case — `timeout` — can exist without
/// awkward nullable-bool tricks, and so the repository's data contract stays
/// separate from `TripOfferState` (the UI-facing state machine).
enum DispatchAcceptStatus { success, failed, timeout }

/// Mock result of an accept attempt. No HTTP, no backend — see
/// `docs/project/MVP_DEVELOPMENT_PLAN.md` §2.3.
class DispatchAcceptResult {
  const DispatchAcceptResult({required this.status});

  final DispatchAcceptStatus status;

  bool get isSuccess => status == DispatchAcceptStatus.success;
}
