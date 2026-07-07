import 'dart:async';

import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';
import '../../../../core/auth/auth_state.dart';
import '../../../../core/network/api_client.dart';
import '../../../../core/storage/token_storage.dart';
import '../../../home/data/availability_repository.dart';
import '../../../location/data/location_upload_repository.dart';
import '../../../location/services/location_upload_service.dart';

// ─── Location state machine ───────────────────────────────────────────────────

enum _LocationStatus {
  loading,
  permissionDenied,
  permissionPermanentlyDenied,
  gpsDisabled,
  ready,
}

// ─── Page ─────────────────────────────────────────────────────────────────────

class MapPage extends StatefulWidget {
  const MapPage({
    super.key,
    required this.authState,
    required this.tokenStorage,
    required this.apiClient,
  });

  final AuthState authState;
  final TokenStorage tokenStorage;
  final ApiClient apiClient;

  @override
  State<MapPage> createState() => _MapPageState();
}

class _MapPageState extends State<MapPage> {
  // — Location ——————————————————————————————————————————————————————————————————
  _LocationStatus _locationStatus = _LocationStatus.loading;
  LatLng? _position;
  GoogleMapController? _mapController;
  static const double _defaultZoom = 15.0;

  // — Availability ——————————————————————————————————————————————————————————————
  late final AvailabilityRepository _availRepo;
  bool _isOnline = false;
  bool _isLoadingStatus = true;
  bool _isToggling = false;
  String? _availError;

  // — Location upload ——————————————————————————————————————————————————————————
  late final LocationUploadService _uploadService;
  UploadStatus _uploadStatus = UploadStatus.idle;
  StreamSubscription<UploadStatus>? _uploadStatusSub;

  // ─── Lifecycle ──────────────────────────────────────────────────────────────

  @override
  void initState() {
    super.initState();
    _availRepo = AvailabilityRepository(widget.apiClient);
    _uploadService = LocationUploadService(
      repository: LocationUploadRepository(apiClient: widget.apiClient),
    );
    _uploadStatusSub = _uploadService.statusStream.listen((s) {
      if (mounted) setState(() => _uploadStatus = s);
    });
    _resolveLocation();
    _fetchAvailability();
  }

  @override
  void dispose() {
    _uploadStatusSub?.cancel();
    _uploadService.dispose();
    _mapController?.dispose();
    super.dispose();
  }

  // ─── Location ───────────────────────────────────────────────────────────────

  Future<void> _resolveLocation() async {
    if (!mounted) return;
    setState(() => _locationStatus = _LocationStatus.loading);

    final serviceEnabled = await Geolocator.isLocationServiceEnabled();
    if (!mounted) return;
    if (!serviceEnabled) {
      setState(() => _locationStatus = _LocationStatus.gpsDisabled);
      return;
    }

    LocationPermission permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
    }
    if (!mounted) return;

    if (permission == LocationPermission.denied) {
      setState(() => _locationStatus = _LocationStatus.permissionDenied);
      return;
    }
    if (permission == LocationPermission.deniedForever) {
      setState(() =>
          _locationStatus = _LocationStatus.permissionPermanentlyDenied);
      return;
    }

    try {
      final pos = await Geolocator.getCurrentPosition(
        locationSettings:
            const LocationSettings(accuracy: LocationAccuracy.high),
      ).timeout(const Duration(seconds: 10));
      if (!mounted) return;
      setState(() {
        _position = LatLng(pos.latitude, pos.longitude);
        _locationStatus = _LocationStatus.ready;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => _locationStatus = _LocationStatus.gpsDisabled);
    }
  }

  // ─── Availability ────────────────────────────────────────────────────────────

  Future<void> _fetchAvailability() async {
    try {
      final result = await _availRepo.getAvailability();
      if (mounted) setState(() => _isOnline = result.isOnline);
      if (result.isOnline) unawaited(_uploadService.start());
    } catch (_) {
      // Non-fatal — default to offline
    } finally {
      if (mounted) setState(() => _isLoadingStatus = false);
    }
  }

  Future<void> _toggle() async {
    if (_isToggling) return;
    setState(() {
      _isToggling = true;
      _availError = null;
    });
    try {
      final result = _isOnline
          ? await _availRepo.goOffline()
          : await _availRepo.goOnline();
      if (!mounted) return;
      setState(() => _isOnline = result.isOnline);
      if (result.isOnline) {
        unawaited(_uploadService.start());
      } else {
        _uploadService.stop();
      }
    } on ApiException catch (e) {
      if (mounted) setState(() => _availError = e.message);
    } catch (_) {
      if (mounted) {
        setState(
            () => _availError = 'Could not update status. Please try again.');
      }
    } finally {
      if (mounted) setState(() => _isToggling = false);
    }
  }

