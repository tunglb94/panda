import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:rider/core/auth/auth_state.dart';
import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/router/app_router.dart';
import 'package:rider/core/routing/route_provider.dart';
import 'package:rider/core/storage/token_storage.dart';
import 'package:rider/core/theme/app_theme.dart';

class RiderApp extends StatefulWidget {
  const RiderApp({
    super.key,
    required this.authState,
    required this.tokenStorage,
    required this.apiClient,
    required this.routeProvider,
  });

  final AuthState authState;
  final TokenStorage tokenStorage;
  final ApiClient apiClient;
  final RouteProvider routeProvider;

  @override
  State<RiderApp> createState() => _RiderAppState();
}

class _RiderAppState extends State<RiderApp> {
  late final GoRouter _router;

  @override
  void initState() {
    super.initState();
    _router = AppRouter.create(
      authState: widget.authState,
      tokenStorage: widget.tokenStorage,
      apiClient: widget.apiClient,
      routeProvider: widget.routeProvider,
    );
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'FAIRRIDE',
      theme: AppTheme.light,
      routerConfig: _router,
      debugShowCheckedModeBanner: false,
    );
  }
}
