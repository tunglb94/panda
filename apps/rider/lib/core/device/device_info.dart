import 'dart:io';
import 'dart:math';

import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../config/app_config.dart';

/// Best-effort device metadata sent alongside login (Device & Security
/// phase) — the backend upserts a `user_devices` row and appends a
/// `login_history` row from these fields. None of this blocks or fails
/// login if collection fails; every getter has a safe fallback.
abstract final class DeviceInfo {
  static const _storage = FlutterSecureStorage();
  static const _keyDeviceId = 'device_id';

  /// A random ID generated once per install and persisted in secure
  /// storage — stable across app restarts, distinct across devices.
  static Future<String> deviceId() async {
    final existing = await _storage.read(key: _keyDeviceId);
    if (existing != null && existing.isNotEmpty) return existing;
    final generated = _randomHex(16);
    await _storage.write(key: _keyDeviceId, value: generated);
    return generated;
  }

  static String platform() {
    try {
      if (Platform.isAndroid) return 'android';
      if (Platform.isIOS) return 'ios';
      if (Platform.isWindows) return 'windows';
      if (Platform.isMacOS) return 'macos';
      if (Platform.isLinux) return 'linux';
    } catch (_) {
      // Platform.* throws on web — DeviceInfo.model() below returns ''
      // in that case too. Known Gap: no web build target today anyway.
    }
    return 'unknown';
  }

  /// Best-effort stand-in for a real device model name — this is the OS
  /// version string, not "Pixel 8"/"iPhone 15" etc. Getting the real model
  /// needs the `device_info_plus` plugin, deliberately not added here (see
  /// plan's Known Gaps: avoid a new native-plugin dependency for a
  /// telemetry-only field).
  static String model() {
    try {
      return Platform.operatingSystemVersion;
    } catch (_) {
      return '';
    }
  }

  static String appVersion() => AppConfig.appVersion;

  /// No push-notification integration yet (see plan's Known Gaps) — always
  /// empty until FCM is wired up.
  static String fcmToken() => '';

  static String _randomHex(int bytes) {
    final rand = Random.secure();
    final buffer = StringBuffer();
    for (var i = 0; i < bytes; i++) {
      buffer.write(rand.nextInt(256).toRadixString(16).padLeft(2, '0'));
    }
    return buffer.toString();
  }
}
