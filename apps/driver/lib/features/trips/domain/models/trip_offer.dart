import 'rider_info.dart';

/// An incoming trip (dispatch) offer shown to the driver. Mock data only —
/// the Dispatch service already has a real matching engine
/// (`backend/services/dispatch`), but nothing here calls it (see
/// `docs/project/MVP_DEVELOPMENT_PLAN.md` Driver App Roadmap stage D4).
class TripOffer {
  const TripOffer({
    required this.id,
    required this.rider,
    required this.pickupAddress,
    required this.destinationAddress,
    required this.distanceToPickupKm,
    required this.estimatedTripDistanceKm,
    required this.estimatedTripDurationMin,
    required this.estimatedFareCents,
    this.surgeMultiplier,
  });

  final String id;
  final RiderInfo rider;
  final String pickupAddress;
  final String destinationAddress;
  final double distanceToPickupKm;
  final double estimatedTripDistanceKm;
  final double estimatedTripDurationMin;
  final int estimatedFareCents;

  /// Null (or <= 1.0) means no surge — the Surge indicator is only shown
  /// when this is a real multiplier above 1.0.
  final double? surgeMultiplier;

  bool get hasSurge => surgeMultiplier != null && surgeMultiplier! > 1.0;

  String get formattedFare => '\$${(estimatedFareCents / 100).toStringAsFixed(2)}';
}
