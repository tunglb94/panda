/// Data shown on the Home dashboard: driver identity, vehicle, and today's
/// stats. Mock data only — the Driver service already has a real
/// `DriverProfile` entity (`backend/services/driver`), but nothing here
/// calls it (see `docs/project/MVP_DEVELOPMENT_PLAN.md` §2.3).
class DriverHomeSummary {
  const DriverHomeSummary({
    required this.driverName,
    required this.rating,
    required this.vehicleModel,
    required this.plateNumber,
    required this.completedTripsToday,
    required this.earningsTodayCents,
    required this.onlineDurationMinutes,
  });

  final String driverName;
  final double rating;

  /// Null when the driver has not registered a vehicle yet — this is what
  /// the dashboard's Empty state represents (see `HomePage`'s
  /// `AsyncStateView.isEmpty`).
  final String? vehicleModel;
  final String? plateNumber;

  final int completedTripsToday;
  final int earningsTodayCents;
  final int onlineDurationMinutes;

  bool get hasVehicle => vehicleModel != null && plateNumber != null;

  String get avatarInitial =>
      driverName.isNotEmpty ? driverName[0].toUpperCase() : '?';

  String get formattedEarningsToday =>
      '\$${(earningsTodayCents / 100).toStringAsFixed(2)}';

  String get formattedOnlineDuration {
    final hours = onlineDurationMinutes ~/ 60;
    final minutes = onlineDurationMinutes % 60;
    if (hours == 0) return '${minutes}m';
    return '${hours}h ${minutes}m';
  }
}
