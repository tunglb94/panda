import 'package:flutter/foundation.dart';

/// Build mode of the running app instance.
enum BuildModeKind { debug, profile, release }

extension BuildModeKindX on BuildModeKind {
  String get label => switch (this) {
        BuildModeKind.debug => 'Gỡ lỗi',
        BuildModeKind.profile => 'Hiệu năng',
        BuildModeKind.release => 'Phát hành',
      };
}

/// Diagnostic info shown on the Developer page.
///
/// [appVersion] is a mock string (no build pipeline injects the real
/// version yet). [buildMode] is read for real from Flutter's compile-time
/// constants — not mocked. [flutterVersionPlaceholder] and
/// [environmentPlaceholder] are explicit placeholders: nothing in this
/// project currently captures the Flutter SDK version or environment name
/// at build time (see `docs/project/MVP_DEVELOPMENT_PLAN.md` Backend
/// Roadmap stage B10 — CI hardening — for where that would eventually be
/// wired up).
class AppInfo {
  const AppInfo({
    required this.appVersion,
    required this.buildMode,
    required this.flutterVersionPlaceholder,
    required this.environmentPlaceholder,
  });

  final String appVersion;
  final BuildModeKind buildMode;
  final String flutterVersionPlaceholder;
  final String environmentPlaceholder;

  static BuildModeKind currentBuildMode() {
    if (kReleaseMode) return BuildModeKind.release;
    if (kProfileMode) return BuildModeKind.profile;
    return BuildModeKind.debug;
  }
}
