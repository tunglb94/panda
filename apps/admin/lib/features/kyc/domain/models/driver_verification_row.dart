/// One row of the /admin/driver-verifications table (Phần 1). Vehicle
/// type/service type come from a best-effort join against the driver's
/// separate VehicleVerification record on the backend (may be null if the
/// driver hasn't submitted vehicle info yet).
class DriverVerificationRow {
  const DriverVerificationRow({
    required this.driverId,
    required this.fullName,
    required this.phone,
    required this.nationalIdNumber,
    required this.status,
    required this.submittedAt,
    required this.rejectReason,
    this.vehicleType,
    this.serviceType,
  });

  final String driverId;
  final String fullName;
  final String phone;
  final String nationalIdNumber;
  final String status;
  final DateTime submittedAt;
  final String rejectReason;
  final String? vehicleType;
  final String? serviceType;

  factory DriverVerificationRow.fromJson(Map<String, dynamic> json) => DriverVerificationRow(
        driverId: json['driver_id'] as String? ?? '',
        fullName: json['full_name'] as String? ?? '',
        phone: json['phone'] as String? ?? '',
        nationalIdNumber: json['national_id_number'] as String? ?? '',
        status: json['status'] as String? ?? '',
        submittedAt: DateTime.tryParse(json['submitted_at'] as String? ?? '') ?? DateTime(1970),
        rejectReason: json['reject_reason'] as String? ?? '',
        vehicleType: json['vehicle_type'] as String?,
        serviceType: json['service_type'] as String?,
      );
}
