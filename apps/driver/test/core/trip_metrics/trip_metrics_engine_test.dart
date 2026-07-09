import 'dart:async';

import 'package:flutter_test/flutter_test.dart';

import 'package:driver/core/location/location_engine.dart';
import 'package:driver/core/trip_metrics/trip_metrics_engine.dart';

// ─── Helpers ─────────────────────────────────────────────────────────────────

/// Broadcast stream that never emits, so tests can drive [addLocation] directly.
Stream<LocationUpdate> _neverStream() =>
    StreamController<LocationUpdate>.broadcast().stream;

/// Constructs a [LocationUpdate] at the given [lat]/[lon] with sensible
/// defaults for accuracy and speed.
LocationUpdate _fix({
  required double lat,
  required double lon,
  double accuracy = 10.0,
  double speed = 5.0,
  required DateTime timestamp,
}) =>
    LocationUpdate(
      latitude: lat,
      longitude: lon,
      accuracyMeters: accuracy,
      timestamp: timestamp,
      altitude: 0,
      speed: speed,
      heading: 0,
    );

// Approximate degree offsets for small distances (latitude only, cos(lat)≈1
// correction not needed for the precision we assert on).
//   1 degree latitude ≈ 111 000 m
const _stepLat5 = 5.0 / 111000.0; //  ~5 m per step
const _stepLat10 = 10.0 / 111000.0; // ~10 m per step
const _baseLat = 10.0;
const _baseLon = 106.0;
final _t0 = DateTime(2024, 1, 1, 12);

// ─── Tests ───────────────────────────────────────────────────────────────────