  // ─── Build ────────────────────────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: switch (_locationStatus) {
        _LocationStatus.loading => const _LocationLoadingView(),
        _LocationStatus.permissionDenied => _LocationErrorView(
            icon: Icons.location_off,
            title: 'Location permission denied',
            message:
                'FAIRRIDE Driver needs your location to show your position '
                'on the map.',
            actionLabel: 'Grant permission',
            onAction: _resolveLocation,
          ),
        _LocationStatus.permissionPermanentlyDenied => _LocationErrorView(
            icon: Icons.location_disabled,
            title: 'Location access blocked',
            message:
                'Please enable location permission for FAIRRIDE Driver in '
                'your device Settings.',
            actionLabel: 'Open Settings',
            onAction: () async => Geolocator.openAppSettings(),
          ),
        _LocationStatus.gpsDisabled => _LocationErrorView(
            icon: Icons.gps_off,
            title: 'GPS is turned off',
            message:
                'Turn on Location Services so FAIRRIDE Driver can show your '
                'position on the map.',
            actionLabel: 'Open Location Settings',
            onAction: () async => Geolocator.openLocationSettings(),
          ),
        _LocationStatus.ready => _buildMap(),
      },
    );
  }

  Widget _buildMap() {
    return Stack(
      children: [
        GoogleMap(
          initialCameraPosition: CameraPosition(
            target: _position!,
            zoom: _defaultZoom,
          ),
          onMapCreated: (controller) => _mapController = controller,
          myLocationEnabled: true,
          myLocationButtonEnabled: true,
          zoomControlsEnabled: true,
          compassEnabled: true,
          mapToolbarEnabled: false,
          mapType: MapType.normal,
          // Reserve space above the status card so map controls stay visible.
          padding: const EdgeInsets.only(bottom: 116),
        ),
        Positioned(
          left: 0,
          right: 0,
          bottom: 0,
          child: _StatusCard(
            isOnline: _isOnline,
            isLoading: _isLoadingStatus || _isToggling,
            error: _availError,
            uploadStatus: _isOnline ? _uploadStatus : UploadStatus.idle,
            onToggle: _toggle,
          ),
        ),
      ],
    );
  }
}

// ─── Status card overlay ──────────────────────────────────────────────────────

class _StatusCard extends StatelessWidget {
  const _StatusCard({
    required this.isOnline,
    required this.isLoading,
    this.error,
    required this.uploadStatus,
    required this.onToggle,
  });

  final bool isOnline;
  final bool isLoading;
  final String? error;
  final UploadStatus uploadStatus;
  final VoidCallback onToggle;

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Material(
      elevation: 8,
      borderRadius: const BorderRadius.vertical(top: Radius.circular(20)),
      child: SafeArea(
        top: false,
        child: Padding(
          padding: const EdgeInsets.fromLTRB(20, 16, 20, 12),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Row(
                children: [
                  if (isLoading)
                    const SizedBox(
                      width: 12,
                      height: 12,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  else
                    Container(
                      width: 12,
                      height: 12,
                      decoration: BoxDecoration(
                        color: isOnline
                            ? const Color(0xFF1A8C4E)
                            : cs.outlineVariant,
                        shape: BoxShape.circle,
                      ),
                    ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      isOnline ? 'You are online' : 'You are offline',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                  ),
                  if (isOnline && uploadStatus != UploadStatus.idle) ...[
                    _UploadIndicator(status: uploadStatus),
                    const SizedBox(width: 8),
                  ],
                  Switch(
                    value: isOnline,
                    onChanged: isLoading ? null : (_) => onToggle(),
                  ),
                ],
              ),
              if (error != null) ...[
                const SizedBox(height: 6),
                Text(
                  error!,
                  style: TextStyle(color: cs.error, fontSize: 12),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

// ─── Upload status indicator ─────────────────────────────────────────────────

class _UploadIndicator extends StatelessWidget {
  const _UploadIndicator({required this.status});

  final UploadStatus status;

  @override
  Widget build(BuildContext context) {
    final (color, icon) = switch (status) {
      UploadStatus.uploading => (Colors.orange, Icons.cloud_upload_outlined),
      UploadStatus.success => (Colors.green, Icons.cloud_done_outlined),
      UploadStatus.failed => (Colors.red, Icons.cloud_off_outlined),
      UploadStatus.idle => (Colors.grey, Icons.cloud_outlined),
    };
    return Icon(icon, size: 16, color: color);
  }
}

// ─── Loading view ─────────────────────────────────────────────────────────────

class _LocationLoadingView extends StatelessWidget {
  const _LocationLoadingView();

  @override
  Widget build(BuildContext context) {
    return const ColoredBox(
      color: Colors.white,
      child: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            CircularProgressIndicator(),
            SizedBox(height: 20),
            Text(
              'Finding your location…',
              style: TextStyle(fontSize: 15, color: Colors.black54),
            ),
          ],
        ),
      ),
    );
  }
}

// ─── Error / blocked view ─────────────────────────────────────────────────────

class _LocationErrorView extends StatelessWidget {
  const _LocationErrorView({
    required this.icon,
    required this.title,
    required this.message,
    required this.actionLabel,
    required this.onAction,
  });

  final IconData icon;
  final String title;
  final String message;
  final String actionLabel;
  final VoidCallback onAction;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(icon, size: 72, color: theme.colorScheme.primary),
            const SizedBox(height: 24),
            Text(
              title,
              style: theme.textTheme.titleLarge
                  ?.copyWith(fontWeight: FontWeight.bold),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 12),
            Text(
              message,
              style: theme.textTheme.bodyMedium
                  ?.copyWith(color: Colors.grey.shade600),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 36),
            ElevatedButton(
              onPressed: onAction,
              child: Text(actionLabel),
            ),
          ],
        ),
      ),
    );
  }
}
