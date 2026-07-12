/// The logged-in driver's own profile — fetched from the same
/// `GET /api/v1/drivers/{driverID}/profile` endpoint `apps/rider` already
/// calls to show a *rider* their assigned driver's info. Calling it with
/// the driver's own ID (from `AuthState.driverId`) is not a new API, just a
/// new caller of an existing one.
///
/// The backend does not return a display name, photo, or aggregate rating
/// for a driver anywhere — those fields simply don't exist here. Screens
/// consuming this model must not invent them.
class DriverOwnProfile {
  const DriverOwnProfile({
    required this.driverId,
    required this.vehicleType,
    required this.vehicleBrand,
    required this.vehicleModel,
    required this.vehicleColor,
    required this.plateNumber,
    required this.verificationStatus,
    this.createdAt,
  });

  final String driverId;
  final String vehicleType;
  final String vehicleBrand;
  final String vehicleModel;
  final String vehicleColor;
  final String plateNumber;
  final String verificationStatus;

  /// Real if the backend returned it (optional field) — used for "Gia nhập
  /// từ". Null when the backend omitted it, never guessed.
  final DateTime? createdAt;

  bool get isVerified => verificationStatus.toLowerCase() == 'verified';

  String get vehicleDisplay {
    final brand = vehicleBrand.trim();
    final model = vehicleModel.trim();
    if (brand.isEmpty && model.isEmpty) return 'Chưa cập nhật';
    if (brand.isEmpty) return model;
    if (model.isEmpty) return brand;
    return '$brand $model';
  }

  static const empty = DriverOwnProfile(
    driverId: '',
    vehicleType: '',
    vehicleBrand: '',
    vehicleModel: '',
    vehicleColor: '',
    plateNumber: '',
    verificationStatus: '',
  );
}
