import 'kyc_status.dart';

/// Mirrors the backend's `entity.DriverVerification` — the driver's own
/// personal-identity KYC record.
class DriverVerification {
  const DriverVerification({
    required this.fullName,
    required this.dateOfBirth,
    required this.address,
    required this.nationalIdNumber,
    required this.licenseNumber,
    required this.status,
    required this.rejectReason,
  });

  final String fullName;
  final DateTime? dateOfBirth;
  final String address;

  /// CCCD number (Phần 5 — Duplicate Detection). Required at submit.
  final String nationalIdNumber;
  final String licenseNumber;
  final KYCStatus status;
  final String rejectReason;

  static const empty = DriverVerification(
    fullName: '',
    dateOfBirth: null,
    address: '',
    nationalIdNumber: '',
    licenseNumber: '',
    status: KYCStatus.pending,
    rejectReason: '',
  );

  factory DriverVerification.fromJson(Map<String, dynamic> json) => DriverVerification(
        fullName: json['full_name'] as String? ?? '',
        dateOfBirth: DateTime.tryParse(json['date_of_birth'] as String? ?? ''),
        address: json['address'] as String? ?? '',
        nationalIdNumber: json['national_id_number'] as String? ?? '',
        licenseNumber: json['license_number'] as String? ?? '',
        status: KYCStatusX.fromWire(json['status'] as String? ?? ''),
        rejectReason: json['reject_reason'] as String? ?? '',
      );
}
