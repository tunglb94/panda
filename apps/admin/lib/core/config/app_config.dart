abstract final class AppConfig {
  static const apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'https://mls-management-vendors-parliament.trycloudflare.com',
  );
}
