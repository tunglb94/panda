import '../../../core/network/api_client.dart';
import '../domain/models/document_type.dart';
import '../domain/models/driver_verification.dart';
import '../domain/models/vehicle_verification.dart';

/// Driver KYC + Vehicle Verification. Backed entirely by real endpoints —
/// no mock, no fake OCR, no hardcoded approval. `getDriverVerification`/
/// `getVehicleVerification` return null (not an error) when the driver
/// hasn't submitted yet, so the wizard can distinguish "nothing submitted"
/// from a real network failure.
class KYCRepository {
  const KYCRepository(this._client);

  final ApiClient _client;

  Future<DriverVerification?> getDriverVerification() async {
    try {
      final body = await _client.get('/api/v1/driver/verification');
      return DriverVerification.fromJson(body);
    } on ApiException catch (e) {
      if (e.statusCode == 404) return null;
      rethrow;
    }
  }

  Future<DriverVerification> submitDriverVerification({
    required String fullName,
    required DateTime dateOfBirth,
    required String address,
    required String nationalIdNumber,
    String licenseNumber = '',
  }) async {
    final body = await _client.post('/api/v1/driver/verification', body: _driverBody(fullName, dateOfBirth, address, nationalIdNumber, licenseNumber));
    return DriverVerification.fromJson(body);
  }

  Future<DriverVerification> updateDriverVerification({
    required String fullName,
    required DateTime dateOfBirth,
    required String address,
    required String nationalIdNumber,
    String licenseNumber = '',
  }) async {
    final body = await _client.put('/api/v1/driver/verification', body: _driverBody(fullName, dateOfBirth, address, nationalIdNumber, licenseNumber));
    return DriverVerification.fromJson(body);
  }

  Map<String, dynamic> _driverBody(String fullName, DateTime dob, String address, String nationalIdNumber, String licenseNumber) => {
        'full_name': fullName,
        'date_of_birth': _dateOnly(dob),
        'address': address,
        'national_id_number': nationalIdNumber,
        'license_number': licenseNumber,
      };

  Future<VehicleVerification?> getVehicleVerification() async {
    try {
      final body = await _client.get('/api/v1/vehicle/verification');
      return VehicleVerification.fromJson(body);
    } on ApiException catch (e) {
      if (e.statusCode == 404) return null;
      rethrow;
    }
  }

  Future<VehicleVerification> submitVehicleVerification(VehicleVerificationInput input) async {
    final body = await _client.post('/api/v1/vehicle/verification', body: input.toJson());
    return VehicleVerification.fromJson(body);
  }

  Future<VehicleVerification> updateVehicleVerification(VehicleVerificationInput input) async {
    final body = await _client.put('/api/v1/vehicle/verification', body: input.toJson());
    return VehicleVerification.fromJson(body);
  }

  Future<List<DocumentChecklistItem>> listDocuments() async {
    final body = await _client.get('/api/v1/driver/verification/documents');
    final raw = (body['documents'] as List<dynamic>?) ?? const [];
    return raw.map((e) => DocumentChecklistItem.fromJson(e as Map<String, dynamic>)).toList();
  }

  /// Phần 4/11 — upload history for one document type, newest first.
  Future<List<DocumentChecklistItem>> listDocumentVersions(String documentType) async {
    final body = await _client.get('/api/v1/driver/verification/documents/$documentType/versions');
    final raw = (body['versions'] as List<dynamic>?) ?? const [];
    return raw.map((e) => DocumentChecklistItem.fromJson(e as Map<String, dynamic>)).toList();
  }

  /// expiresAt (Phần 2) is only meaningful for [expiringDocumentTypes] — the
  /// backend silently ignores it for other types, so callers don't need to
  /// branch on documentType before passing it.
  Future<void> uploadDocument({
    required String documentType,
    required List<int> bytes,
    required String filename,
    DateTime? expiresAt,
  }) async {
    await _client.postMultipart(
      '/api/v1/driver/verification/documents',
      fileFieldName: 'file',
      bytes: bytes,
      filename: filename,
      fields: {
        'document_type': documentType,
        if (expiresAt != null) 'expires_at': _dateOnly(expiresAt),
      },
    );
  }

  static String _dateOnly(DateTime d) =>
      '${d.year.toString().padLeft(4, '0')}-${d.month.toString().padLeft(2, '0')}-${d.day.toString().padLeft(2, '0')}';
}

/// Input for submit/update vehicle verification.
class VehicleVerificationInput {
  const VehicleVerificationInput({
    required this.vehicleType,
    required this.serviceType,
    required this.brand,
    required this.model,
    required this.year,
    required this.color,
    required this.plateNumber,
    this.vin = '',
    this.engineNumber = '',
    this.chassisNumber = '',
    required this.licenseClass,
    required this.rideEnabled,
    required this.deliveryEnabled,
  });

  final String vehicleType;
  final String serviceType;
  final String brand;
  final String model;
  final int year;
  final String color;
  final String plateNumber;

  /// Optional vehicle-identity fields (Phần 6).
  final String vin;
  final String engineNumber;
  final String chassisNumber;

  final String licenseClass;
  final bool rideEnabled;
  final bool deliveryEnabled;

  Map<String, dynamic> toJson() => {
        'vehicle_type': vehicleType,
        'service_type': serviceType,
        'brand': brand,
        'model': model,
        'year': year,
        'color': color,
        'plate_number': plateNumber,
        'vin': vin,
        'engine_number': engineNumber,
        'chassis_number': chassisNumber,
        'license_class': licenseClass,
        'ride_enabled': rideEnabled,
        'delivery_enabled': deliveryEnabled,
      };
}
