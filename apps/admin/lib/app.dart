import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'core/auth/auth_state.dart';
import 'core/network/api_client.dart';
import 'core/router/app_router.dart';
import 'core/storage/token_storage.dart';
import 'core/theme/app_theme.dart';

class AdminApp extends StatefulWidget {
  const AdminApp({
    super.key,
    required this.authState,
    required this.tokenStorage,
    required this.apiClient,
  });

  final AuthState authState;
  final TokenStorage tokenStorage;
  final ApiClient apiClient;

  @override
  State<AdminApp> createState() => _AdminAppState();
}

class _AdminAppState extends State<AdminApp> {
  late final GoRouter _router;

  @override
  void initState() {
    super.initState();
    _router = AppRouter.create(
      authState: widget.authState,
      tokenStorage: widget.tokenStorage,
      apiClient: widget.apiClient,
    );
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'Panda Admin',
      theme: AppTheme.light,
      routerConfig: _router,
      debugShowCheckedModeBanner: false,
    );
  }
}
