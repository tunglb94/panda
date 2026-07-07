import 'dart:async';
import 'dart:math';

import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';
import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/routing/route_model.dart';
import 'package:rider/core/routing/route_service.dart';
import 'package:rider/features/booking/presentation/widgets/booking_bottom_sheet.dart';
import 'package:rider/features/map/data/driver_tracking_repository.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';

// ─── Location resolution state machine ───────────────────────────────────────

enum _LocationStatus {
  loading,
  permissionDenied,
  permissionPermanentlyDenied,
  gpsDisabled,
  ready,
}

// ─── Pickup / destination selection state machine ─────────────────────────────

enum _SelectionMode {
  pickupPending,
  destinationPending,
  confirmed,
}

// ─── Page ─────────────────────────────────────────────────────────────────────

class MapPage extends StatefulWidget {
  const MapPage({
    super.key,
    required this.apiClient,
    required this.routeService,
  });

  final ApiClient apiClient;
  final RouteService routeService;

  @override
  State<MapPage> createState() => MapPageState();
}

class MapPageState extends State<MapPage> {
  // — Location ——————————————————————————————————————————————————————————————————
  _LocationStatus _status = _LocationStatus.loading;
  GoogleMapController? _controller;
  LatLng? _position;
  static const double _defaultZoom = 15.0;

  // — Trip selection ————————————————————————————————————————————————————————————
  _SelectionMode _selectionMode = _SelectionMode.pickupPending;
  LatLng _cameraCenter = const LatLng(0, 0);
  LatLng? _pickupPoint;
  LatLng? _destinationPoint;
  Set<Marker> _markers = {};
  Set<Polyline> _polylines = {};
  RouteModel? _routeInfo;
  bool _routeLoading = false;

  // — Driver tracking (Phase 25) ————————————————————————————————————————————————
  late final DriverTrackingRepository _trackingRepo;
  String? _trackingDriverId;
  Timer? _trackingTimer;
  LatLng? _driverPosition;
  double _driverHeading = 0;

  // ─── Lifecycle ──────────────────────────────────────────────────────────────

  @override
  void initState() {
    super.initState();
    _trackingRepo = DriverTrackingRepository(apiClient: widget.apiClient);
    _resolveLocation();
  }

  @override
  void dispose() {
    _stopTracking();
    _controller?.dispose();
    super.dispose();
  }

  // ─── Driver tracking ──────────────────────────────────────────────────────────

  void startTracking(String driverID) {
    if (_trackingDriverId == driverID && _trackingTimer != null) return;
    _trackingDriverId = driverID;
    _trackingTimer?.cancel();
    _trackingTimer = Timer.periodic(
      const Duration(seconds: 5),
      (_) => _fetchDriverLocation(),
    );
    _fetchDriverLocation();
  }

  void _stopTracking() {
    _trackingTimer?.cancel();
    _trackingTimer = null;
    _trackingDriverId = null;
    if (mounted) {
      setState(() {
        _driverPosition = null;
        _rebuildMarkers();
      });
    }
  }

  Future<void> _fetchDriverLocation() async {
    final driverID = _trackingDriverId;
    if (driverID == null) return;
    try {
      final loc = await _trackingRepo.getDriverLocation(driverID);
      if (!mounted || _trackingDriverId != driverID) return;
      if (!loc.isActive) {
        _stopTracking();
        return;
      }
      final newPos = LatLng(loc.lat, loc.lon);
      setState(() {
        if (_driverPosition != null) {
          _driverHeading = _computeHeading(_driverPosition!, newPos);
        }
        _driverPosition = newPos;
        _rebuildMarkers();
      });
    } catch (_) {
      // Network failure — skip this tick, retry next
    }
  }

  static double _computeHeading(LatLng from, LatLng to) {
    final lat1 = from.latitude * pi / 180;
    final lat2 = to.latitude * pi / 180;
    final dLon = (to.longitude - from.longitude) * pi / 180;
    final y = sin(dLon) * cos(lat2);
    final x =
        cos(lat1) * sin(lat2) - sin(lat1) * cos(lat2) * cos(dLon);
    return (atan2(y, x) * 180 / pi + 360) % 360;
  }

  // ─── Location resolution ──────────────────────────────────────────────────────

  Future<void> _resolveLocation() async {
    if (!mounted) return;
    setState(() => _status = _LocationStatus.loading);

    final serviceEnabled = await Geolocator.isLocationServiceEnabled();
    if (!mounted) return;
    if (!serviceEnabled) {
      setState(() => _status = _LocationStatus.gpsDisabled);
      return;
    }

    LocationPermission permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
    }
    if (!mounted) return;

