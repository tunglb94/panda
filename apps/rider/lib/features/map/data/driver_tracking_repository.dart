import '../../../core/network/api_client.dart';

class DriverLocation {
  const DriverLocation({
    required this.lat,
    required this.lon,
    required this.isActive,
  });

  final double lat;
  final double lon;
  final bool isActive;
}

class DriverTrackingRepository {
  DriverTrackingRepository({required ApiClient apiClient})
      : _client = apiClient;

  final ApiClient _client;

  Future<DriverLocation> getDriverLocation(String driverID) async {
    final data = await _client.get('/api/v1/driver/$driverID/location');
    return DriverLocation(
      lat: (data['lat'] as num).toDouble(),
      lon: (data['lon'] as num).toDouble(),
      isActive: data['is_active'] as bool,
    );
  }
}
