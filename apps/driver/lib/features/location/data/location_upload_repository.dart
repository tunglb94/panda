import '../../../core/network/api_client.dart';

class LocationUploadRepository {
  LocationUploadRepository({required ApiClient apiClient}) : _client = apiClient;

  final ApiClient _client;

  Future<void> uploadLocation(double lat, double lon) async {
    await _client.post('/api/v1/driver/location', body: {'lat': lat, 'lon': lon});
  }
}
