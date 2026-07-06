/// A driver as shown to the rider during an active trip.
///
/// Mock data only — the Driver service already has a real `DriverProfile`
/// entity (`backend/services/driver`), but there is no API wiring here (see
/// `docs/project/MVP_DEVELOPMENT_PLAN.md` Rider App Roadmap stage R5).
class MockDriver {
  const MockDriver({
    required this.name,
    required this.vehicleModel,
    required this.plateNumber,
    required this.rating,
  });

  final String name;
  final String vehicleModel;
  final String plateNumber;
  final double rating;

  /// Single-letter placeholder shown in a `CircleAvatar` in place of a real
  /// driver photo — no image asset or network fetch is involved.
  String get avatarInitial => name.isNotEmpty ? name[0].toUpperCase() : '?';
}
