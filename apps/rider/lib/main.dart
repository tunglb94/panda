import 'package:flutter/material.dart';
import 'package:rider/app.dart';
import 'package:rider/core/auth/auth_state.dart';
import 'package:rider/core/config/app_config.dart';
import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/routing/osrm_route_service.dart';
import 'package:rider/core/storage/token_storage.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  final tokenStorage = TokenStorage();
  final authState = AuthState();
  await authState.initialize(tokenStorage);

  final apiClient = ApiClient(
    baseUrl: AppConfig.apiBaseUrl,
    authState: authState,
  );

  // OSRM's free public routing server — no API key/billing needed (Google
  // Directions is unusable right now: the Cloud project's Billing is
  // broken). See osrm_route_service.dart's doc comment.
  final routeProvider = const OsrmRouteProvider();

  runApp(RiderApp(
    authState: authState,
    tokenStorage: tokenStorage,
    apiClient: apiClient,
    routeProvider: routeProvider,
  ));
}
