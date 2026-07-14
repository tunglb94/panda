abstract final class AppConfig {
  static const apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'https://mls-management-vendors-parliament.trycloudflare.com',
  );
  static const googleMapsApiKey = String.fromEnvironment(
    'GOOGLE_MAPS_API_KEY',
    defaultValue: '',
  );

  /// OAuth 2.0 Client ID from Google Cloud Console (Android/iOS/Web —
  /// whichever matches the build target). Empty by default: the "Đăng nhập
  /// bằng Google" button disables itself and shows a hint until this is
  /// supplied via --dart-define=GOOGLE_CLIENT_ID=... (see plan's Known Gaps).
  static const googleClientId = String.fromEnvironment(
    'GOOGLE_CLIENT_ID',
    defaultValue: '',
  );

  /// Must track pubspec.yaml's `version:` (currently `1.0.0+1`) — used for
  /// the App Version startup check (Splash) and sent as device metadata on
  /// login. There is no runtime way to read pubspec's version without the
  /// `package_info_plus` plugin, deliberately not added here (see plan's
  /// Known Gaps) — this constant must be bumped by hand alongside pubspec.
  static const appVersion = '1.0.0';
}
