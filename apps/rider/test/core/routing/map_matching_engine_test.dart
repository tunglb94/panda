import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:rider/core/location/location_engine.dart';
import 'package:rider/core/routing/map_matching_engine.dart';
import 'package:rider/core/routing/matched_location.dart';
import 'package:rider/core/routing/route_engine.dart';
import 'package:rider/core/routing/route_model.dart';
import 'package:rider/core/routing/route_point.dart';
import 'package:rider/core/routing/route_provider.dart';

// ─── Test doubles ─────────────────────────────────────────────────────────────

class _StubProvider implements RouteProvider {
  _StubProvider(this._route);
  final RouteModel _route;

  @override
  Future<RouteModel> calculateRoute(RoutePoint origin, RoutePoint destination) async => _route;
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// At the equator: 1° ≈ 111 320 m for both latitude and longitude.
// 30 m ≈ 0.000 270°   10 m ≈ 0.000 090°   100 m ≈ 0.000 899°

LocationUpdate _loc(double lat, double lng) => LocationUpdate(
      latitude: lat,
      longitude: lng,
      accuracyMeters: 5,
      timestamp: DateTime(2024),
      altitude: 0,
      speed: 0,
      heading: 0,
    );

/// Loads [route] into a fresh [RouteEngine], emits [gps] through a sync
/// [StreamController], and returns the single [MatchedLocation] emitted.
Future<MatchedLocation> _match({
  required RouteModel route,
  required LocationUpdate gps,
}) async {
  final engine = RouteEngine(provider: _StubProvider(route));
  await engine.loadRoute(
    const RoutePoint(latitude: 0, longitude: 0),
    const RoutePoint(latitude: 0, longitude: 0.01),
  );

  final ctrl = StreamController<LocationUpdate>.broadcast(sync: true);
  final matcher = MapMatchingEngine(
    locationStream: ctrl.stream,
    routeEngine: engine,
  );

  MatchedLocation? result;
  final sub = matcher.matchedLocationStream.listen((m) => result = m);
  matcher.start();
  ctrl.add(gps);

  await sub.cancel();
  matcher.dispose();
  await ctrl.close();
  engine.dispose();

  return result!;
}

// ─── Route factories ──────────────────────────────────────────────────────────

/// Straight horizontal segment: (0,0) → (0,0.01)  ≈ 1 113 m.
RouteModel _straightRoute() => RouteModel.fromDecodedPoints(
      decodedPoints: const [
        RoutePoint(latitude: 0, longitude: 0),
        RoutePoint(latitude: 0, longitude: 0.01),
      ],
      distanceMeters: 1113,
      durationSeconds: 120,
    );

/// L-shaped route: (0,0) → (0,0.01) → (0.01,0.01).
RouteModel _curvedRoute() => RouteModel.fromDecodedPoints(
      decodedPoints: const [
        RoutePoint(latitude: 0, longitude: 0),
        RoutePoint(latitude: 0, longitude: 0.01),
        RoutePoint(latitude: 0.01, longitude: 0.01),
      ],
      distanceMeters: 2226,
      durationSeconds: 240,
    );

// ─── Tests ────────────────────────────────────────────────────────────────────

void main() {
  group('MapMatchingEngine', () {
    // ── Straight-line projection ──────────────────────────────────────────────

    test('straight line — projection lands on route', () async {
      // GPS is 10 m north of the midpoint of a horizontal segment.
      // Expected: projected onto (0, 0.005) — the perpendicular foot.
      final m = await _match(
        route: _straightRoute(),
        gps: _loc(0.00009, 0.005), // ≈ 10 m north of midpoint
      );

      expect(m.isMatched, isTrue);
      expect(m.matchedPoint.latitude, closeTo(0.0, 1e-6));
      expect(m.matchedPoint.longitude, closeTo(0.005, 1e-5));
      expect(m.distanceFromRouteMeters, closeTo(10.0, 2.0));
      expect(m.nearestSegmentIndex, 0);
    });

    // ── Curved route — nearest segment selected ───────────────────────────────

    test('curved route — selects the closer segment', () async {
      // GPS is just west of the vertical leg (segment 1: lon=0.01).
      // Segment 0 is horizontal (lat=0); segment 1 is vertical (lon=0.01).
      // Point (0.005, 0.0099) is ~11 m from segment 1 and ~556 m from segment 0.
      final engine = RouteEngine(provider: _StubProvider(_curvedRoute()));
      await engine.loadRoute(
        const RoutePoint(latitude: 0, longitude: 0),
        const RoutePoint(latitude: 0.01, longitude: 0.01),
      );

      final ctrl = StreamController<LocationUpdate>.broadcast(sync: true);
      final matcher = MapMatchingEngine(
        locationStream: ctrl.stream,
        routeEngine: engine,
      );

      MatchedLocation? result;
      final sub = matcher.matchedLocationStream.listen((m) => result = m);
      matcher.start();
      ctrl.add(_loc(0.005, 0.0099));

      await sub.cancel();
      matcher.dispose();
      await ctrl.close();
      engine.dispose();

      expect(result, isNotNull);
      expect(result!.nearestSegmentIndex, 1);
      expect(result!.isMatched, isTrue);
      // Projected point should lie on the vertical leg (lon ≈ 0.01).
      expect(result!.matchedPoint.longitude, closeTo(0.01, 1e-5));
    });

    // ── GPS exactly on route ──────────────────────────────────────────────────

    test('GPS exactly on route — matched point equals original', () async {
      final m = await _match(
        route: _straightRoute(),
        gps: _loc(0, 0.005), // exactly on the midpoint
      );

      expect(m.isMatched, isTrue);
      expect(m.distanceFromRouteMeters, closeTo(0.0, 0.01));
      expect(m.matchedPoint.latitude, closeTo(0.0, 1e-9));
      expect(m.matchedPoint.longitude, closeTo(0.005, 1e-9));
    });

    // ── GPS 10 m away — projection succeeds ──────────────────────────────────

    test('GPS 10 m from route — isMatched true, projection close', () async {
      final m = await _match(
        route: _straightRoute(),
        gps: _loc(0.00009, 0.005), // ≈ 10 m north
      );

      expect(m.isMatched, isTrue);
      expect(m.distanceFromRouteMeters, lessThan(30.0));
      expect(m.matchedPoint.latitude, closeTo(0.0, 1e-5));
    });

    // ── GPS 100 m away — isMatched false ─────────────────────────────────────

    test('GPS 100 m from route — isMatched false, matchedPoint = original', () async {
      // 0.0009° lat ≈ 100 m at the equator.
      final m = await _match(
        route: _straightRoute(),
        gps: _loc(0.0009, 0.005), // ≈ 100 m north
      );

      expect(m.isMatched, isFalse);
      expect(m.distanceFromRouteMeters, greaterThan(30.0));
      // matchedPoint must equal originalPoint when not matched.
      expect(m.matchedPoint.latitude, m.originalPoint.latitude);
      expect(m.matchedPoint.longitude, m.originalPoint.longitude);
    });

    // ── End of route — projection clamps ─────────────────────────────────────

    test('GPS past end of route — progressPercent clamps to 1.0', () async {
      // GPS is past the last point; t should clamp to 1.0 → end of route.
      final m = await _match(
        route: _straightRoute(),
        gps: _loc(0, 0.015), // 0.005° past the endpoint (0,0.01)
      );

      expect(m.nearestSegmentIndex, 0);
      expect(m.progressPercent, closeTo(1.0, 0.01));
    });

    // ── No route loaded — no emission ─────────────────────────────────────────

    test('no route loaded — no emission', () async {
      final engine = RouteEngine(
        provider: _StubProvider(_straightRoute()),
      );
      // Deliberately skip loadRoute.

      final ctrl = StreamController<LocationUpdate>.broadcast(sync: true);
      final matcher = MapMatchingEngine(
        locationStream: ctrl.stream,
        routeEngine: engine,
      );

      MatchedLocation? result;
      final sub = matcher.matchedLocationStream.listen((m) => result = m);
      matcher.start();
      ctrl.add(_loc(0, 0.005));

      await sub.cancel();
      matcher.dispose();
      await ctrl.close();
      engine.dispose();

      expect(result, isNull);
    });
  });
}
