import 'audit_log_entry.dart';
import 'kyc_document_item.dart';

/// Personal-info half of the combined detail response (Phần 2's info
/// panel — Tên/Ngày sinh/CCCD/GPLX).
class DriverVerificationDetail {
  const DriverVerificationDetail({
    required this.id,
    required this.fullName,
    required this.dateOfBirth,
    required this.address,
    required this.nationalIdNumber,
    required this.licenseNumber,
    required this.status,
    required this.submittedAt,
    required this.rejectReason,
  });

  final String id;
  final String fullName;
  final String dateOfBirth;
  final String address;
  final String nationalIdNumber;
  final String licenseNumber;
  final String status;
  final DateTime submittedAt;
  final String rejectReason;

  factory DriverVerificationDetail.fromJson(Map<String, dynamic> json) => DriverVerificationDetail(
        id: json['id'] as String? ?? '',
        fullName: json['full_name'] as String? ?? '',
        dateOfBirth: json['date_of_birth'] as String? ?? '',
        address: json['address'] as String? ?? '',
        nationalIdNumber: json['national_id_number'] as String? ?? '',
        licenseNumber: json['license_number'] as String? ?? '',
        status: json['status'] as String? ?? '',
        submittedAt: DateTime.tryParse(json['submitted_at'] as String? ?? '') ?? DateTime(1970),
        rejectReason: json['reject_reason'] as String? ?? '',
      );
}

/// Vehicle-info half of the combined detail response (Phần 2's info panel —
/// Biển số/VIN/Số máy/Số khung/Ngày hết hạn).
class VehicleVerificationDetail {
  const VehicleVerificationDetail({
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
    required this.status,
    required this.expiredAt,
  });

  final String vehicleType;
  final String serviceType;
  final String brand;
  final String model;
  final int year;
  final String color;
  final String plateNumber;
  final String vin;
  final String engineNumber;
  final String chassisNumber;
  final String licenseClass;
  final String status;
  final DateTime? expiredAt;

  factory VehicleVerificationDetail.fromJson(Map<String, dynamic> json) => VehicleVerificationDetail(
        vehicleType: json['vehicle_type'] as String? ?? '',
        serviceType: json['service_type'] as String? ?? '',
        brand: json['brand'] as String? ?? '',
        model: json['model'] as String? ?? '',
        year: json['year'] as int? ?? 0,
        color: json['color'] as String? ?? '',
        plateNumber: json['plate_number'] as String? ?? '',
        vin: json['vin'] as String? ?? '',
        engineNumber: json['engine_number'] as String? ?? '',
        chassisNumber: json['chassis_number'] as String? ?? '',
        licenseClass: json['license_class'] as String? ?? '',
        status: json['status'] as String? ?? '',
        expiredAt: json['expired_at'] != null ? DateTime.tryParse(json['expired_at'] as String) : null,
      );
}

/// `GET /api/v1/admin/verifications/drivers/{driverID}/detail` — everything
/// the review Drawer/Dialog needs in one call (Phần 2). Each sub-object is
/// independently nullable: a driver may have submitted personal info but not
/// vehicle info yet, or vice versa.
class KYCDetail {
  const KYCDetail({
    required this.driverId,
    required this.phone,
    this.driverVerification,
    this.vehicleVerification,
    required this.documents,
    required this.auditLog,
  });

  final String driverId;
  final String phone;
  final DriverVerificationDetail? driverVerification;
  final VehicleVerificationDetail? vehicleVerification;
  final List<KYCDocumentItem> documents;
  final List<AuditLogEntry> auditLog;

  factory KYCDetail.fromJson(Map<String, dynamic> json) => KYCDetail(
        driverId: json['driver_id'] as String? ?? '',
        phone: json['phone'] as String? ?? '',
        driverVerification: json['driver_verification'] != null
            ? DriverVerificationDetail.fromJson(json['driver_verification'] as Map<String, dynamic>)
            : null,
        vehicleVerification: json['vehicle_verification'] != null
            ? VehicleVerificationDetail.fromJson(json['vehicle_verification'] as Map<String, dynamic>)
            : null,
        documents: (json['documents'] as List<dynamic>? ?? [])
            .map((e) => KYCDocumentItem.fromJson(e as Map<String, dynamic>))
            .toList(),
        auditLog: (json['audit_log'] as List<dynamic>? ?? [])
            .map((e) => AuditLogEntry.fromJson(e as Map<String, dynamic>))
            .toList(),
      );
}
