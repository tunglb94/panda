abstract final class AppConfig {
  static const apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'https://basement-nevertheless-tahoe-manchester.trycloudflare.com',
  );
}