    if (permission == LocationPermission.denied) {
      setState(() => _status = _LocationStatus.permissionDenied);
      return;
    }
    if (permission == LocationPermission.deniedForever) {
      setState(() => _status = _LocationStatus.permissionPermanentlyDenied);
      return;
    }

    try {
      final pos = await Geolocator.getCurrentPosition(
        locationSettings:
            const LocationSettings(accuracy: LocationAccuracy.high),
      ).timeout(const Duration(seconds: 10));
      if (!mounted) return;
      final latLng = LatLng(pos.latitude, pos.longitude);
      setState(() {
        _position = latLng;
        _cameraCenter = latLng;
        _status = _LocationStatus.ready;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => _status = _LocationStatus.gpsDisabled);
    }
  }

  // ─── Map callbacks ────────────────────────────────────────────────────────────

  void _onMapCreated(GoogleMapController controller) {
    _controller = controller;
  }

  void _onCameraMove(CameraPosition pos) {
    if (_selectionMode == _SelectionMode.confirmed) return;
    setState(() => _cameraCenter = pos.target);
  }

  // ─── Selection actions ────────────────────────────────────────────────────────

  void _confirmPickup() {
    setState(() {
      _pickupPoint = _cameraCenter;
      _selectionMode = _destinationPoint != null
          ? _SelectionMode.confirmed
          : _SelectionMode.destinationPending;
      _rebuildMarkers();
    });
    if (_selectionMode == _SelectionMode.confirmed) {
      _fetchRoute();
    }
  }

  void _confirmDestination() {
    setState(() {
      _destinationPoint = _cameraCenter;
      _selectionMode = _SelectionMode.confirmed;
      _rebuildMarkers();
    });
    _fetchRoute();
  }

  void _editPickup() {
    final lastPickup = _pickupPoint;
    setState(() {
      _pickupPoint = null;
      _selectionMode = _SelectionMode.pickupPending;
      _polylines = {};
      _routeInfo = null;
      _rebuildMarkers();
    });
    if (lastPickup != null) {
      _controller?.animateCamera(CameraUpdate.newLatLng(lastPickup));
    }
  }

  void _editDestination() {
    final lastDestination = _destinationPoint;
    setState(() {
      _destinationPoint = null;
      _selectionMode = _SelectionMode.destinationPending;
      _polylines = {};
      _routeInfo = null;
      _rebuildMarkers();
    });
    if (lastDestination != null) {
      _controller?.animateCamera(CameraUpdate.newLatLng(lastDestination));
    }
  }

