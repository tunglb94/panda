import 'package:flutter/material.dart';
import 'package:rider/core/auth/auth_state.dart';
import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/router/app_router.dart';
import 'package:rider/core/storage/token_storage.dart';
import 'package:rider/core/theme/app_theme.dart';

class RiderApp extends StatelessWidget {
  const RiderApp({
    super.key,
    required this.authState,
    required this.tokenStorage,
    required this.apiClient,
  });

  final AuthState authState;
  final TokenStorage tokenStorage;
  final ApiClient apiClient;

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'FAIRRIDE',
      theme: AppTheme.light,
      routerConfig: AppRouter.create(apiClient: apiClient),
      debugShowCheckedModeBanner: false,
    );
  }
}
