import 'dart:async';
import 'dart:math';

import '../location/location_engine.dart';
import 'trip_metrics.dart';

/// Production-quality trip metrics engine.
///
/// Subscribes to an existing [Stream<LocationUpdate>] (from [LocationEngine])
/// and computes accurate distance and duration figures suitable for fare
/// calculation.
///
/// ## GPS quality filters applied in [addLocation]:
/// 1. Accuracy > 20 m — low-quality fix; discarded.
/// 2. Movement < 5 m from previous accepted fix — GPS drift / noise; discarded.
/// 3. Duplicate coordinates — discarded.
/// 4. Implied speed > 180 km/h (50 m/s) — impossible jump; discarded.
///
/// ## Moving vs idle
/// Moving time accumulates only for intervals between consecutive accepted
/// fixes that are ≤ 45 seconds apart.  Longer gaps (vehicle stationary or GPS
/// signal lost) flow into [TripMetrics.idleDurationSeconds].
///
/// ## Thread safety
/// Not thread-safe.  Call all methods from the Flutter UI isolate.
class TripMetricsEngine {
  TripMetricsEngine({required Stream<LocationUpdate> locationStream})
      : _locationStream = locationStream;

  final Stream<LocationUpdate> _locationStream;
  StreamSubscription<LocationUpdate>? _sub;

  // ─── Internal accumulator state ──────────────────────────────────────────

  DateTime? _startedAt;
  DateTime? _finishedAt;

  double _totalDistanceMeters = 0;
  double _movingDurationSeconds = 0;
  double _maxSpeedMps = 0;
  int _gpsSampleCount = 0;

  double? _lastLat;
  double? _lastLon;
  DateTime? _lastAcceptedTime;

  bool _running = false;

  // ─── Public API ──────────────────────────────────────────────────────────

  /// Current metrics snapshot (safe to call at any time).
  TripMetrics get metrics => _buildMetrics();

  /// Starts recording. Subscribes to [locationStream] and records [startedAt].
  /// No-op if already running.
  void start() {
    if (_running) return;
    _running = true;
    _startedAt = DateTime.now();
    _sub = _locationStream.listen(addLocation);
  }

  /// Processes a single GPS sample through all quality filters and, if it
  /// passes, accumulates distance and duration.
  ///
  /// Called automatically from the stream subscription; can also be called
  /// directly for testing.
  void addLocation(LocationUpdate update) {
    if (!_running) return;

    // ── Filter 1: poor accuracy ──────────────────────────────────────────
    if (update.accuracyMeters > 20) return;

    final lat = update.latitude;
    final lon = update.longitude;

    // ── Filter 3: duplicate coordinates ─────────────────────────────────
    if (lat == _lastLat && lon == _lastLon) return;

    // ── Filter 4a: impossible GPS-reported speed ─────────────────────────
    if (update.speed >= 0 && update.speed > 50.0) return;

    // ── First accepted fix: initialise tracking, no distance to accumulate ─
    if (_lastLat == null) {
      _lastLat = lat;
      _lastLon = lon;
      _lastAcceptedTime = update.timestamp;
      _gpsSampleCount++;
      if (update.speed >= 0 && update.speed > _maxSpeedMps) {
        _maxSpeedMps = update.speed;
      }
      return;
    }

    // ── Subsequent fixes ─────────────────────────────────────────────────
    final distM = _haversine(_lastLat!, _lastLon!, lat, lon);

    // Filter 2: micro-movement / GPS drift
    if (distM < 5) return;

    // Filter 4b: impossible jump via implied speed
    final elapsedMs =
        update.timestamp.difference(_lastAcceptedTime!).inMilliseconds;
    if (elapsedMs > 0) {
      final jumpSpeedMps = distM / (elapsedMs / 1000.0);
      if (jumpSpeedMps > 50.0) return;

      // Count gap as moving only if it is short enough (≤ 45 s).
      // Longer gaps indicate the vehicle was stopped; they flow to idle time.
      if (elapsedMs <= 45000) {
        _movingDurationSeconds += elapsedMs / 1000.0;
      }
    }

    _totalDistanceMeters += distM;

    if (update.speed >= 0 && update.speed > _maxSpeedMps) {
      _maxSpeedMps = update.speed;
    }

    _gpsSampleCount++;
    _lastLat = lat;
    _lastLon = lon;
    _lastAcceptedTime = update.timestamp;
  }

  /// Stops recording and returns the final immutable [TripMetrics].
  ///
  /// Idempotent: calling [finish] multiple times records [finishedAt] only
  /// once and always returns the same metrics snapshot.
  TripMetrics finish() {
    _sub?.cancel();
    _sub = null;
    _running = false;
    _finishedAt ??= DateTime.now();
    return _buildMetrics();
  }

  /// Resets all accumulated state.  Cancels any active subscription.
  void reset() {
    _sub?.cancel();
    _sub = null;
    _running = false;
    _startedAt = null;
    _finishedAt = null;
    _totalDistanceMeters = 0;
    _movingDurationSeconds = 0;
    _maxSpeedMps = 0;
    _gpsSampleCount = 0;
    _lastLat = null;
    _lastLon = null;
    _lastAcceptedTime = null;
  }

  // ─── Internals ────────────────────────────────────────────────────────────

  TripMetrics _buildMetrics() {
    final totalSecs = _startedAt != null
        ? ((_finishedAt ?? DateTime.now())
                .difference(_startedAt!)
                .inMilliseconds /
            1000.0)
        : 0.0;
    final idleSecs =
        (totalSecs - _movingDurationSeconds).clamp(0.0, double.infinity);
    final avgSpeedMps = _movingDurationSeconds > 0
        ? _totalDistanceMeters / _movingDurationSeconds
        : 0.0;

    return TripMetrics(
      totalDistanceMeters: _totalDistanceMeters,
      movingDistanceMeters: _totalDistanceMeters,
      idleDurationSeconds: idleSecs,
      movingDurationSeconds: _movingDurationSeconds,
      averageSpeedMps: avgSpeedMps,
      maxSpeedMps: _maxSpeedMps,
      gpsSampleCount: _gpsSampleCount,
      startedAt: _startedAt,
      finishedAt: _finishedAt,
    );
  }

  /// Haversine distance between two WGS-84 coordinates in metres.
  static double _haversine(
      double lat1, double lon1, double lat2, double lon2) {
    const earthRadius = 6371000.0;
    final dLat = _rad(lat2 - lat1);
    final dLon = _rad(lon2 - lon1);
    final a = sin(dLat / 2) * sin(dLat / 2) +
        cos(_rad(lat1)) * cos(_rad(lat2)) * sin(dLon / 2) * sin(dLon / 2);
    final c = 2 * atan2(sqrt(a), sqrt(1 - a));
    return earthRadius * c;
  }

  static double _rad(double deg) => deg * pi / 180.0;
}
