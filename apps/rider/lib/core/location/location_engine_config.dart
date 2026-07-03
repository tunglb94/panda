import 'package:geolocator/geolocator.dart';

/// Configuration for [LocationEngine].
///
/// All fields are immutable. Use [copyWith] to derive a modified copy.
class LocationEngineConfig {
  const LocationEngineConfig({
    this.accuracy = LocationAccuracy.high,
    this.distanceFilter = 5.0,
    this.updateIntervalMs = 5000,
  });

  /// Desired accuracy of each GPS fix.
  final LocationAccuracy accuracy;

  /// Minimum movement in metres before a new fix is emitted.
  /// Applies on both Android and iOS.
  final double distanceFilter;

  /// Desired milliseconds between successive fixes (Android only).
  /// iOS uses [distanceFilter] to gate updates instead.
  final int updateIntervalMs;

  LocationEngineConfig copyWith({
    LocationAccuracy? accuracy,
    double? distanceFilter,
    int? updateIntervalMs,
  }) =>
      LocationEngineConfig(
        accuracy: accuracy ?? this.accuracy,
        distanceFilter: distanceFilter ?? this.distanceFilter,
        updateIntervalMs: updateIntervalMs ?? this.updateIntervalMs,
      );

  @override
  String toString() =>
      'LocationEngineConfig(accuracy: $accuracy, '
      'distanceFilter: ${distanceFilter}m, '
      'updateIntervalMs: ${updateIntervalMs}ms)';
}
