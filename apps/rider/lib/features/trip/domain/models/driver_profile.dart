class DriverProfile {
  const DriverProfile({
    required this.vehicleBrand,
    required this.vehicleModel,
    required this.vehicleColor,
    required this.plateNumber,
    this.verificationStatus = '',
  });

  final String vehicleBrand;
  final String vehicleModel;
  final String vehicleColor;
  final String plateNumber;

  /// Raw backend verification status (e.g. "verified", "pending"). Empty
  /// while loading or if the backend omitted it.
  final String verificationStatus;

  bool get isVerified => verificationStatus.toLowerCase() == 'verified';

  String get vehicleDisplay {
    final brand = vehicleBrand.trim();
    final model = vehicleModel.trim();
    if (brand.isEmpty) return model;
    if (model.isEmpty) return brand;
    return '$brand $model';
  }

  static const DriverProfile loading = DriverProfile(
    vehicleBrand: '',
    vehicleModel: 'Đang tải…',
    vehicleColor: '',
    plateNumber: '—',
  );
}