void main() {
  group('GPS filter — accuracy > 20 m', () {
    test('samples with accuracy > 20 m are discarded entirely', () {
      final engine = TripMetricsEngine(locationStream: _neverStream());
      engine.start();

      // All 10 samples have accuracy = 25 m → all rejected.
      for (int i = 0; i < 10; i++) {
        engine.addLocation(_fix(
          lat: _baseLat + i * _stepLat10,
          lon: _baseLon,
          accuracy: 25,
          timestamp: _t0.add(Duration(seconds: i * 2)),
        ));
      }

      final m = engine.finish();
      expect(m.totalDistanceMeters, 0.0);
      expect(m.gpsSampleCount, 0);
    });

    test('only good-accuracy samples are counted', () {
      final engine = TripMetricsEngine(locationStream: _neverStream());
      engine.start();

      // Sample 1: bad accuracy → rejected; no baseline established.
      engine.addLocation(_fix(
        lat: _baseLat,
        lon: _baseLon,
        accuracy: 30,
        timestamp: _t0,
      ));
      // Sample 2: good accuracy → baseline.
      engine.addLocation(_fix(
        lat: _baseLat,
        lon: _baseLon,
        accuracy: 10,
        timestamp: _t0.add(const Duration(seconds: 2)),
      ));
      // Sample 3: good, 10 m ahead.
      engine.addLocation(_fix(
        lat: _baseLat + _stepLat10,
        lon: _baseLon,
        accuracy: 10,
        timestamp: _t0.add(const Duration(seconds: 4)),
      ));

      final m = engine.finish();
      expect(m.totalDistanceMeters, closeTo(10.0, 1.0));
      expect(m.gpsSampleCount, 2); // samples 2 and 3
    });
  });

  group('GPS filter — movement < 5 m (drift / standing still)', () {
    test('GPS drift ±2 m around a fixed point: all samples after baseline rejected',
        () {
      final engine = TripMetricsEngine(locationStream: _neverStream());
      engine.start();
      const driftLat = 2.0 / 111000.0; // ~2 m

      for (int i = 0; i < 20; i++) {
        // Alternate between +2 m and −2 m. Each hop from the last-accepted
        // fix is ≤ 4 m — below the 5 m threshold.
        final lat = _baseLat + (i.isEven ? driftLat : -driftLat);
        engine.addLocation(_fix(
          lat: lat,
          lon: _baseLon,
          timestamp: _t0.add(Duration(seconds: i * 2)),
        ));
      }

      final m = engine.finish();
      expect(m.totalDistanceMeters, 0.0);
    });

    test('duplicate coordinates are discarded', () {
      final engine = TripMetricsEngine(locationStream: _neverStream());
      engine.start();

      for (int i = 0; i < 10; i++) {
        engine.addLocation(_fix(
          lat: _baseLat,
          lon: _baseLon,
          timestamp: _t0.add(Duration(seconds: i * 10)),
        ));
      }

      final m = engine.finish();
      expect(m.totalDistanceMeters, 0.0);
      expect(m.gpsSampleCount, 1); // only the initial baseline fix
    });
  });

  group('GPS filter — impossible speed > 50 m/s', () {
    test('GPS-reported speed > 50 m/s rejects that sample', () {
      final engine = TripMetricsEngine(locationStream: _neverStream());
      engine.start();

      // Sample 1: baseline, good speed.
      engine.addLocation(_fix(
        lat: _baseLat,
        lon: _baseLon,
        speed: 10,
        timestamp: _t0,
      ));
      // Sample 2: impossible GPS speed (> 50 m/s) → rejected even though
      // position moved; last-accepted stays at sample 1.
      engine.addLocation(_fix(
        lat: _baseLat + _stepLat10,
        lon: _baseLon,
        speed: 100,
        timestamp: _t0.add(const Duration(seconds: 2)),
      ));
      // Sample 3: reasonable speed, 20 m from baseline (sample 1).
      // elapsed from sample 1 = 4 s → implied speed = 20/4 = 5 m/s → accepted.
      engine.addLocation(_fix(
        lat: _baseLat + 2 * _stepLat10,
        lon: _baseLon,
        speed: 10,
        timestamp: _t0.add(const Duration(seconds: 4)),
      ));

      final m = engine.finish();
      expect(m.gpsSampleCount, 2); // samples 1 and 3
      expect(m.totalDistanceMeters, closeTo(20.0, 2.0));
    });
  });

  group('Distance accumulation', () {
    test('walking ~100 m: 20 samples × 5 m → ~95 m total (19 intervals)', () {
      final engine = TripMetricsEngine(locationStream: _neverStream());
      engine.start();

      for (int i = 0; i < 20; i++) {
        engine.addLocation(_fix(
          lat: _baseLat + i * _stepLat5,
          lon: _baseLon,
          speed: 2.5,
          timestamp: _t0.add(Duration(seconds: i * 2)),
        ));
      }

      final m = engine.finish();
      // First sample sets baseline; 19 intervals × ~5 m ≈ 95 m.
      expect(m.totalDistanceMeters, closeTo(95.0, 2.0));
      expect(m.gpsSampleCount, 20);
    });

    test('driving ~2 km: 200 samples × 10 m → ~1990 m total (199 intervals)',
        () {
      final engine = TripMetricsEngine(locationStream: _neverStream());
      engine.start();

      for (int i = 0; i < 200; i++) {
        engine.addLocation(_fix(
          lat: _baseLat + i * _stepLat10,
          lon: _baseLon,
          speed: 10.0,
          timestamp: _t0.add(Duration(seconds: i)),
        ));
      }

      final m = engine.finish();
      // 199 intervals × ~10 m ≈ 1990 m.
      expect(m.totalDistanceMeters, closeTo(1990.0, 5.0));
      expect(m.gpsSampleCount, 200);
    });
  });

  group('Engine API — reset()', () {
    test('reset clears accumulated state so a fresh trip starts from zero', () {
      final engine = TripMetricsEngine(locationStream: _neverStream());
      engine.start();

      engine.addLocation(_fix(lat: _baseLat, lon: _baseLon, timestamp: _t0));
      engine.addLocation(_fix(
        lat: _baseLat + _stepLat10,
        lon: _baseLon,
        timestamp: _t0.add(const Duration(seconds: 2)),
      ));

      engine.reset();

      final m = engine.metrics;
      expect(m.totalDistanceMeters, 0.0);
      expect(m.gpsSampleCount, 0);
      expect(m.startedAt, isNull);
      expect(m.finishedAt, isNull);
    });
  });

  group('Engine API — finish() idempotency', () {
    test('calling finish() twice returns the same distance', () {
      final engine = TripMetricsEngine(locationStream: _neverStream());
      engine.start();

      engine.addLocation(_fix(lat: _baseLat, lon: _baseLon, timestamp: _t0));
      engine.addLocation(_fix(
        lat: _baseLat + _stepLat10,
        lon: _baseLon,
        timestamp: _t0.add(const Duration(seconds: 2)),
      ));

      final first = engine.finish();
      final second = engine.finish();

      expect(first.totalDistanceMeters,
          closeTo(second.totalDistanceMeters, 0.001));
      expect(first.finishedAt, second.finishedAt);
    });
  });

  group('distanceKm / durationMinutes derived helpers', () {
    test('distanceKm is totalDistanceMeters / 1000', () {
      final engine = TripMetricsEngine(locationStream: _neverStream());
      engine.start();

      for (int i = 0; i < 20; i++) {
        engine.addLocation(_fix(
          lat: _baseLat + i * _stepLat5,
          lon: _baseLon,
          timestamp: _t0.add(Duration(seconds: i * 2)),
        ));
      }

      final m = engine.finish();
      expect(m.distanceKm, closeTo(m.totalDistanceMeters / 1000.0, 0.0001));
    });
  });
}
