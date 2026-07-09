import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:rider/core/location/location_engine.dart';
import 'package:rider/core/routing/route_engine.dart';
import 'package:rider/core/routing/route_model.dart';
import 'package:rider/core/routing/route_point.dart';
import 'package:rider/core/routing/route_progress.dart';
import 'package:rider/core/routing/route_progress_engine.dart';
import 'package:rider/core/routing/route_provider.dart';

// ─── Test doubles ─────────────────────────────────────────────────────────────

class _StubProvider implements RouteProvider {
  _StubProvider(this._route);
  final RouteModel _route;

  @override
  Future<RouteModel> calculateRoute(RoutePoint origin, RoutePoint destination) async => _route;
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

LocationUpdate _loc(double lat, double lng) => LocationUpdate(
      latitude: lat,
      longitude: lng,
      accuracyMeters: 5,
      timestamp: DateTime(2024),
      altitude: 0,
      speed: 0,
      heading: 0,
    );

// ─── Test route ───────────────────────────────────────────────────────────────
//
// Three points: (0,0) → (0,0.001) → (0,0.002)
// At lat≈0 each 0.001° longitude ≈ 111.32 m, so total ≈ 222.64 m, duration 300 s.

RouteModel _makeRoute() => RouteModel.fromDecodedPoints(
      decodedPoints: const [
        RoutePoint(latitude: 0, longitude: 0),
        RoutePoint(latitude: 0, longitude: 0.001),
        RoutePoint(latitude: 0, longitude: 0.002),
      ],
      distanceMeters: 223,
      durationSeconds: 300,
    );

// ─── Helper to create engine and drive it ────────────────────────────────────

Future<RouteProgress?> _driveEngine({
  required RouteModel route,
  required LocationUpdate gpsUpdate, // use _loc() helper
}) async {
  final engine = RouteEngine(provider: _StubProvider(route));
  await engine.loadRoute(
    const RoutePoint(latitude: 0, longitude: 0),
    const RoutePoint(latitude: 0, longitude: 0.002),
  );

  final ctrl = StreamController<LocationUpdate>.broadcast(sync: true);
  final progressEngine = RouteProgressEngine(
    locationStream: ctrl.stream,
    routeEngine: engine,
    jitterThresholdMeters: 0, // disable jitter for deterministic tests
  );

  RouteProgress? result;
  final sub = progressEngine.progressStream.listen((p) => result = p);
  progressEngine.start();
  ctrl.add(gpsUpdate);
  await sub.cancel();
  progressEngine.dispose();
  await ctrl.close();
  engine.dispose();
  return result;
}

// ─── Tests ────────────────────────────────────────────────────────────────────

void main() {
  final route = _makeRoute();

  test('at start → progressPercent ≈ 0', () async {
    final p = await _driveEngine(
      route: route,
      gpsUpdate: _loc(0, 0),
    );
    expect(p, isNotNull);
    expect(p!.progressPercent, closeTo(0.0, 0.01));
    expect(p.isOnRoute, isTrue);
  });

  test('at midpoint → progressPercent ≈ 0.5', () async {
    final p = await _driveEngine(
      route: route,
      gpsUpdate: _loc(0, 0.001),
    );
    expect(p, isNotNull);
    expect(p!.progressPercent, closeTo(0.5, 0.02));
    expect(p.isOnRoute, isTrue);
  });

  test('at end → progressPercent ≈ 1.0', () async {
    final p = await _driveEngine(
      route: route,
      gpsUpdate: _loc(0, 0.002),
    );
    expect(p, isNotNull);
    expect(p!.progressPercent, closeTo(1.0, 0.01));
    expect(p.isOnRoute, isTrue);
  });

  test('far off route → isOnRoute = false', () async {
    // 0.01° lat ≈ 1.1 km — well beyond 30 m threshold
    final p = await _driveEngine(
      route: route,
      gpsUpdate: _loc(0.01, 0.001),
    );
    expect(p, isNotNull);
    expect(p!.isOnRoute, isFalse);
  });

  test('remainingMeters decreases as position advances', () async {
    final route = _makeRoute();
    final engine = RouteEngine(provider: _StubProvider(route));
    await engine.loadRoute(
      const RoutePoint(latitude: 0, longitude: 0),
      const RoutePoint(latitude: 0, longitude: 0.002),
    );

    final ctrl = StreamController<LocationUpdate>.broadcast(sync: true);
    final progressEngine = RouteProgressEngine(
      locationStream: ctrl.stream,
      routeEngine: engine,
      jitterThresholdMeters: 0,
    );

    final results = <RouteProgress>[];
    final sub = progressEngine.progressStream.listen(results.add);
    progressEngine.start();

    ctrl.add(_loc(0, 0));
    ctrl.add(_loc(0, 0.0005));
    ctrl.add(_loc(0, 0.001));
    ctrl.add(_loc(0, 0.002));

    await sub.cancel();
    progressEngine.dispose();
    await ctrl.close();
    engine.dispose();

    expect(results.length, 4);
    for (int i = 1; i < results.length; i++) {
      expect(
        results[i].remainingMeters,
        lessThanOrEqualTo(results[i - 1].remainingMeters),
        reason: 'remainingMeters should not increase as we advance',
      );
    }
  });

  test('no route loaded → no emission', () async {
    final engine = RouteEngine(provider: _StubProvider(route));
    // deliberately do NOT call loadRoute

    final ctrl = StreamController<LocationUpdate>.broadcast(sync: true);
    final progressEngine = RouteProgressEngine(
      locationStream: ctrl.stream,
      routeEngine: engine,
    );

    RouteProgress? result;
    final sub = progressEngine.progressStream.listen((p) => result = p);
    progressEngine.start();
    ctrl.add(_loc(0, 0));

    await sub.cancel();
    progressEngine.dispose();
    await ctrl.close();
    engine.dispose();

    expect(result, isNull);
  });
}
