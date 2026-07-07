import 'dart:async';

import '../../../core/location/location.dart';
import '../../../core/location/location_engine_config.dart';
import '../data/location_upload_repository.dart';

enum UploadStatus { idle, uploading, success, failed }

/// Continuously obtains GPS location from [LocationEngine] and uploads it to
/// the backend at [uploadInterval] intervals.
///
/// - Starts / stops the [LocationEngine] alongside itself.
/// - Uses simple exponential backoff (up to 3 retries) on upload failure.
/// - Exposes [statusStream] for debug display in the UI.
/// - Ignored insignificant movements via [LocationEngineConfig.distanceFilter].
class LocationUploadService {
  LocationUploadService({
    required LocationUploadRepository repository,
    this.uploadInterval = const Duration(seconds: 15),
  }) : _repo = repository;

  final LocationUploadRepository _repo;
  final Duration uploadInterval;

  final _engine = LocationEngine(
    config: const LocationEngineConfig(distanceFilter: 10),
  );
  final _statusCtrl = StreamController<UploadStatus>.broadcast();

  StreamSubscription<LocationUpdate>? _locationSub;
  Timer? _uploadTimer;
  LocationUpdate? _latestLocation;
  UploadStatus _status = UploadStatus.idle;

  Stream<UploadStatus> get statusStream => _statusCtrl.stream;
  UploadStatus get status => _status;

  bool get isRunning => _locationSub != null;

  /// Starts the GPS engine and the periodic upload timer.
  /// No-op if already running.
  Future<void> start() async {
    if (isRunning) return;
    await _engine.start();
    _locationSub = _engine.locationStream.listen(_onLocation);
    _uploadTimer = Timer.periodic(uploadInterval, (_) => _doUpload());
  }

  /// Stops uploads and the GPS engine.
  void stop() {
    _uploadTimer?.cancel();
    _uploadTimer = null;
    _locationSub?.cancel();
    _locationSub = null;
    _engine.stop();
    _setStatus(UploadStatus.idle);
  }

  void dispose() {
    stop();
    _engine.dispose();
    if (!_statusCtrl.isClosed) _statusCtrl.close();
  }

  void _onLocation(LocationUpdate update) {
    _latestLocation = update;
  }

  Future<void> _doUpload() async {
    final loc = _latestLocation;
    if (loc == null) return;
    _setStatus(UploadStatus.uploading);
    // Exponential backoff: 3 attempts, delays of 2s and 4s between retries.
    for (int attempt = 0; attempt < 3; attempt++) {
      try {
        await _repo.uploadLocation(loc.latitude, loc.longitude);
        _setStatus(UploadStatus.success);
        return;
      } catch (_) {
        if (attempt < 2) {
          await Future.delayed(Duration(seconds: 1 << (attempt + 1)));
        }
      }
    }
    _setStatus(UploadStatus.failed);
  }

  void _setStatus(UploadStatus s) {
    _status = s;
    if (!_statusCtrl.isClosed) _statusCtrl.add(s);
  }
}
