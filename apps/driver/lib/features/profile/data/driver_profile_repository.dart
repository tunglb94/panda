import '../../../core/network/api_client.dart';
import '../domain/models/driver_own_profile.dart';

/// Fetches the logged-in driver's own profile via the existing
/// `GET /api/v1/drivers/{driverID}/profile` endpoint. No new API.
class DriverProfileRepository {
  const DriverProfileRepository(this._client);

  final ApiClient _client;

  Future<DriverOwnProfile> fetchOwnProfile(String driverId) async {
    final body = await _client.get('/api/v1/drivers/$driverId/profile');
    DateTime? createdAt;
    final rawCreatedAt = body['created_at'] as String?;
    if (rawCreatedAt != null) {
      try {
        createdAt = DateTime.parse(rawCreatedAt).toLocal();
      } catch (_) {
        createdAt = null;
      }
    }
    return DriverOwnProfile(
      driverId: body['driver_id'] as String? ?? driverId,
      vehicleType: body['vehicle_type'] as String? ?? '',
      vehicleBrand: body['vehicle_brand'] as String? ?? '',
      vehicleModel: body['vehicle_model'] as String? ?? '',
      vehicleColor: body['vehicle_color'] as String? ?? '',
      plateNumber: body['plate_number'] as String? ?? '',
      verificationStatus: body['verification_status'] as String? ?? '',
      createdAt: createdAt,
    );
  }
}
