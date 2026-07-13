/// A real fare estimate returned by the backend's Pricing service (via the
/// gateway's `POST /api/v1/rides/estimate-fare`, see
/// `PricingRepository.estimateFare`). Every field here is exactly what the
/// backend computed — Flutter performs no fare math of its own; this is a
/// pure DTO.
///
/// Backend's `EstimateFare` RPC has no discount/voucher or surge input at
/// all (no promotion engine is wired to any RPC yet), so this shape
/// intentionally has no `discount`/`originalFare`/`surge` fields — there is
/// nothing to parse, and the UI must never fabricate one. Add those fields
/// here only once the backend response actually carries them.
class FareEstimate {
  const FareEstimate({
    required this.serviceType,
    required this.vehicleType,
    required this.distanceKm,
    required this.durationMinutes,
    required this.baseFare,
    required this.distanceFare,
    required this.timeFare,
    required this.bookingFee,
    required this.rideFare,
    required this.total,
    required this.currencyCode,
    required this.isFinal,
  });

  /// The rider-facing tier this was quoted for (motorcycle/bike_plus/car/car_xl).
  final String serviceType;

  /// The backend Pricing VehicleType actually used (car/motorcycle/van) —
  /// bike_plus/car_xl map onto the nearest real tier server-side.
  final String vehicleType;

  final double distanceKm;
  final double durationMinutes;

  final int baseFare;
  final int distanceFare;
  final int timeFare;
  final int bookingFee;
  final int rideFare;
  final int total;
  final String currencyCode;

  /// false for a pre-booking estimate, true for a post-trip final fare —
  /// this endpoint always returns false.
  final bool isFinal;

  factory FareEstimate.fromJson(Map<String, dynamic> json) => FareEstimate(
        serviceType: json['service_type'] as String,
        vehicleType: json['vehicle_type'] as String,
        distanceKm: (json['distance_km'] as num).toDouble(),
        durationMinutes: (json['duration_minutes'] as num).toDouble(),
        baseFare: json['base_fare'] as int,
        distanceFare: json['distance_fare'] as int,
        timeFare: json['time_fare'] as int,
        bookingFee: json['booking_fee'] as int,
        rideFare: json['ride_fare'] as int,
        total: json['total'] as int,
        currencyCode: json['currency_code'] as String,
        isFinal: json['is_final'] as bool,
      );
}
