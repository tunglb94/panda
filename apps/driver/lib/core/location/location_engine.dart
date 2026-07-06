import 'dart:async';
import 'dart:io' show Platform;

import 'package:geolocator/geolocator.dart';
import 'location_engine_config.dart';

// ─── Public value types ───────────────────────────────────────────────────────

class LocationUpdate {
  const LocationUpdate({
    required this.latitude,
    required this.longitude,
    required this.accuracyMeters,
    required this.timestamp,
    required this.altitude,
    required this.speed,
    required this.heading,
  });

  final double latitude;
  final double longitude;
  final double accuracyMeters;
  final DateTime timestamp;
  final double altitude;
  final double speed;
  final double heading;
}

// ─── Status enums ─────────────────────────────────────────────────────────────

enum GpsStatus { enabled, disabled }

enum LocationPermissionStatus { granted, denied, permanentlyDenied }

enum LocationEngineState { stopped, running, paused }

// ─── Engine ───────────────────────────────────────────────────────────────────

/// Reusable, stream-based GPS engine.
///
/// Exposes three broadcast streams:
/// - [locationStream]  — continuous [LocationUpdate] events
/// - [gpsStatusStream] — [GpsStatus] changes (GPS on/off)
/// - [permissionStream] — [LocationPermissionStatus] changes
class LocationEngine {
  LocationEngine({
    LocationEngineConfig config = const LocationEngineConfig(),
  }) : _config = config;

  LocationEngineConfig _config;
  LocationEngineState _state = LocationEngineState.stopped;

  final _locationCtrl = StreamController<LocationUpdate>.broadcast();
  final _gpsStatusCtrl = StreamController<GpsStatus>.broadcast();
  final _permissionCtrl =
      StreamController<LocationPermissionStatus>.broadcast();

  StreamSubscription<Position>? _positionSub;
  StreamSubscription<ServiceStatus>? _gpsStatusSub;

  Stream<LocationUpdate> get locationStream => _locationCtrl.stream;
  Stream<GpsStatus> get gpsStatusStream => _gpsStatusCtrl.stream;
  Stream<LocationPermissionStatus> get permissionStream =>
      _permissionCtrl.stream;

  LocationEngineState get state => _state;
  LocationEngineConfig get config => _config;

  Future<void> start() async {
    if (_state != LocationEngineState.stopped) return;

    final permission = await _resolvePermission();
    if (permission != LocationPermissionStatus.granted) {
      _permissionCtrl.add(permission);
      return;
    }

    _startGpsStatusListener();

    final serviceEnabled = await Geolocator.isLocationServiceEnabled();
    _gpsStatusCtrl.add(
      serviceEnabled ? GpsStatus.enabled : GpsStatus.disabled,
    );

    if (serviceEnabled) {
      _startPositionStream();
    }

    _state = LocationEngineState.running;
  }

  void stop() {
    _cancelPositionSub();
    _gpsStatusSub?.cancel();
    _gpsStatusSub = null;
    _state = LocationEngineState.stopped;
  }

  void pause() {
    if (_state != LocationEngineState.running) return;
    _positionSub?.pause();
    _state = LocationEngineState.paused;
  }

  void resume() {
    if (_state != LocationEngineState.paused) return;
    _positionSub?.resume();
    _state = LocationEngineState.running;
  }

  Future<void> updateConfig(LocationEngineConfig config) async {
    _config = config;
    if (_state == LocationEngineState.running) {
      _cancelPositionSub();
      _startPositionStream();
    }
  }

  void dispose() {
    stop();
    _locationCtrl.close();
    _gpsStatusCtrl.close();
    _permissionCtrl.close();
  }

  void _startGpsStatusListener() {
    _gpsStatusSub?.cancel();
    _gpsStatusSub = Geolocator.getServiceStatusStream().listen((status) {
      final gpsStatus =
          status == ServiceStatus.enabled ? GpsStatus.enabled : GpsStatus.disabled;
      _gpsStatusCtrl.add(gpsStatus);

      if (gpsStatus == GpsStatus.enabled &&
          _state == LocationEngineState.running) {
        _startPositionStream();
      } else if (gpsStatus == GpsStatus.disabled) {
        _cancelPositionSub();
      }
    });
  }

  void _startPositionStream() {
    _cancelPositionSub();
    _positionSub = Geolocator.getPositionStream(
      locationSettings: _buildLocationSettings(),
    ).listen(
      (position) {
        _locationCtrl.add(LocationUpdate(
          latitude: position.latitude,
          longitude: position.longitude,
          accuracyMeters: position.accuracy,
          timestamp: position.timestamp,
          altitude: position.altitude,
          speed: position.speed,
          heading: position.heading,
        ));
      },
      onError: (Object error) {
        if (error is PermissionDeniedException) {
          _permissionCtrl.add(LocationPermissionStatus.denied);
          _cancelPositionSub();
        }
      },
    );
  }

  void _cancelPositionSub() {
    _positionSub?.cancel();
    _positionSub = null;
  }

  LocationSettings _buildLocationSettings() {
    final distanceM = _config.distanceFilter.round();
    if (Platform.isAndroid) {
      return AndroidSettings(
        accuracy: _config.accuracy,
        distanceFilter: distanceM,
        intervalDuration: Duration(milliseconds: _config.updateIntervalMs),
        forceLocationManager: false,
      );
    }
    if (Platform.isIOS || Platform.isMacOS) {
      return AppleSettings(
        accuracy: _config.accuracy,
        distanceFilter: distanceM,
        activityType: ActivityType.other,
        pauseLocationUpdatesAutomatically: false,
      );
    }
    return LocationSettings(
      accuracy: _config.accuracy,
      distanceFilter: distanceM,
    );
  }

  Future<LocationPermissionStatus> _resolvePermission() async {
    LocationPermission perm = await Geolocator.checkPermission();
    if (perm == LocationPermission.denied) {
      perm = await Geolocator.requestPermission();
    }
    return _toStatus(perm);
  }

  static LocationPermissionStatus _toStatus(LocationPermission perm) =>
      switch (perm) {
        LocationPermission.always ||
        LocationPermission.whileInUse =>
          LocationPermissionStatus.granted,
        LocationPermission.deniedForever =>
          LocationPermissionStatus.permanentlyDenied,
        _ => LocationPermissionStatus.denied,
      };
}
