import 'rider_info.dart';
import 'trip_offer.dart';

/// Sample offer shared by `DriverTripOfferRepository` (the "normal" demo
/// mode) and the state preview pages, so both use exactly the same data.
class MockTripOfferCatalog {
  const MockTripOfferCatalog._();

  static const TripOffer sample = TripOffer(
    id: 'offer-1',
    rider: RiderInfo(name: 'Alex Rider', rating: 4.9),
    pickupAddress: 'District 1 Market',
    destinationAddress: 'Tan Son Nhat Airport',
    distanceToPickupKm: 1.8,
    estimatedTripDistanceKm: 11.2,
    estimatedTripDurationMin: 26,
    estimatedFareCents: 950,
    surgeMultiplier: 1.5,
  );
}
