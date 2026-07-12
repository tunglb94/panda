import 'dart:async';
import 'dart:math';

import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';
import 'package:rider/core/location/location_engine.dart';
import 'package:rider/core/location/location_engine_config.dart';
import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/places/nominatim_places_service.dart';
import 'package:rider/core/routing/route_engine.dart';
import 'package:rider/core/routing/route_model.dart';
import 'package:rider/core/routing/route_point.dart';
import 'package:rider/core/routing/route_progress.dart';
import 'package:rider/core/routing/route_progress_engine.dart';
import 'package:rider/core/routing/route_provider.dart';
import 'package:rider/features/booking/presentation/widgets/booking_bottom_sheet.dart';
import 'package:rider/features/delivery/presentation/pages/delivery_form_page.dart';
import 'package:rider/features/map/data/driver_tracking_repository.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';
import 'package:rider/features/map/presentation/widgets/place_search_field.dart';
import 'package:rider/shared/widgets/mascot_image.dart';

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
    required this.routeProvider,
  });

  final ApiClient apiClient;
  final RouteProvider routeProvider;

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
  String? _pickupAddress;
  String? _destinationAddress;
  bool _isAnimatingCamera = false;
  Set<Marker> _markers = {};
  Set<Polyline> _polylines = {};
  RouteModel? _routeInfo;
  bool _routeLoading = false;

  // — Place search (Phase R-05) ————————————————————————————————————————————————
  late final NominatimPlacesService _placesService;

  // — Route engine (Phase 32) ──────────────────────────────────────────────────
  late final RouteEngine _routeEngine;

  // — Route progress (Phase 33) ————————————————————————————————————————————————
  late final LocationEngine _locationEngine;
  late final RouteProgressEngine _progressEngine;
  StreamSubscription<RouteProgress>? _progressSub;
  RouteProgress? _routeProgress;

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
    _routeEngine = RouteEngine(provider: widget.routeProvider);
    _locationEngine = LocationEngine(
      config: const LocationEngineConfig(distanceFilter: 5.0),
    );
    _progressEngine = RouteProgressEngine(
      locationStream: _locationEngine.locationStream,
      routeEngine: _routeEngine,
    );
    _trackingRepo = DriverTrackingRepository(apiClient: widget.apiClient);
    _placesService = const NominatimPlacesService();
    _resolveLocation();
  }

  @override
  void dispose() {
    _progressSub?.cancel();
    _progressEngine.dispose();
    _routeEngine.dispose();
    _locationEngine.dispose();
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
    setState(() {
      _cameraCenter = pos.target;
      // A manual drag invalidates whichever address came from search — but
      // ignore moves caused by our own animateCamera() (see _isAnimatingCamera).
      if (!_isAnimatingCamera) {
        if (_selectionMode == _SelectionMode.pickupPending) {
          _pickupAddress = null;
        } else if (_selectionMode == _SelectionMode.destinationPending) {
          _destinationAddress = null;
        }
      }
    });
  }

  void _onCameraIdle() {
    _isAnimatingCamera = false;
  }

  // ─── Place search ─────────────────────────────────────────────────────────────

  void _onPickupPlaceSelected(String address, LatLng location) {
    _isAnimatingCamera = true;
    setState(() {
      _cameraCenter = location;
      _pickupAddress = address;
    });
    _controller?.animateCamera(CameraUpdate.newLatLngZoom(location, _defaultZoom));
  }

  void _onDestinationPlaceSelected(String address, LatLng location) {
    _isAnimatingCamera = true;
    setState(() {
      _cameraCenter = location;
      _destinationAddress = address;
    });
    _controller?.animateCamera(CameraUpdate.newLatLngZoom(location, _defaultZoom));
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
    _stopProgressTracking();
    _routeEngine.clear();
    setState(() {
      _pickupPoint = null;
      _selectionMode = _SelectionMode.pickupPending;
      _polylines = {};
      _routeInfo = null;
      _routeProgress = null;
      _rebuildMarkers();
    });
    if (lastPickup != null) {
      _controller?.animateCamera(CameraUpdate.newLatLng(lastPickup));
    }
  }

  void _editDestination() {
    final lastDestination = _destinationPoint;
    _stopProgressTracking();
    _routeEngine.clear();
    setState(() {
      _destinationPoint = null;
      _selectionMode = _SelectionMode.destinationPending;
      _polylines = {};
      _routeInfo = null;
      _routeProgress = null;
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
      final route = await _routeEngine.loadRoute(
        RoutePoint(latitude: pickup.latitude, longitude: pickup.longitude),
        RoutePoint(latitude: destination.latitude, longitude: destination.longitude),
      );
      if (!mounted) return;
      if (_pickupPoint != pickup || _destinationPoint != destination) return;
      final polylinePoints = route.decodedPolyline
          .map((p) => LatLng(p.latitude, p.longitude))
          .toList();
      setState(() {
        _routeInfo = route;
        _routeProgress = null;
        _polylines = {
          Polyline(
            polylineId: const PolylineId('route'),
            points: polylinePoints,
            color: const Color(0xFF1565C0),
            width: 5,
          ),
        };
        _routeLoading = false;
      });
      _startProgressTracking();
    } catch (_) {
      if (!mounted) return;
      setState(() => _routeLoading = false);
    }
  }

  void _startProgressTracking() {
    _stopProgressTracking();
    if (_locationEngine.state == LocationEngineState.stopped) {
      unawaited(_locationEngine.start());
    }
    _progressSub = _progressEngine.progressStream.listen((p) {
      if (!mounted) return;
      setState(() => _routeProgress = p);
    });
    _progressEngine.start();
  }

  void _stopProgressTracking() {
    _progressSub?.cancel();
    _progressSub = null;
    _progressEngine.stop();
    _locationEngine.stop();
  }

  void _rebuildMarkers() {
    _markers = {
      if (_pickupPoint != null)
        Marker(
          markerId: const MarkerId('pickup'),
          position: _pickupPoint!,
          icon:
              BitmapDescriptor.defaultMarkerWithHue(BitmapDescriptor.hueGreen),
          infoWindow: const InfoWindow(title: 'Điểm đón'),
        ),
      if (_destinationPoint != null)
        Marker(
          markerId: const MarkerId('destination'),
          position: _destinationPoint!,
          icon: BitmapDescriptor.defaultMarkerWithHue(BitmapDescriptor.hueRed),
          infoWindow: const InfoWindow(title: 'Điểm đến'),
        ),
      if (_driverPosition != null)
        Marker(
          markerId: const MarkerId('driver'),
          position: _driverPosition!,
          icon:
              BitmapDescriptor.defaultMarkerWithHue(BitmapDescriptor.hueAzure),
          flat: true,
          rotation: _driverHeading,
          infoWindow: const InfoWindow(title: 'Tài xế'),
          anchor: const Offset(0.5, 0.5),
        ),
    };
  }

  TripSelection? get _tripSelection {
    if (_pickupPoint == null || _destinationPoint == null) return null;
    return TripSelection(
      pickup: _pickupPoint!,
      destination: _destinationPoint!,
      pickupAddress: _pickupAddress,
      destinationAddress: _destinationAddress,
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
            title: 'Không có quyền truy cập vị trí',
            message: 'Panda cần vị trí của bạn để hiển thị tài xế gần đó và '
                'ước tính cước phí.',
            actionLabel: 'Cấp quyền',
            onAction: _resolveLocation,
          ),
        _LocationStatus.permissionPermanentlyDenied => _LocationErrorView(
            icon: Icons.location_disabled,
            title: 'Quyền truy cập vị trí bị chặn',
            message: 'Vui lòng bật quyền truy cập vị trí cho Panda trong '
                'Cài đặt thiết bị.',
            actionLabel: 'Mở Cài đặt',
            onAction: () async => Geolocator.openAppSettings(),
          ),
        _LocationStatus.gpsDisabled => _LocationErrorView(
            icon: Icons.gps_off,
            title: 'GPS đang tắt',
            message:
                'Bật Dịch vụ định vị để Panda có thể hiển thị vị trí của bạn trên bản đồ.',
            actionLabel: 'Mở Cài đặt định vị',
            onAction: () async => Geolocator.openLocationSettings(),
            mascotAsset: 'mascot_no_gps.png',
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
          onCameraIdle: _onCameraIdle,
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
        // Delivery entry point — only shown before the rider starts picking
        // a Ride pickup point, so it never overlaps or interferes with the
        // Ride pickup/destination selection flow below.
        if (_selectionMode == _SelectionMode.pickupPending)
          Positioned(
            top: MediaQuery.of(context).padding.top + 12,
            right: 12,
            child: _DeliveryEntryButton(
              onTap: () => Navigator.of(context).push(MaterialPageRoute(
                builder: (_) => DeliveryFormPage(apiClient: widget.apiClient, initialBias: _cameraCenter),
              )),
            ),
          ),
        Positioned(
          left: 0,
          right: 0,
          bottom: 0,
          child: _SelectionPanel(
            mode: _selectionMode,
            cameraCenter: _cameraCenter,
            pickupPoint: _pickupPoint,
            destinationPoint: _destinationPoint,
            pickupAddress: _pickupAddress,
            destinationAddress: _destinationAddress,
            placesService: _placesService,
            onPickupPlaceSelected: _onPickupPlaceSelected,
            onDestinationPlaceSelected: _onDestinationPlaceSelected,
            onConfirmPickup: _confirmPickup,
            onConfirmDestination: _confirmDestination,
            onEditPickup: _editPickup,
            onEditDestination: _editDestination,
            onBookRide: _tripSelection != null
                ? () => BookingBottomSheet.show(context, tripSelection: _tripSelection!, apiClient: widget.apiClient, onDriverAssigned: startTracking)
                : null,
            routeDistanceText: _routeInfo?.distanceText,
            routeDurationText: _routeInfo?.durationText,
            routeLoading: _routeLoading,
            routeProgress: _routeProgress,
          ),
        ),
      ],
    );
  }
}