  Future<void> _fetchRoute() async {
    final pickup = _pickupPoint;
    final destination = _destinationPoint;
    if (pickup == null || destination == null) return;
    setState(() => _routeLoading = true);
    try {
      final route = await widget.routeService.getRoute(pickup, destination);
      if (!mounted) return;
      if (_pickupPoint != pickup || _destinationPoint != destination) return;
      setState(() {
        _routeInfo = route;
        _polylines = {
          Polyline(
            polylineId: const PolylineId('route'),
            points: route.polylinePoints,
            color: const Color(0xFF1565C0),
            width: 5,
          ),
        };
        _routeLoading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => _routeLoading = false);
    }
  }

  void _rebuildMarkers() {
    _markers = {
      if (_pickupPoint != null)
        Marker(
          markerId: const MarkerId('pickup'),
          position: _pickupPoint!,
          icon:
              BitmapDescriptor.defaultMarkerWithHue(BitmapDescriptor.hueGreen),
          infoWindow: const InfoWindow(title: 'Pickup'),
        ),
      if (_destinationPoint != null)
        Marker(
          markerId: const MarkerId('destination'),
          position: _destinationPoint!,
          icon: BitmapDescriptor.defaultMarkerWithHue(BitmapDescriptor.hueRed),
          infoWindow: const InfoWindow(title: 'Destination'),
        ),
      if (_driverPosition != null)
        Marker(
          markerId: const MarkerId('driver'),
          position: _driverPosition!,
          icon:
              BitmapDescriptor.defaultMarkerWithHue(BitmapDescriptor.hueAzure),
          flat: true,
          rotation: _driverHeading,
          infoWindow: const InfoWindow(title: 'Driver'),
          anchor: const Offset(0.5, 0.5),
        ),
    };
  }

  TripSelection? get _tripSelection {
    if (_pickupPoint == null || _destinationPoint == null) return null;
    return TripSelection(
      pickup: _pickupPoint!,
      destination: _destinationPoint!,
    );
  }

  // ─── Build ────────────────────────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: switch (_status) {
        _LocationStatus.loading => const _LocationLoadingView(),
        _LocationStatus.permissionDenied => _LocationErrorView(
            icon: Icons.location_off,
            title: 'Location permission denied',
            message: 'FAIRRIDE needs your location to show nearby drivers and '
                'estimate fares.',
            actionLabel: 'Grant permission',
            onAction: _resolveLocation,
          ),
        _LocationStatus.permissionPermanentlyDenied => _LocationErrorView(
            icon: Icons.location_disabled,
            title: 'Location access blocked',
            message: 'Please enable location permission for FAIRRIDE in your '
                'device Settings.',
            actionLabel: 'Open Settings',
            onAction: () async => Geolocator.openAppSettings(),
          ),
        _LocationStatus.gpsDisabled => _LocationErrorView(
            icon: Icons.gps_off,
            title: 'GPS is turned off',
            message:
                'Turn on Location Services so FAIRRIDE can show you on the map.',
            actionLabel: 'Open Location Settings',
            onAction: () async => Geolocator.openLocationSettings(),
          ),
        _LocationStatus.ready => _buildSelectionMap(),
      },
    );
  }

  Widget _buildSelectionMap() {
    final showPin = _selectionMode != _SelectionMode.confirmed;
    return Stack(
      children: [
        GoogleMap(
          initialCameraPosition:
              CameraPosition(target: _position!, zoom: _defaultZoom),
          onMapCreated: _onMapCreated,
          onCameraMove: _onCameraMove,
          markers: _markers,
          polylines: _polylines,
          myLocationEnabled: true,
          myLocationButtonEnabled: true,
          zoomControlsEnabled: true,
          compassEnabled: true,
          mapToolbarEnabled: false,
          mapType: MapType.normal,
          padding: const EdgeInsets.only(bottom: 240),
        ),
        if (showPin) const _CenterPin(),
        Positioned(
          left: 0,
          right: 0,
          bottom: 0,
          child: _SelectionPanel(
            mode: _selectionMode,
            cameraCenter: _cameraCenter,
            pickupPoint: _pickupPoint,
            destinationPoint: _destinationPoint,
            onConfirmPickup: _confirmPickup,
            onConfirmDestination: _confirmDestination,
            onEditPickup: _editPickup,
            onEditDestination: _editDestination,
            onBookRide: _tripSelection != null
                ? () => BookingBottomSheet.show(context, tripSelection: _tripSelection!)
                : null,
            routeDistanceText: _routeInfo?.distanceText,
            routeDurationText: _routeInfo?.durationText,
            routeLoading: _routeLoading,
          ),
        ),
      ],
    );
  }
}

// ─── Center pin overlay ───────────────────────────────────────────────────────

class _CenterPin extends StatelessWidget {
  const _CenterPin();

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.only(bottom: 48),
        child: Icon(
          Icons.location_pin,
          size: 48,
          color: Theme.of(context).colorScheme.primary,
          shadows: const [Shadow(blurRadius: 6, color: Colors.black26)],
        ),
      ),
    );
  }
}

// ─── Selection panel ──────────────────────────────────────────────────────────

class _SelectionPanel extends StatelessWidget {
  const _SelectionPanel({
    required this.mode,
    required this.cameraCenter,
    required this.pickupPoint,
    required this.destinationPoint,
    required this.onConfirmPickup,
    required this.onConfirmDestination,
    required this.onEditPickup,
    required this.onEditDestination,
    this.onBookRide,
    this.routeDistanceText,
    this.routeDurationText,
    this.routeLoading = false,
  });

  final _SelectionMode mode;
  final LatLng cameraCenter;
  final LatLng? pickupPoint;
  final LatLng? destinationPoint;
  final VoidCallback onConfirmPickup;
  final VoidCallback onConfirmDestination;
  final VoidCallback onEditPickup;
  final VoidCallback onEditDestination;
  final VoidCallback? onBookRide;
  final String? routeDistanceText;
  final String? routeDurationText;
  final bool routeLoading;

  @override
  Widget build(BuildContext context) {
    return Material(
      elevation: 12,
      borderRadius: const BorderRadius.vertical(top: Radius.circular(20)),
      child: SafeArea(
        top: false,
        child: Padding(
          padding: const EdgeInsets.fromLTRB(20, 20, 20, 16),
          child: switch (mode) {
            _SelectionMode.pickupPending => _buildPickupPending(context),
            _SelectionMode.destinationPending =>
              _buildDestinationPending(context),
            _SelectionMode.confirmed => _buildConfirmed(context),
          },
        ),
      ),
    );
  }

