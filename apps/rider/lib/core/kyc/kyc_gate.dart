import 'package:flutter/foundation.dart';
import '../network/api_client.dart' show ApiClient;

/// Tracks whether the logged-in rider is KYC-approved, driving the
/// app-startup gate: Splash → Check Token → Refresh Token → Load Profile →
/// Check KYC → route.
///
/// [approved] is `null` until the first [refresh] completes — go_router's
/// redirect logic treats `null` as "still resolving, stay on Splash".
class KycGate extends ChangeNotifier {
  bool? _approved;

  bool? get approved => _approved;

  /// Fetches the rider's verification status. 404 (nothing submitted yet)
  /// — or any other failure, e.g. a network hiccup — is treated as "not
  /// approved" rather than thrown: it routes the rider to RiderKycPage,
  /// which has its own error/retry UI, instead of stranding Splash.
  Future<void> refresh(ApiClient apiClient) async {
    bool approved = false;
    try {
      final v = await apiClient.get('/api/v1/rider/verification');
      approved = v['status'] == 'approved';
    } catch (_) {
      approved = false;
    }
    _approved = approved;
    notifyListeners();
  }

  /// Called before a new login so a stale value from a previous
  /// session/account never survives into this one.
  void reset() {
    _approved = null;
    notifyListeners();
  }
}
