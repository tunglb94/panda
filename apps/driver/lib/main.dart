import 'package:flutter/material.dart';
import 'app.dart';
import 'core/auth/auth_state.dart';
import 'core/config/app_config.dart';
import 'core/network/api_client.dart';
import 'core/storage/token_storage.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  final storage = TokenStorage();
  final authState = AuthState();
  await authState.initialize(storage);

  final apiClient = ApiClient(
    baseUrl: AppConfig.apiBaseUrl,
    authState: authState,
  );

  runApp(DriverApp(
    authState: authState,
    tokenStorage: storage,
    apiClient: apiClient,
  ));
}
