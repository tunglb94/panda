import 'dart:async';
import 'dart:io' show Platform;

import 'package:geolocator/geolocator.dart';
import 'package:rider/core/location/location_engine_config.dart';

// ─── Public value types ───────────────────────────────────────────────────────

/// An immutable GPS fix emitted by [LocationEngine.locationStream].
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

  /// Horizontal accuracy radius in metres.
  final double accuracyMeters;

  final DateTime timestamp;

  /// Altitude in metres above WGS84 ellipsoid. 0 if unavailable.
  final double altitude;

  /// Speed in m/s. Negative when the platform cannot determine speed.
  final double speed;

  /// Bearing in degrees clockwise from north. Negative when unavailable.
  final double heading;

  @override
  String toString() =>
      'LocationUpdate(lat: $latitude, lng: $longitude, '
      'acc: ${accuracyMeters.toStringAsFixed(1)} m, ts: $timestamp)';
}

// ─── Status enums ─────────────────────────────────────────────────────────────

/// Whether the device's Location Services switch is on or off.
enum GpsStatus { enabled, disabled }

/// The app's current location permission level.
enum LocationPermissionStatus { granted, denied, permanentlyDenied }

/// Lifecycle state of the [LocationEngine].
enum LocationEngineState { stopped, running, paused }

// ─── Engine ───────────────────────────────────────────────────────────────────

/// Reusable, stream-based GPS engine.
///
/// Exposes three broadcast streams:
/// - [locationStream]  — continuous [LocationUpdate] events
/// - [gpsStatusStream] — [GpsStatus] changes (GPS on/off)
/// - [permissionStream] — [LocationPermissionStatus] changes
///
/// Typical usage:
/// ```dart
/// final engine = LocationEngine();
/// engine.locationStream.listen((u) { ... });
/// engine.gpsStatusStream.listen((s) { ... });
/// engine.permissionStream.listen((p) { ... });
/// await engine.start();
/// // …
/// engine.pause();
/// engine.resume();
/// engine.stop();
/// engine.dispose(); // call when permanently done
/// ```
class LocationEngine {
  LocationEngine({
    LocationEngineConfig config = const LocationEngineConfig(),
  }) : _config = config;

  LocationEngineConfig _config;
  LocationEngineState _state = LocationEngineState.stopped;

  final _locationCtrl = StreamController<LocationUpdate>.broadcast();
  final _gpsStatusCtrl = StreamController<GpsStatus>.broadcast();
  final _permissionCtrl = StreamController<LocationPermissionStatus>.broadcast();

  StreamSubscription<Position>? _positionSub;
  StreamSubscription<ServiceStatus>? _gpsStatusSub;

  // ─── Public streams ─────────────────────────────────────────────────────────

  /// Emits a [LocationUpdate] for every GPS fix that passes the
  /// [LocationEngineConfig.distanceFilter].
  Stream<LocationUpdate> get locationStream => _locationCtrl.stream;

  /// Emits [GpsStatus.enabled] or [GpsStatus.disabled] whenever the device's
  /// Location Services switch changes while the engine is running.
  Stream<GpsStatus> get gpsStatusStream => _gpsStatusCtrl.stream;

  /// Emits [LocationPermissionStatus] when the app's permission level changes
  /// while the engine is running (e.g. user revokes access in Settings).
  Stream<LocationPermissionStatus> get permissionStream =>
      _permissionCtrl.stream;

  // ─── Public state ───────────────────────────────────────────────────────────

  LocationEngineState get state => _state;
  LocationEngineConfig get config => _config;

  // ─── Lifecycle ──────────────────────────────────────────────────────────────

  /// Checks permissions, starts the GPS status listener, then begins streaming
  /// position updates. No-op if already [running] or [paused].
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

  /// Stops position updates and the GPS status listener.
  /// The engine returns to [LocationEngineState.stopped] and may be
  /// [start]ed again.
  void stop() {
    _cancelPositionSub();
    _gpsStatusSub?.cancel();
    _gpsStatusSub = null;
    _state = LocationEngineState.stopped;
  }

  /// Suspends delivery of position events without cancelling the subscription.
  /// The GPS status listener continues running.
  /// No-op unless currently [LocationEngineState.running].
  void pause() {
    if (_state != LocationEngineState.running) return;
    _positionSub?.pause();
    _state = LocationEngineState.paused;
  }

  /// Resumes position event delivery after a [pause].
  /// No-op unless currently [LocationEngineState.paused].
  void resume() {
    if (_state != LocationEngineState.paused) return;
    _positionSub?.resume();
    _state = LocationEngineState.running;
  }

  /// Applies a new configuration. If the engine is running, the position
  /// stream is restarted immediately with the updated settings.
  Future<void> updateConfig(LocationEngineConfig config) async {
    _config = config;
    if (_state == LocationEngineState.running) {
      _cancelPositionSub();
      _startPositionStream();
    }
  }

  /// Closes all streams and releases resources. The engine cannot be reused
  /// after [dispose].
  void dispose() {
    stop();
    _locationCtrl.close();
    _gpsStatusCtrl.close();
    _permissionCtrl.close();
  }

  // ─── Internal ───────────────────────────────────────────────────────────────

  void _startGpsStatusListener() {
    _gpsStatusSub?.cancel();
    _gpsStatusSub = Geolocator.getServiceStatusStream().listen((status) {
      final gpsStatus =
          status == ServiceStatus.enabled ? GpsStatus.enabled : GpsStatus.disabled;
      _gpsStatusCtrl.add(gpsStatus);

      if (gpsStatus == GpsStatus.enabled &&
          _state == LocationEngineState.running) {
        // GPS re-enabled while engine is running — restart position stream.
        _startPositionStream();
      } else if (gpsStatus == GpsStatus.disabled) {
        // GPS turned off — cancel position stream; keep engine state as
        // `running` so that position updates resume automatically when GPS
        // comes back on.
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
        // LocationServiceDisabledException is handled via _startGpsStatusListener.
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
    // Fallback for non-mobile targets.
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
