import 'package:rider/core/network/api_client.dart';

import '../domain/models/contact_info.dart';

/// Phone Call + Contact Card (Part 1 &amp; 4). `getContact` is safe to call
/// anytime a driver is assigned — it only ever returns [ContactInfo.maskedPhone]
/// for display. `call` is the one place a real phone number ever leaves the
/// backend, and only so the caller can immediately hand it to `url_launcher`'s
/// `tel:` scheme — never render it, never persist it.
class ContactRepository {
  const ContactRepository(this._client);

  final ApiClient _client;

  Future<ContactInfo> getContact(String tripId) async {
    final body = await _client.get('/api/v1/rides/$tripId/contact');
    return ContactInfo.fromJson(body);
  }

  /// Returns the real phone number for [tripId]'s other participant.
  /// Callers must pass this straight to `launchUrl(Uri(scheme: 'tel', ...))`
  /// and never store or display it.
  Future<String> call(String tripId) async {
    final body = await _client.post('/api/v1/rides/$tripId/call');
    return body['phone'] as String? ?? '';
  }
}
