import 'package:flutter/material.dart';

import '../../shared/widgets/app_loading_view.dart';
import '../auth/auth_state.dart';
import '../config/app_config.dart';
import '../kyc/kyc_gate.dart';
import '../network/api_client.dart';
import '../theme/app_colors.dart';
import '../theme/app_spacing.dart';
import '../version/app_version_checker.dart';

/// APP STARTUP state machine: Splash → Check App Version (this page, first)
/// → Check Token (already resolved by [AuthState.initialize] before
/// runApp) → Refresh Token (transparent, inside [ApiClient]) → Load
/// Profile / Check KYC (this page triggers [KycGate.refresh]) → route.
/// This page never navigates itself — go_router's `redirect`, listening to
/// both [AuthState] and [KycGate], recomputes the destination once
/// [KycGate] resolves and pushes this widget off the stack. While a
/// blocking version check is pending/failed, the page stays put on
/// purpose — the redirect logic never moves off Splash while
/// `kycGate.approved` is still null.
class SplashPage extends StatefulWidget {
  const SplashPage({
    super.key,
    required this.authState,
    required this.kycGate,
    required this.apiClient,
  });

  final AuthState authState;
  final KycGate kycGate;
  final ApiClient apiClient;

  @override
  State<SplashPage> createState() => _SplashPageState();
}

class _SplashPageState extends State<SplashPage> {
  AppVersionInfo? _blockingVersionInfo;

  @override
  void initState() {
    super.initState();
    _checkVersionThenProceed();
  }

  Future<void> _checkVersionThenProceed() async {
    try {
      final info = await fetchAppVersionInfo(widget.apiClient, 'driver');
      if (!mounted) return;
      if (info.isBlocked(AppConfig.appVersion)) {
        setState(() => _blockingVersionInfo = info);
        return;
      }
    } catch (_) {
      // Fail open — a transient version-check failure (network hiccup,
      // gateway restart) must never lock a user out of the app.
    }
    if (widget.authState.isLoggedIn && widget.kycGate.approved == null) {
      widget.kycGate.refresh(widget.apiClient);
    }
  }

  @override
  Widget build(BuildContext context) {
    final blocking = _blockingVersionInfo;
    if (blocking != null) {
      return _ForceUpdateView(info: blocking);
    }
    return const Scaffold(
      backgroundColor: AppColors.surface,
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.local_shipping_outlined, size: 72, color: AppColors.primary),
            SizedBox(height: 24),
            AppLoadingView(),
          ],
        ),
      ),
    );
  }
}

/// No real app-store deep link is configured yet (see plan's Known Gaps) —
/// this is a hard, non-dismissible block with instructions only.
class _ForceUpdateView extends StatelessWidget {
  const _ForceUpdateView({required this.info});

  final AppVersionInfo info;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.surface,
      body: SafeArea(
        child: Center(
          child: Padding(
            padding: const EdgeInsets.all(AppSpacing.xl),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(Icons.system_update_outlined, size: 64, color: AppColors.primary),
                const SizedBox(height: AppSpacing.lg),
                Text('Cần cập nhật ứng dụng', style: Theme.of(context).textTheme.titleLarge),
                const SizedBox(height: AppSpacing.sm),
                Text(
                  'Phiên bản hiện tại không còn được hỗ trợ. Vui lòng cập nhật lên phiên bản mới nhất (${info.latestVersion}) để tiếp tục sử dụng.',
                  textAlign: TextAlign.center,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
