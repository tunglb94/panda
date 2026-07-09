import 'dart:async';
import 'dart:math';

import 'package:rider/core/location/location_engine.dart';
import 'route_engine.dart';
import 'route_model.dart';
import 'route_point.dart';
import 'route_progress.dart';

/// Continuously matches GPS position against the active [RouteEngine] route
/// and emits [RouteProgress] updates via [progressStream].
///
/// The engine subscribes to [locationStream] on [start] and unsubscribes on
/// [stop] / [dispose]. It never owns the [RouteEngine] lifecycle.
class RouteProgressEngine {
  RouteProgressEngine({
    required Stream<LocationUpdate> locationStream,
    required RouteEngine routeEngine,
    double onRouteThresholdMeters = 30.0,
    double jitterThresholdMeters = 5.0,
  })  : _locationStream = locationStream,
        _routeEngine = routeEngine,
        _threshold = onRouteThresholdMeters,
        _jitter = jitterThresholdMeters;

  final Stream<LocationUpdate> _locationStream;
  final RouteEngine _routeEngine;
  final double _threshold;
  final double _jitter;

  final _ctrl = StreamController<RouteProgress>.broadcast();
  StreamSubscription<LocationUpdate>? _sub;

  RouteModel? _cachedRoute;
  List<double> _cumDist = const [];
  double _totalDistance = 0.0;
  RoutePoint? _lastPos;

  Stream<RouteProgress> get progressStream => _ctrl.stream;

  // ─── Lifecycle ───────────────────────────────────────────────────────────────

  void start() {
    if (_sub != null) return;
    _sub = _locationStream.listen(_onLocationUpdate);
  }

  void stop() {
    _sub?.cancel();
    _sub = null;
  }

  void dispose() {
    stop();
    if (!_ctrl.isClosed) _ctrl.close();
  }

  // ─── GPS update handler ──────────────────────────────────────────────────────

  void _onLocationUpdate(LocationUpdate update) {
    if (_ctrl.isClosed) return;

    final route = _routeEngine.currentRoute;
    if (route == null) return;

    // Recompute cumulative distances only when the route object changes.
    if (!identical(route, _cachedRoute)) {
      _cachedRoute = route;
      _rebuildCumulativeDistances(route);
      _lastPos = null; // reset jitter baseline on new route
    }

    final pos = RoutePoint(latitude: update.latitude, longitude: update.longitude);

    // Suppress GPS jitter — skip if position hasn't moved meaningfully.
    if (_lastPos != null && _haversine(_lastPos!, pos) < _jitter) return;
    _lastPos = pos;

    final nearest = _findNearest(pos, route);
    final progress = _totalDistance > 0
        ? (nearest.cumulative / _totalDistance).clamp(0.0, 1.0)
        : 0.0;
    final completedMeters = nearest.cumulative.round();
    final remainingMeters =
        (_totalDistance - nearest.cumulative).clamp(0.0, double.infinity).round();
    final remainingSecs = (route.durationSeconds * (1.0 - progress)).round();

    _ctrl.add(RouteProgress(
      progressPercent: progress,
      travelledMeters: completedMeters,
      remainingMeters: remainingMeters,
      completedDistanceMeters: completedMeters,
      remainingDurationSeconds: remainingSecs,
      nearestRouteIndex: nearest.segmentIndex,
      isOnRoute: nearest.distToRoute < _threshold,
    ));
  }

  // ─── Precomputation ──────────────────────────────────────────────────────────

  void _rebuildCumulativeDistances(RouteModel route) {
    final pts = route.decodedPolyline;
    if (pts.length < 2) {
      _cumDist = pts.isEmpty ? const [] : [0.0];
      _totalDistance = 0.0;
      return;
    }
    final d = List<double>.filled(pts.length, 0.0);
    for (int i = 1; i < pts.length; i++) {
      d[i] = d[i - 1] + _haversine(pts[i - 1], pts[i]);
    }
    _cumDist = d;
    _totalDistance = d.last;
  }

  // ─── Geometry ────────────────────────────────────────────────────────────────

  _NearestResult _findNearest(RoutePoint pos, RouteModel route) {
    final pts = route.decodedPolyline;
    if (pts.isEmpty) return _NearestResult(0, double.infinity, 0.0);
    if (pts.length == 1) {
      return _NearestResult(0, _haversine(pos, pts.first), 0.0);
    }

    double minDist = double.infinity;
    int bestIndex = 0;
    double bestCum = 0.0;

    for (int i = 0; i < pts.length - 1; i++) {
      final sr = _closestOnSegment(pos, pts[i], pts[i + 1]);
      final d = _haversine(pos, sr.point);
      if (d < minDist) {
        minDist = d;
        bestIndex = i;
        bestCum = _cumDist[i] + sr.t * (_cumDist[i + 1] - _cumDist[i]);
      }
    }
    return _NearestResult(bestIndex, minDist, bestCum);
  }

  static _SegmentResult _closestOnSegment(
      RoutePoint p, RoutePoint a, RoutePoint b) {
    final abx = b.longitude - a.longitude;
    final aby = b.latitude - a.latitude;
    final apx = p.longitude - a.longitude;
    final apy = p.latitude - a.latitude;
    final ab2 = abx * abx + aby * aby;
    if (ab2 < 1e-18) return _SegmentResult(0.0, a);
    final t = ((apx * abx + apy * aby) / ab2).clamp(0.0, 1.0);
    return _SegmentResult(
      t,
      RoutePoint(
        latitude: a.latitude + t * aby,
        longitude: a.longitude + t * abx,
      ),
    );
  }

  static double _haversine(RoutePoint a, RoutePoint b) {
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
  const _NearestResult(this.segmentIndex, this.distToRoute, this.cumulative);
  final int segmentIndex;
  final double distToRoute;
  final double cumulative;
}

class _SegmentResult {
  const _SegmentResult(this.t, this.point);
  final double t;
  final RoutePoint point;
}
