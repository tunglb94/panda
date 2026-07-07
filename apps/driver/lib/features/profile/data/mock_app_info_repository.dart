import '../domain/models/app_info.dart';

/// "Data" layer backing the Developer page. No HTTP, no backend — this is
/// the whole of the data source. See [AppInfo] for which fields are real
/// vs. mock/placeholder.
class MockAppInfoRepository {
  const MockAppInfoRepository();

  AppInfo current() {
    return AppInfo(
      appVersion: '1.0.0+1 (mock)',
      buildMode: AppInfo.currentBuildMode(),
      flutterVersionPlaceholder: '3.35.4 (placeholder — not read at runtime)',
      environmentPlaceholder: 'development (placeholder)',
    );
  }
}
