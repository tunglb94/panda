import '../../../core/network/api_client.dart';

class AvailabilityResult {
  const AvailabilityResult({required this.isOnline, required this.driverId});

  final bool isOnline;
  final String driverId;

  factory AvailabilityResult.fromJson(Map<String, dynamic> json) =>
      AvailabilityResult(
        isOnline: json['is_online'] as bool? ?? false,
        driverId: json['driver_id'] as String? ?? '',
      );
}

class AvailabilityRepository {
  const AvailabilityRepository(this._client);

  final ApiClient _client;

  Future<AvailabilityResult> goOnline() async {
    final body = await _client.post('/api/v1/driver/go-online');
    return AvailabilityResult.fromJson(body);
  }

  Future<AvailabilityResult> goOffline() async {
    final body = await _client.post('/api/v1/driver/go-offline');
    return AvailabilityResult.fromJson(body);
  }

  Future<AvailabilityResult> getAvailability() async {
    final body = await _client.get('/api/v1/driver/availability');
    return AvailabilityResult.fromJson(body);
  }
}
