/// Mirrors the backend's `entity.DocumentType` constants
/// (driver/domain/entity/kyc_document.go) exactly.
abstract final class DocumentType {
  static const cccdFront = 'cccd_front';
  static const cccdBack = 'cccd_back';
  static const selfie = 'selfie';
  static const license = 'license'; // GPLX
  static const vehicleRegistration = 'vehicle_registration';
  static const vehicleInsurance = 'vehicle_insurance';
  static const vehicleInspection = 'vehicle_inspection'; // Đăng kiểm — optional
}

/// Document types that carry an expiry date (Phần 2 of the Hardening spec —
/// GPLX, Đăng ký xe, Bảo hiểm, Đăng kiểm). CCCD/Selfie never expire.
const Set<String> expiringDocumentTypes = {
  DocumentType.license,
  DocumentType.vehicleRegistration,
  DocumentType.vehicleInsurance,
  DocumentType.vehicleInspection,
};

String documentTypeLabel(String type) => switch (type) {
      DocumentType.cccdFront => 'CCCD mặt trước',
      DocumentType.cccdBack => 'CCCD mặt sau',
      DocumentType.selfie => 'Ảnh chân dung (Selfie)',
      DocumentType.license => 'Giấy phép lái xe (GPLX)',
      DocumentType.vehicleRegistration => 'Đăng ký xe',
      DocumentType.vehicleInsurance => 'Bảo hiểm xe',
      DocumentType.vehicleInspection => 'Đăng kiểm xe',
      _ => type,
    };

/// One entry in the document-upload checklist (`GET .../verification/documents`).
/// Phần 4/11 — carries version + expiry so the UI can show version history
/// and expiry banners (yellow within 30 days, red once expired).
class DocumentChecklistItem {
  const DocumentChecklistItem({
    required this.documentType,
    required this.uploaded,
    this.version = 0,
    this.expiresAt,
    this.expired = false,
    this.expiringSoon = false,
  });

  final String documentType;
  final bool uploaded;
  final int version;
  final DateTime? expiresAt;
  final bool expired;
  final bool expiringSoon;

  factory DocumentChecklistItem.fromJson(Map<String, dynamic> json) => DocumentChecklistItem(
        documentType: json['document_type'] as String? ?? '',
        uploaded: json['uploaded'] as bool? ?? false,
        version: (json['version'] as num?)?.toInt() ?? 0,
        expiresAt: DateTime.tryParse(json['expires_at'] as String? ?? ''),
        expired: json['expired'] as bool? ?? false,
        expiringSoon: json['expiring_soon'] as bool? ?? false,
      );
}
