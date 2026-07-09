/// Immutable snapshot of all metrics collected during a trip.
///
/// Produced by [TripMetricsEngine] at any point during or after a trip.
/// All speed values are in metres-per-second; distances in metres.
class TripMetrics {
  const TripMetrics({
    this.totalDistanceMeters = 0,
    this.movingDistanceMeters = 0,
    this.idleDurationSeconds = 0,
    this.movingDurationSeconds = 0,
    this.averageSpeedMps = 0,
    this.maxSpeedMps = 0,
    this.gpsSampleCount = 0,
    this.startedAt,
    this.finishedAt,
  });

  /// Total GPS-filtered distance accumulated during the trip.
  final double totalDistanceMeters;

  /// Distance while the vehicle was in motion.
  /// Equal to [totalDistanceMeters] until map-matching is introduced.
  final double movingDistanceMeters;

  /// Seconds spent stationary (total elapsed – moving).
  final double idleDurationSeconds;

  /// Seconds during which the vehicle was moving between consecutive
  /// accepted GPS fixes.
  final double movingDurationSeconds;

  /// Average speed over the moving portion (m/s).
  final double averageSpeedMps;

  /// Highest GPS-reported speed observed (m/s).
  final double maxSpeedMps;

  /// Number of GPS samples that passed all quality filters.
  final int gpsSampleCount;

  final DateTime? startedAt;
  final DateTime? finishedAt;

  // ─── Derived helpers ──────────────────────────────────────────────────────

  /// Total elapsed seconds (wall-clock, start → finish).
  double get totalDurationSeconds {
    if (startedAt == null) return 0;
    final end = finishedAt ?? DateTime.now();
    return end.difference(startedAt!).inMilliseconds / 1000.0;
  }

  /// Distance in kilometres — used when reporting to the backend.
  double get distanceKm => totalDistanceMeters / 1000.0;

  /// Total elapsed duration in minutes — used when reporting to the backend.
  double get durationMinutes => totalDurationSeconds / 60.0;
}
