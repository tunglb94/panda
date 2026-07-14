import '../network/api_client.dart';

class AppVersionInfo {
  const AppVersionInfo({
    required this.minimumVersion,
    required this.latestVersion,
    required this.forceUpdate,
  });

  factory AppVersionInfo.fromJson(Map<String, dynamic> json) => AppVersionInfo(
        minimumVersion: json['minimum_version'] as String? ?? '0.0.0',
        latestVersion: json['latest_version'] as String? ?? '0.0.0',
        forceUpdate: json['force_update'] as bool? ?? false,
      );

  final String minimumVersion;
  final String latestVersion;
  final bool forceUpdate;

  /// True when the running app must not be allowed to continue: either the
  /// backend set an unconditional force_update, or the current version is
  /// below minimumVersion. Flutter decides this — the backend only reports
  /// the three raw fields (see plan's Startup Flow phase).
  bool isBlocked(String currentVersion) {
    if (forceUpdate) return true;
    return _compareVersions(currentVersion, minimumVersion) < 0;
  }
}

/// Fetches the App Version startup-check payload for [app] ("driver" or
/// "rider"). No auth required — this must work before login.
Future<AppVersionInfo> fetchAppVersionInfo(ApiClient apiClient, String app) async {
  final body = await apiClient.get('/api/v1/app/version?app=$app');
  return AppVersionInfo.fromJson(body);
}

/// Compares two dotted-numeric version strings ("1.2.0" vs "1.10.0").
/// Returns negative if [a] < [b], zero if equal, positive if [a] > [b].
/// Non-numeric segments compare as 0 — versions here are always our own
/// x.y.z scheme, never arbitrary user input.
int _compareVersions(String a, String b) {
  final partsA = a.split('.');
  final partsB = b.split('.');
  final length = partsA.length > partsB.length ? partsA.length : partsB.length;
  for (var i = 0; i < length; i++) {
    final na = i < partsA.length ? int.tryParse(partsA[i]) ?? 0 : 0;
    final nb = i < partsB.length ? int.tryParse(partsB[i]) ?? 0 : 0;
    if (na != nb) return na - nb;
  }
  return 0;
}
