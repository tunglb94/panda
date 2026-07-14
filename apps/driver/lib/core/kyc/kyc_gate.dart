import 'package:flutter/foundation.dart';
import '../network/api_client.dart' show ApiClient;

/// Tracks whether the logged-in driver is fully cleared for Home:
/// `driver_enabled` on the account (see `GET /auth/me`) AND both identity
/// and vehicle KYC verification approved. Drives the app-startup gate:
/// Splash → Check Token → Refresh Token → Load Profile → Check KYC → route.
///
/// [approved] is `null` until the first [refresh] completes — go_router's
/// redirect logic treats `null` as "still resolving, stay on Splash".
class KycGate extends ChangeNotifier {
  bool? _approved;

  bool? get approved => _approved;

  /// Fetches driver_enabled plus the driver's and vehicle's verification
  /// status. 404 on either verification endpoint (nothing submitted yet) —
  /// or any other failure, e.g. a network hiccup — is treated as "not
  /// approved" rather than thrown: it routes the driver to KYCStatusPage,
  /// which has its own real error/retry UI, instead of stranding Splash on
  /// an unhandled exception. A pre-existing session whose account never
  /// went through a fresh driver-app login (see plan's Known Gaps) may show
  /// driver_enabled=false until they log in again.
  Future<void> refresh(ApiClient apiClient) async {
    bool driverEnabled = false;
    try {
      final me = await apiClient.get('/api/v1/auth/me');
      driverEnabled = me['driver_enabled'] == true;
    } catch (_) {
      driverEnabled = false;
    }
    bool driverApproved = false;
    bool vehicleApproved = false;
    try {
      final driver = await apiClient.get('/api/v1/driver/verification');
      driverApproved = driver['status'] == 'approved';
    } catch (_) {
      driverApproved = false;
    }
    try {
      final vehicle = await apiClient.get('/api/v1/vehicle/verification');
      vehicleApproved = vehicle['status'] == 'approved';
    } catch (_) {
      vehicleApproved = false;
    }
    _approved = driverEnabled && driverApproved && vehicleApproved;
    notifyListeners();
  }

  /// Called on logout so a subsequent login (possibly a different account)
  /// re-resolves from scratch instead of reusing a stale `approved` value.
  void reset() {
    _approved = null;
    notifyListeners();
  }
}
