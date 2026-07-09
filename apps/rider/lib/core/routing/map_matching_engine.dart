import 'dart:async';
import 'dart:math';

import 'package:rider/core/location/location_engine.dart';
import 'matched_location.dart';
import 'route_engine.dart';
import 'route_model.dart';
import 'route_point.dart';

/// Projects each GPS fix onto the active route polyline and emits a
/// [MatchedLocation] for every update.
///
/// The engine is the single source of truth for corrected position. Downstream
/// consumers (Route Progress Engine, ETA Engine, Dispatch) must read from
/// [matchedLocationStream] instead of raw GPS.
///
/// Lifecycle: call [start] to begin listening, [stop] to pause (e.g., when the
/// app backgrounds), and [dispose] when the owning widget is torn down.
/// The engine never owns [RouteEngine] or the [locationStream] lifecycle.
class MapMatchingEngine {
  MapMatchingEngine({
    required Stream<LocationUpdate> locationStream,
    required RouteEngine routeEngine,
    double matchThresholdMeters = 30.0,
  })  : _locationStream = locationStream,
        _routeEngine = routeEngine,
        _threshold = matchThresholdMeters;

  final Stream<LocationUpdate> _locationStream;
  final RouteEngine _routeEngine;
  final double _threshold;

  final _ctrl = StreamController<MatchedLocation>.broadcast();
  StreamSubscription<LocationUpdate>? _sub;

  // Precomputed route state — rebuilt only when the route object changes.
  RouteModel? _cachedRoute;
  List<double> _cumDist = const [];
  double _totalDistance = 0.0;

  /// Broadcast stream of corrected locations.
  ///
  /// Emits once per GPS update whenever [start] has been called and a route is
  /// loaded in [RouteEngine].  Completes when [dispose] is called.
  Stream<MatchedLocation> get matchedLocationStream => _ctrl.stream;

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

    // Rebuild precomputed data only when the route object identity changes.
    if (!identical(route, _cachedRoute)) {
      _cachedRoute = route;
      _rebuildCumulativeDistances(route);
    }

    final original = RoutePoint(
      latitude: update.latitude,
      longitude: update.longitude,
    );

    final result = _findNearest(original, route);
    final dist = _haversine(original, result.projectedPoint);
    final isMatched = dist <= _threshold;
    final progress = _totalDistance > 0
        ? (result.cumulative / _totalDistance).clamp(0.0, 1.0)
        : 0.0;

    _ctrl.add(MatchedLocation(
      originalPoint: original,
      matchedPoint: isMatched ? result.projectedPoint : original,
      distanceFromRouteMeters: dist,
      nearestSegmentIndex: result.segmentIndex,
      progressMeters: result.cumulative,
      progressPercent: progress,
      isMatched: isMatched,
      timestamp: update.timestamp,
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
    if (pts.isEmpty) {
      return _NearestResult(
        segmentIndex: 0,
        projectedPoint: pos,
        cumulative: 0.0,
      );
    }
    if (pts.length == 1) {
      return _NearestResult(
        segmentIndex: 0,
        projectedPoint: pts.first,
        cumulative: 0.0,
      );
    }

    double minDist = double.infinity;
    int bestIndex = 0;
    RoutePoint bestProjection = pts.first;
    double bestCumulative = 0.0;

    for (int i = 0; i < pts.length - 1; i++) {
      final sr = _projectOntoSegment(pos, pts[i], pts[i + 1]);
      final d = _haversine(pos, sr.point);
      if (d < minDist) {
        minDist = d;
        bestIndex = i;
        bestProjection = sr.point;
        // Interpolate cumulative distance at the projected point within segment.
        bestCumulative = _cumDist[i] + sr.t * (_cumDist[i + 1] - _cumDist[i]);
      }
    }

    return _NearestResult(
      segmentIndex: bestIndex,
      projectedPoint: bestProjection,
      cumulative: bestCumulative,
    );
  }

  /// Perpendicular projection of [p] onto segment [a]→[b].
  ///
  /// Uses an equirectangular (degree-space) dot product — accurate for the
  /// short segments produced by the Directions API.  The parameter [t] is
  /// clamped to [0, 1] so the result always lies on the segment.
  static _SegmentResult _projectOntoSegment(
    RoutePoint p,
    RoutePoint a,
    RoutePoint b,
  ) {
    final abLat = b.latitude - a.latitude;
    final abLon = b.longitude - a.longitude;
    final apLat = p.latitude - a.latitude;
    final apLon = p.longitude - a.longitude;
    final ab2 = abLat * abLat + abLon * abLon;
    if (ab2 < 1e-18) {
      // Zero-length segment — return the segment start.
      return _SegmentResult(t: 0.0, point: a);
    }
    final t = ((apLat * abLat + apLon * abLon) / ab2).clamp(0.0, 1.0);
    return _SegmentResult(
      t: t,
      point: RoutePoint(
        latitude: a.latitude + t * abLat,
        longitude: a.longitude + t * abLon,
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
  const _NearestResult({
    required this.segmentIndex,
    required this.projectedPoint,
    required this.cumulative,
  });
  final int segmentIndex;
  final RoutePoint projectedPoint;
  final double cumulative;
}

class _SegmentResult {
  const _SegmentResult({required this.t, required this.point});
  final double t;
  final RoutePoint point;
}