// ─── Delivery entry point ───────────────────────────────────────────────────

class _DeliveryEntryButton extends StatelessWidget {
  const _DeliveryEntryButton({required this.onTap});

  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.white,
      elevation: 3,
      borderRadius: BorderRadius.circular(24),
      child: InkWell(
        borderRadius: BorderRadius.circular(24),
        onTap: onTap,
        child: const Padding(
          padding: EdgeInsets.symmetric(horizontal: 14, vertical: 10),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.local_shipping_outlined, size: 18, color: Colors.black87),
              SizedBox(width: 6),
              Text('Gửi hàng', style: TextStyle(fontWeight: FontWeight.w600, color: Colors.black87)),
            ],
          ),
        ),
      ),
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
    required this.placesService,
    required this.onPickupPlaceSelected,
    required this.onDestinationPlaceSelected,
    required this.onConfirmPickup,
    required this.onConfirmDestination,
    required this.onEditPickup,
    required this.onEditDestination,
    this.pickupAddress,
    this.destinationAddress,
    this.onBookRide,
    this.routeDistanceText,
    this.routeDurationText,
    this.routeLoading = false,
    this.routeProgress,
  });

  final _SelectionMode mode;
  final LatLng cameraCenter;
  final LatLng? pickupPoint;
  final LatLng? destinationPoint;
  final String? pickupAddress;
  final String? destinationAddress;
  final NominatimPlacesService placesService;
  final void Function(String address, LatLng location) onPickupPlaceSelected;
  final void Function(String address, LatLng location) onDestinationPlaceSelected;
  final VoidCallback onConfirmPickup;
  final VoidCallback onConfirmDestination;
  final VoidCallback onEditPickup;
  final VoidCallback onEditDestination;
  final VoidCallback? onBookRide;
  final String? routeDistanceText;
  final String? routeDurationText;
  final bool routeLoading;
  final RouteProgress? routeProgress;

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
        PlaceSearchField(
          key: const ValueKey('pickup-search'),
          placesService: placesService,
          hintText: 'Tìm điểm đón (VD: Chợ Bến Thành)',
          biasCenter: cameraCenter,
          onSelected: onPickupPlaceSelected,
        ),
        const SizedBox(height: 12),
        _PointRow(
          icon: Icons.my_location,
          iconColor: primary,
          label: 'Điểm đón',
          subtitle: 'Hoặc kéo bản đồ để điều chỉnh',
          coordinate: cameraCenter,
          addressText: pickupAddress,
          active: true,
        ),
        const SizedBox(height: 8),
        _PointRow(
          icon: Icons.flag_outlined,
          iconColor: Colors.red,
          label: 'Điểm đến',
          subtitle: 'Xác nhận điểm đón trước',
          coordinate: null,
          active: false,
        ),
        const SizedBox(height: 20),
        FilledButton(
          onPressed: onConfirmPickup,
          child: const Text('Xác nhận điểm đón'),
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
          label: 'Điểm đón',
          coordinate: pickupPoint,
          addressText: pickupAddress,
          active: false,
          trailing: TextButton(
            onPressed: onEditPickup,
            child: const Text('Sửa'),
          ),
        ),
        const Divider(height: 20),
        PlaceSearchField(
          key: const ValueKey('destination-search'),
          placesService: placesService,
          hintText: 'Tìm điểm đến (VD: Sân bay Tân Sơn Nhất)',
          biasCenter: cameraCenter,
          onSelected: onDestinationPlaceSelected,
        ),
        const SizedBox(height: 12),
        _PointRow(
          icon: Icons.flag,
          iconColor: Colors.red,
          label: 'Chọn điểm đến',
          subtitle: 'Hoặc kéo bản đồ để điều chỉnh',
          coordinate: cameraCenter,
          addressText: destinationAddress,
          active: true,
        ),
        const SizedBox(height: 20),
        FilledButton(
          onPressed: onConfirmDestination,
          child: const Text('Xác nhận điểm đến'),
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
          label: 'Điểm đón',
          coordinate: pickupPoint,
          addressText: pickupAddress,
          active: false,
          trailing: TextButton(
            onPressed: onEditPickup,
            child: const Text('Sửa'),
          ),
        ),
        const Divider(height: 20),
        _PointRow(
          icon: Icons.flag,
          iconColor: Colors.red,
          label: 'Điểm đến',
          coordinate: destinationPoint,
          addressText: destinationAddress,
          active: false,
          trailing: TextButton(
            onPressed: onEditDestination,
            child: const Text('Sửa'),
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
        ] else if (routeProgress != null) ...[
          const Divider(height: 20),
          _RouteProgressBar(progress: routeProgress!),
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
          label: const Text('Đặt xe ngay'),
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
    this.addressText,
    this.trailing,
  });

  final IconData icon;
  final Color iconColor;
  final String label;
  final String? subtitle;
  final LatLng? coordinate;
  final String? addressText;
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
              if (addressText != null) ...[
                const SizedBox(height: 2),
                Text(
                  addressText!,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: textTheme.bodySmall?.copyWith(
                    color: active ? primary : Colors.grey.shade700,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ] else if (coordinate != null) ...[
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

// ─── Route progress bar ───────────────────────────────────────────────────────

class _RouteProgressBar extends StatelessWidget {
  const _RouteProgressBar({required this.progress});

  final RouteProgress progress;

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    final textStyle = Theme.of(context)
        .textTheme
        .bodySmall
        ?.copyWith(color: Colors.grey.shade700);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        ClipRRect(
          borderRadius: BorderRadius.circular(4),
          child: LinearProgressIndicator(
            value: progress.progressPercent,
            minHeight: 6,
            backgroundColor: Colors.grey.shade200,
            color: progress.isOnRoute ? primary : Colors.orange,
          ),
        ),
        const SizedBox(height: 8),
        Row(
          children: [
            Icon(Icons.route, size: 16, color: Colors.grey.shade600),
            const SizedBox(width: 6),
            Text(_formatDistance(progress.remainingMeters), style: textStyle),
            const SizedBox(width: 16),
            Icon(Icons.access_time, size: 16, color: Colors.grey.shade600),
            const SizedBox(width: 6),
            Text(_formatDuration(progress.remainingDurationSeconds), style: textStyle),
            const Spacer(),
            if (!progress.isOnRoute)
              Text(
                'Lệch lộ trình',
                style: textStyle?.copyWith(color: Colors.orange),
              ),
          ],
        ),
      ],
    );
  }

  static String _formatDistance(int meters) {
    if (meters < 1000) return '${meters}m';
    return '${(meters / 1000).toStringAsFixed(1)}km';
  }

  static String _formatDuration(int seconds) {
    if (seconds < 60) return '< 1 phút';
    final mins = (seconds / 60).round();
    if (mins < 60) return '$mins phút';
    final hours = mins ~/ 60;
    final rem = mins % 60;
    return rem == 0 ? '${hours}h' : '${hours}h ${rem}phút';
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
              'Đang tìm vị trí của bạn…',
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
    this.mascotAsset,
  });

  final IconData icon;
  final String title;
  final String message;
  final String actionLabel;
  final VoidCallback onAction;

  /// Optional Panda mascot (file name under `assets/mascot/`) shown instead
  /// of the plain icon for the one state where a mascot genuinely matches
  /// the content (GPS disabled) — see `docs/design/MASCOT_CATALOG.md`.
  /// Left unset for the permission-related states to avoid overusing it.
  final String? mascotAsset;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            if (mascotAsset != null)
              MascotImage(asset: mascotAsset!, size: MascotSize.large, animation: MascotAnimation.scale)
            else
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
