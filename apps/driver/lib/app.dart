import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'core/auth/auth_state.dart';
import 'core/network/api_client.dart';
import 'core/router/app_router.dart';
import 'core/storage/token_storage.dart';
import 'core/theme/app_theme.dart';
import 'features/location/data/location_upload_repository.dart';
import 'features/location/services/location_upload_service.dart';

class DriverApp extends StatefulWidget {
  const DriverApp({
    super.key,
    required this.authState,
    required this.tokenStorage,
    required this.apiClient,
  });

  final AuthState authState;
  final TokenStorage tokenStorage;
  final ApiClient apiClient;

  @override
  State<DriverApp> createState() => _DriverAppState();
}

class _DriverAppState extends State<DriverApp> {
  late final LocationUploadService _uploadService;
  late final GoRouter _router;

  @override
  void initState() {
    super.initState();
    _uploadService = LocationUploadService(
      repository: LocationUploadRepository(apiClient: widget.apiClient),
    );
    _router = AppRouter.create(
      authState: widget.authState,
      tokenStorage: widget.tokenStorage,
      apiClient: widget.apiClient,
      uploadService: _uploadService,
    );
  }

  @override
  void dispose() {
    _uploadService.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'FAIRRIDE Driver',
      theme: AppTheme.light,
      routerConfig: _router,
      debugShowCheckedModeBanner: false,
    );
  }
}
