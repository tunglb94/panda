/// One entry of the document checklist shown in the review Drawer (Phần 2) —
/// mirrors the backend's `documentChecklistItemJSONForAdmin` shape.
/// `documentId` is only present when [uploaded] is true; use it to fetch the
/// image bytes via `GET /api/v1/admin/verifications/documents/{documentId}`.
class KYCDocumentItem {
  const KYCDocumentItem({
    required this.documentType,
    required this.uploaded,
    this.uploadedAt,
    this.version,
    this.documentId,
    this.expiresAt,
    this.expired,
    this.expiringSoon,
  });

  final String documentType;
  final bool uploaded;
  final DateTime? uploadedAt;
  final int? version;
  final String? documentId;
  final DateTime? expiresAt;
  final bool? expired;
  final bool? expiringSoon;

  factory KYCDocumentItem.fromJson(Map<String, dynamic> json) => KYCDocumentItem(
        documentType: json['document_type'] as String? ?? '',
        uploaded: json['uploaded'] as bool? ?? false,
        uploadedAt: json['uploaded_at'] != null ? DateTime.tryParse(json['uploaded_at'] as String) : null,
        version: json['version'] as int?,
        documentId: json['document_id'] as String?,
        expiresAt: json['expires_at'] != null ? DateTime.tryParse(json['expires_at'] as String) : null,
        expired: json['expired'] as bool?,
        expiringSoon: json['expiring_soon'] as bool?,
      );
}

/// Vietnamese display labels for each document type (Phần 2's required set
/// plus vehicle_inspection, which the backend checklist also tracks).
const kDocumentTypeLabels = {
  'cccd_front': 'CCCD mặt trước',
  'cccd_back': 'CCCD mặt sau',
  'selfie': 'Ảnh chân dung (Selfie)',
  'license': 'GPLX',
  'vehicle_registration': 'Đăng ký xe',
  'vehicle_insurance': 'Bảo hiểm xe',
  'vehicle_inspection': 'Đăng kiểm xe',
};
