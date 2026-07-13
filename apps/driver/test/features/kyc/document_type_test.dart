// Unit tests for the Driver KYC Hardening spec (Phần 2/4) — document
// checklist items must carry version + expiry metadata parsed from the
// gateway's response, never a storage path.
import 'package:flutter_test/flutter_test.dart';

import 'package:driver/features/kyc/domain/models/document_type.dart';

void main() {
  test('DocumentChecklistItem.fromJson parses version and expiry fields', () {
    final item = DocumentChecklistItem.fromJson({
      'document_type': 'vehicle_insurance',
      'uploaded': true,
      'uploaded_at': '2026-01-01T00:00:00Z',
      'version': 2,
      'expires_at': '2026-06-01',
      'expired': false,
      'expiring_soon': true,
    });

    expect(item.uploaded, isTrue);
    expect(item.version, 2);
    expect(item.expiresAt, DateTime.parse('2026-06-01'));
    expect(item.expired, isFalse);
    expect(item.expiringSoon, isTrue);
  });

  test('DocumentChecklistItem.fromJson defaults safely when not uploaded', () {
    final item = DocumentChecklistItem.fromJson({
      'document_type': 'selfie',
      'uploaded': false,
    });

    expect(item.uploaded, isFalse);
    expect(item.version, 0);
    expect(item.expiresAt, isNull);
    expect(item.expired, isFalse);
    expect(item.expiringSoon, isFalse);
  });

  test('expiringDocumentTypes matches the backend expiry-eligible set', () {
    expect(expiringDocumentTypes, {
      DocumentType.license,
      DocumentType.vehicleRegistration,
      DocumentType.vehicleInsurance,
      DocumentType.vehicleInspection,
    });
    expect(expiringDocumentTypes.contains(DocumentType.cccdFront), isFalse);
    expect(expiringDocumentTypes.contains(DocumentType.selfie), isFalse);
  });
}
