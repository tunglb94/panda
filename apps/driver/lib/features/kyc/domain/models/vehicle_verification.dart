import 'kyc_status.dart';

/// Mirrors the backend's `entity.VehicleVerification`.
class VehicleVerification {
  const VehicleVerification({
    required this.vehicleType,
    required this.serviceType,
    required this.brand,
    required this.model,
    required this.year,
    required this.color,
    required this.plateNumber,
    required this.vin,
    required this.engineNumber,
    required this.chassisNumber,
    required this.licenseClass,
    required this.rideEnabled,
    required this.deliveryEnabled,
    required this.permissions,
    required this.status,
    required this.rejectReason,
  });

  final String vehicleType;
  final String serviceType;
  final String brand;
  final String model;
  final int year;
  final String color;
  final String plateNumber;

  /// Optional vehicle-identity fields (Phần 6) — never required, unique
  /// across vehicles when present.
  final String vin;
  final String engineNumber;
  final String chassisNumber;

  final String licenseClass;
  final bool rideEnabled;
  final bool deliveryEnabled;

  /// Derived ServicePermissions (Phần 8) — e.g. "ride_bike", "delivery_car".
  final List<String> permissions;

  final KYCStatus status;
  final String rejectReason;

  static const empty = VehicleVerification(
    vehicleType: '',
    serviceType: '',
    brand: '',
    model: '',
    year: 0,
    color: '',
    plateNumber: '',
    vin: '',
    engineNumber: '',
    chassisNumber: '',
    licenseClass: '',
    rideEnabled: false,
    deliveryEnabled: false,
    permissions: [],
    status: KYCStatus.pending,
    rejectReason: '',
  );

  factory VehicleVerification.fromJson(Map<String, dynamic> json) => VehicleVerification(
        vehicleType: json['vehicle_type'] as String? ?? '',
        serviceType: json['service_type'] as String? ?? '',
        brand: json['brand'] as String? ?? '',
        model: json['model'] as String? ?? '',
        year: (json['year'] as num?)?.toInt() ?? 0,
        color: json['color'] as String? ?? '',
        plateNumber: json['plate_number'] as String? ?? '',
        vin: json['vin'] as String? ?? '',
        engineNumber: json['engine_number'] as String? ?? '',
        chassisNumber: json['chassis_number'] as String? ?? '',
        licenseClass: json['license_class'] as String? ?? '',
        rideEnabled: json['ride_enabled'] as bool? ?? false,
        deliveryEnabled: json['delivery_enabled'] as bool? ?? false,
        permissions: ((json['permissions'] as List<dynamic>?) ?? const [])
            .map((e) => e as String)
            .toList(),
        status: KYCStatusX.fromWire(json['status'] as String? ?? ''),
        rejectReason: json['reject_reason'] as String? ?? '',
      );
}
