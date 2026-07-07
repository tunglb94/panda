import 'dart:async';
import 'dart:math';

import 'package:google_maps_flutter/google_maps_flutter.dart';

import 'package:rider/core/location/location_engine.dart';
import 'route_model.dart';
import 'route_progress_model.dart';

/// Continuously matches device GPS position against the active [RouteModel]
/// and emits [RouteProgressModel] updates via [progressStream].
///
/// The engine does NOT start or stop the [LocationEngine] — the caller owns
/// that lifecycle. Call [start] to begin listening and [stop] / [dispose]
/// when done.
class RouteProgressEngine {
  RouteProgressEngine({
    required RouteModel route,
    required LocationEngine locationEngine,
    double onRouteThresholdMeters = 50.0,
    double jitterThresholdMeters = 5.0,
  })  : _route = route,
        _locationEngine = locationEngine,
        _threshold = onRouteThresholdMeters,
        _jitter = jitterThresholdMeters {
    _precompute();
  }

  final RouteModel _route;
  final LocationEngine _locationEngine;
  final double _threshold;
  final double _jitter;

  // Precomputed cumulative distances (metres) from route start to each point.
  late final List<double> _cumDist;
  late final double _totalDistance;

  final _ctrl = StreamController<RouteProgressModel>.broadcast();
  StreamSubscription<LocationUpdate>? _sub;
  LatLng? _lastPos;

  Stream<RouteProgressModel> get progressStream => _ctrl.stream;

  // ─── Lifecycle ──────────────────────────────────────────────────────────────

  void start() {
    if (_sub != null) return;
    _sub = _locationEngine.locationStream.listen(_onLocationUpdate);
  }

  void stop() {
    _sub?.cancel();
    _sub = null;
  }

  void dispose() {
    stop();
    if (!_ctrl.isClosed) _ctrl.close();
  }

  // ─── Precomputation ──────────────────────────────────────────────────────────

  void _precompute() {
    final pts = _route.polylinePoints;
    _cumDist = List.filled(pts.length, 0.0);
    for (int i = 1; i < pts.length; i++) {
      _cumDist[i] = _cumDist[i - 1] + _haversine(pts[i - 1], pts[i]);
    }
    _totalDistance = pts.length > 1 ? _cumDist.last : 0.0;
  }

  // ─── GPS update handler ───────────────────────────────────────────────────────

  void _onLocationUpdate(LocationUpdate update) {
    if (_ctrl.isClosed) return;
    final pos = LatLng(update.latitude, update.longitude);

    // Ignore GPS jitter — don't recompute if barely moved.
    if (_lastPos != null && _haversine(_lastPos!, pos) < _jitter) return;
    _lastPos = pos;

    final nearest = _findNearest(pos);
    final progress = _totalDistance > 0
        ? (nearest.cumulative / _totalDistance).clamp(0.0, 1.0)
        : 0.0;
    final remainingMeters =
        (_totalDistance - nearest.cumulative).clamp(0.0, double.infinity).round();
    final remainingSecs =
        (_route.durationSeconds * (1.0 - progress)).round();

    _ctrl.add(RouteProgressModel(
      progressPercent: progress,
      remainingDistance: remainingMeters,
      remainingDuration: remainingSecs,
      isOnRoute: nearest.distToRoute < _threshold,
      nearestRoutePoint: nearest.point,
    ));
  }

  // ─── Geometry ─────────────────────────────────────────────────────────────────

  _NearestResult _findNearest(LatLng pos) {
    final pts = _route.polylinePoints;
    if (pts.isEmpty) return _NearestResult(pos, double.infinity, 0.0);
    if (pts.length == 1) {
      return _NearestResult(pts.first, _haversine(pos, pts.first), 0.0);
    }

    double minDist = double.infinity;
    LatLng nearestPt = pts.first;
    double nearestCum = 0.0;

    for (int i = 0; i < pts.length - 1; i++) {
      final sr = _closestOnSegment(pos, pts[i], pts[i + 1]);
      final d = _haversine(pos, sr.point);
      if (d < minDist) {
        minDist = d;
        nearestPt = sr.point;
        // Interpolate cumulative distance within the segment.
        nearestCum = _cumDist[i] + sr.t * (_cumDist[i + 1] - _cumDist[i]);
      }
    }
    return _NearestResult(nearestPt, minDist, nearestCum);
  }

  /// Returns the closest point on segment [a]→[b] to [p] using an
  /// equirectangular projection — accurate enough for the short polyline
  /// segments produced by the Directions API.
  static _SegmentResult _closestOnSegment(LatLng p, LatLng a, LatLng b) {
    final abx = b.longitude - a.longitude;
    final aby = b.latitude - a.latitude;
    final apx = p.longitude - a.longitude;
    final apy = p.latitude - a.latitude;
    final ab2 = abx * abx + aby * aby;
    if (ab2 < 1e-18) return _SegmentResult(0.0, a); // zero-length segment
    final t = ((apx * abx + apy * aby) / ab2).clamp(0.0, 1.0);
    return _SegmentResult(
      t,
      LatLng(a.latitude + t * aby, a.longitude + t * abx),
    );
  }

  static double _haversine(LatLng a, LatLng b) {
    const r = 6371000.0;
    final lat1 = a.latitude * pi / 180;
    final lat2 = b.latitude * pi / 180;
    final dLat = (b.latitude - a.latitude) * pi / 180;
    final dLon = (b.longitude - a.longitude) * pi / 180;
    final sinDLat = sin(dLat / 2);
    final sinDLon = sin(dLon / 2);
    final x = sinDLat * sinDLat + cos(lat1) * cos(lat2) * sinDLon * sinDLon;
    return r * 2 * atan2(sqrt(x), sqrt(1 - x));
  }
}

// ─── Internal result types ────────────────────────────────────────────────────

class _NearestResult {
  const _NearestResult(this.point, this.distToRoute, this.cumulative);
  final LatLng point;
  final double distToRoute;
  final double cumulative;
}

class _SegmentResult {
  const _SegmentResult(this.t, this.point);
  final double t;
  final LatLng point;
}