  Widget _buildPickupPending(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        _PointRow(
          icon: Icons.my_location,
          iconColor: primary,
          label: 'Set pickup',
          subtitle: 'Drag the map to adjust your pickup location',
          coordinate: cameraCenter,
          active: true,
        ),
        const SizedBox(height: 8),
        _PointRow(
          icon: Icons.flag_outlined,
          iconColor: Colors.red,
          label: 'Destination',
          subtitle: 'Confirm pickup first',
          coordinate: null,
          active: false,
        ),
        const SizedBox(height: 20),
        FilledButton(
          onPressed: onConfirmPickup,
          child: const Text('Confirm Pickup'),
        ),
      ],
    );
  }

  Widget _buildDestinationPending(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        _PointRow(
          icon: Icons.my_location,
          iconColor: primary,
          label: 'Pickup',
          coordinate: pickupPoint,
          active: false,
          trailing: TextButton(
            onPressed: onEditPickup,
            child: const Text('Edit'),
          ),
        ),
        const Divider(height: 20),
        _PointRow(
          icon: Icons.flag,
          iconColor: Colors.red,
          label: 'Set destination',
          subtitle: 'Drag the map to set your destination',
          coordinate: cameraCenter,
          active: true,
        ),
        const SizedBox(height: 20),
        FilledButton(
          onPressed: onConfirmDestination,
          child: const Text('Confirm Destination'),
        ),
      ],
    );
  }

  Widget _buildConfirmed(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    final textStyle = Theme.of(context)
        .textTheme
        .bodySmall
        ?.copyWith(color: Colors.grey.shade700);
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        _PointRow(
          icon: Icons.my_location,
          iconColor: primary,
          label: 'Pickup',
          coordinate: pickupPoint,
          active: false,
          trailing: TextButton(
            onPressed: onEditPickup,
            child: const Text('Edit'),
          ),
        ),
        const Divider(height: 20),
        _PointRow(
          icon: Icons.flag,
          iconColor: Colors.red,
          label: 'Destination',
          coordinate: destinationPoint,
          active: false,
          trailing: TextButton(
            onPressed: onEditDestination,
            child: const Text('Edit'),
          ),
        ),
        if (routeLoading) ...[
          const SizedBox(height: 12),
          const Center(
            child: SizedBox(
              width: 20,
              height: 20,
              child: CircularProgressIndicator(strokeWidth: 2),
            ),
          ),
        ] else if (routeDistanceText != null && routeDurationText != null) ...[
          const Divider(height: 20),
          Row(
            children: [
              Icon(Icons.route, size: 16, color: Colors.grey.shade600),
              const SizedBox(width: 6),
              Text(routeDistanceText!, style: textStyle),
              const SizedBox(width: 16),
              Icon(Icons.access_time, size: 16, color: Colors.grey.shade600),
              const SizedBox(width: 6),
              Text(routeDurationText!, style: textStyle),
            ],
          ),
        ],
        const SizedBox(height: 16),
        FilledButton.icon(
          onPressed: onBookRide,
          icon: const Icon(Icons.local_taxi),
          label: const Text('Book this ride'),
        ),
      ],
    );
  }
}

// ─── Point row ────────────────────────────────────────────────────────────────

class _PointRow extends StatelessWidget {
  const _PointRow({
    required this.icon,
    required this.iconColor,
    required this.label,
    required this.coordinate,
    required this.active,
    this.subtitle,
    this.trailing,
  });

  final IconData icon;
  final Color iconColor;
  final String label;
  final String? subtitle;
  final LatLng? coordinate;
  final bool active;
  final Widget? trailing;

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    final primary = Theme.of(context).colorScheme.primary;

    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.only(top: 2),
          child: Icon(icon, color: iconColor, size: 22),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: textTheme.titleSmall?.copyWith(
                  fontWeight: active ? FontWeight.bold : FontWeight.w500,
                ),
              ),
              if (subtitle != null) ...[
                const SizedBox(height: 2),
                Text(
                  subtitle!,
                  style: textTheme.bodySmall
                      ?.copyWith(color: Colors.grey.shade600),
                ),
              ],
              if (coordinate != null) ...[
                const SizedBox(height: 2),
                Text(
                  _formatCoord(coordinate!),
                  style: textTheme.bodySmall?.copyWith(
                    color: active ? primary : Colors.grey.shade700,
                    fontFamily: 'monospace',
                  ),
                ),
              ],
            ],
          ),
        ),
        if (trailing != null) trailing!,
      ],
    );
  }

  static String _formatCoord(LatLng p) =>
      '${p.latitude.toStringAsFixed(5)}, ${p.longitude.toStringAsFixed(5)}';
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
