class DriverProfile {
  const DriverProfile({
    required this.vehicleBrand,
    required this.vehicleModel,
    required this.vehicleColor,
    required this.plateNumber,
  });

  final String vehicleBrand;
  final String vehicleModel;
  final String vehicleColor;
  final String plateNumber;

  String get vehicleDisplay {
    final brand = vehicleBrand.trim();
    final model = vehicleModel.trim();
    if (brand.isEmpty) return model;
    if (model.isEmpty) return brand;
    return '$brand $model';
  }

  static const DriverProfile loading = DriverProfile(
    vehicleBrand: '',
    vehicleModel: 'Loading…',
    vehicleColor: '',
    plateNumber: '—',
  );
}
