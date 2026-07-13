// Unit tests for Phần 8 (Driver KYC spec) — the Rider Contact Card must
// parse the gateway's enriched contact response (`is_verified`, `joined_at`,
// `trip_count`) without ever needing/exposing any raw KYC document.
import 'package:flutter_test/flutter_test.dart';

import 'package:rider/features/contact/domain/models/contact_info.dart';

void main() {
  test('ContactInfo.fromJson parses verification, join date, and trip count', () {
    final info = ContactInfo.fromJson({
      'name': 'Trần Văn B',
      'masked_phone': '090****123',
      'rating': 4.8,
      'rating_count': 120,
      'vehicle_type': 'motorcycle',
      'vehicle_brand': 'Honda',
      'vehicle_model': 'Wave',
      'plate_number': '59-X1 123.45',
      'is_verified': true,
      'joined_at': '2024-03-15T00:00:00Z',
      'trip_count': 342,
    });

    expect(info.isVerified, isTrue);
    expect(info.joinedAt, DateTime.utc(2024, 3, 15));
    expect(info.tripCount, 342);
  });

  test('ContactInfo.fromJson defaults verification fields safely when absent', () {
    final info = ContactInfo.fromJson({
      'name': 'Trần Văn B',
      'masked_phone': '090****123',
      'rating': 0,
      'rating_count': 0,
    });

    expect(info.isVerified, isFalse);
    expect(info.joinedAt, isNull);
    expect(info.tripCount, 0);
  });
}
