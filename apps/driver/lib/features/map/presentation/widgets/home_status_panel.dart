import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_radius.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_button.dart';
import '../../../../shared/widgets/app_status_chip.dart';

/// The 8 states the Home bottom panel can be in. Purely presentational —
/// [MapPage] computes this by *reading* the same `TripOfferRepository`/
/// `ActiveTripRepository` that `TripPage` already owns (no new endpoints,
/// no new mutating calls). Accepting/rejecting an offer and every trip
/// action (arrive/start/finish) still only happen on the Trips tab, which
/// remains the single source of truth — Home is a live mirror, not a
/// second control surface, so there is exactly one place in the app that
/// can mutate trip state.
enum HomePhase {
  offline,
  online,
  incomingTrip,
  pickingUp,
  waiting,
  inTrip,
  awaitingPayment,
  completed,
}

class HomeStatusPanel extends StatelessWidget {
  const HomeStatusPanel({
    super.key,
    required this.phase,
    required this.isBusy,
    this.error,
    this.pickupAddress,
    this.dropoffAddress,
    this.countdownSeconds,
    this.fareLabel,
    this.canGoOnline = true,
    required this.onToggleOnline,
    required this.onViewTrip,
  });

  final HomePhase phase;

  /// True while the online/offline toggle request is in flight — the one
  /// piece of state that *can* mutate from Home, since it's Home's own
  /// control, not trip-flow state.
  final bool isBusy;
  final String? error;
  final String? pickupAddress;
  final String? dropoffAddress;
  final int? countdownSeconds;
  final String? fareLabel;

  /// Phần 7 (Driver KYC) — false when Driver KYC + Vehicle Verification
  /// aren't both Approved yet. Only gates going *online*: a driver who
  /// somehow already has an active session can always go offline.
  final bool canGoOnline;
  final VoidCallback onToggleOnline;

  /// Navigates to the Trips tab, where the real action buttons for the
  /// current trip phase live.
  final VoidCallback onViewTrip;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: const BoxDecoration(
        color: AppColors.surface,
        borderRadius: AppRadius.topXl,
        boxShadow: [
          BoxShadow(color: Color(0x1A000000), blurRadius: 20, offset: Offset(0, -4)),
        ],
      ),
      child: SafeArea(
        top: false,
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: Padding(
              padding: const EdgeInsets.fromLTRB(
                AppSpacing.xl,
                AppSpacing.lg,
                AppSpacing.xl,
                AppSpacing.md,
              ),
              child: AnimatedSwitcher(
                duration: const Duration(milliseconds: 280),
                switchInCurve: Curves.easeOut,
                switchOutCurve: Curves.easeIn,
                transitionBuilder: (child, animation) => FadeTransition(
                  opacity: animation,
                  child: SizeTransition(
                    sizeFactor: animation,
                    axisAlignment: -1,
                    child: child,
                  ),
                ),
                child: Column(
                  key: ValueKey(phase),
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    _buildContent(context),
                    if (error != null) ...[
                      const SizedBox(height: AppSpacing.xs),
                      Text(
                        error!,
                        style: const TextStyle(color: AppColors.error, fontSize: 12),
                      ),
                    ],
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildContent(BuildContext context) {
    return switch (phase) {
      HomePhase.offline => _AvailabilityRow(
          isOnline: false,
          isBusy: isBusy,
          canToggle: canGoOnline,
          onToggle: onToggleOnline,
        ),
      HomePhase.online => Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            _AvailabilityRow(isOnline: true, isBusy: isBusy, canToggle: true, onToggle: onToggleOnline),
            const SizedBox(height: AppSpacing.md),
            const _SearchingRow(),
          ],
        ),
      HomePhase.incomingTrip => _TripPreviewCard(
          badgeLabel: countdownSeconds != null ? '${countdownSeconds}s' : 'Mới',
          badgeColor: (countdownSeconds ?? 99) <= 10 ? AppColors.error : AppColors.warning,
          title: 'Yêu cầu chuyến mới',
          pickupAddress: pickupAddress,
          dropoffAddress: dropoffAddress,
          ctaLabel: 'Xem chi tiết',
          onCta: onViewTrip,
          highlighted: true,
        ),
      HomePhase.pickingUp => _TripPreviewCard(
          badgeLabel: 'Đang đến điểm đón',
          badgeColor: AppColors.warning,
          title: 'Đang đến đón khách',
          pickupAddress: pickupAddress,
          dropoffAddress: dropoffAddress,
          ctaLabel: 'Xem chuyến đi',
          onCta: onViewTrip,
        ),
      HomePhase.waiting => _TripPreviewCard(
          badgeLabel: 'Đã đến điểm đón',
          badgeColor: AppColors.primary,
          title: 'Đang chờ khách',
          pickupAddress: pickupAddress,
          dropoffAddress: dropoffAddress,
          ctaLabel: 'Xem chuyến đi',
          onCta: onViewTrip,
        ),
      HomePhase.inTrip => _TripPreviewCard(
          badgeLabel: 'Đang thực hiện',
          badgeColor: AppColors.info,
          title: 'Đang trong chuyến đi',
          pickupAddress: pickupAddress,
          dropoffAddress: dropoffAddress,
          ctaLabel: 'Xem chuyến đi',
          onCta: onViewTrip,
        ),
      HomePhase.awaitingPayment => _TripPreviewCard(
          badgeLabel: 'Chờ thanh toán',
          badgeColor: AppColors.warning,
          title: 'Đang chờ khách thanh toán',
          pickupAddress: pickupAddress,
          dropoffAddress: dropoffAddress,
          fareLabel: fareLabel,
          ctaLabel: 'Xem chi tiết',
          onCta: onViewTrip,
        ),
      HomePhase.completed => const _CompletedRow(),
    };
  }
}

// ─── Offline / Online / Searching ─────────────────────────────────────────────

class _AvailabilityRow extends StatelessWidget {
  const _AvailabilityRow({
    required this.isOnline,
    required this.isBusy,
    required this.onToggle,
    this.canToggle = true,
  });

  final bool isOnline;
  final bool isBusy;
  final bool canToggle;
  final VoidCallback onToggle;

  @override
  Widget build(BuildContext context) {
    final blocked = !isOnline && !canToggle;
    return Row(
      children: [
        if (isBusy)
          const SizedBox(
            width: 12,
            height: 12,
            child: CircularProgressIndicator(strokeWidth: 2),
          )
        else
          AnimatedContainer(
            duration: const Duration(milliseconds: 300),
            width: 12,
            height: 12,
            decoration: BoxDecoration(
              color: isOnline ? AppColors.primary : AppColors.textTertiary,
              shape: BoxShape.circle,
              boxShadow: isOnline
                  ? [BoxShadow(color: AppColors.primary.withValues(alpha: 0.4), blurRadius: 8)]
                  : null,
            ),
          ),
        const SizedBox(width: AppSpacing.md),
        Expanded(
          child: Text(
            isOnline ? 'Bạn đang online' : 'Bạn đang offline',
            style: Theme.of(context).textTheme.titleMedium,
          ),
        ),
        // Phần 7: disabled + tooltip when Driver KYC / Vehicle Verification
        // aren't both Approved yet — Semantics carries the same reason to
        // screen readers, not just the visual Tooltip.
        Tooltip(
          message: blocked ? 'Cần hoàn thành xác minh.' : (isOnline ? 'Chuyển sang offline' : 'Chuyển sang online'),
          child: Semantics(
            label: blocked ? 'Không thể bật online — cần hoàn thành xác minh' : null,
            child: Switch(
              value: isOnline,
              onChanged: (isBusy || blocked) ? null : (_) => onToggle(),
            ),
          ),
        ),
      ],
    );
  }
}

class _SearchingRow extends StatefulWidget {
  const _SearchingRow();

  @override
  State<_SearchingRow> createState() => _SearchingRowState();
}

class _SearchingRowState extends State<_SearchingRow>
    with SingleTickerProviderStateMixin {
  late final AnimationController _pulse;

  @override
  void initState() {
    super.initState();
    _pulse = AnimationController(vsync: this, duration: const Duration(seconds: 2))
      ..repeat(reverse: true);
  }

  @override
  void dispose() {
    _pulse.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        ScaleTransition(
          scale: Tween(begin: 0.85, end: 1.15)
              .animate(CurvedAnimation(parent: _pulse, curve: Curves.easeInOut)),
          child: const Icon(Icons.radar, size: AppIconSize.lg, color: AppColors.primary),
        ),
        const SizedBox(width: AppSpacing.md),
        Expanded(
          child: Text(
            'Đang tìm chuyến gần bạn…',
            style: Theme.of(context)
                .textTheme
                .bodyMedium
                ?.copyWith(color: AppColors.textSecondary),
          ),
        ),
      ],
    );
  }
}

// ─── Trip lifecycle preview ────────────────────────────────────────────────────

class _TripPreviewCard extends StatelessWidget {
  const _TripPreviewCard({
    required this.badgeLabel,
    required this.badgeColor,
    required this.title,
    required this.ctaLabel,
    required this.onCta,
    this.pickupAddress,
    this.dropoffAddress,
    this.fareLabel,
    this.highlighted = false,
  });

  final String badgeLabel;
  final Color badgeColor;
  final String title;
  final String? pickupAddress;
  final String? dropoffAddress;
  final String? fareLabel;
  final String ctaLabel;
  final VoidCallback onCta;
  final bool highlighted;

  @override
  Widget build(BuildContext context) {
    return AnimatedContainer(
      duration: const Duration(milliseconds: 300),
      padding: const EdgeInsets.all(AppSpacing.lg),
      decoration: BoxDecoration(
        color: highlighted ? AppColors.primaryLight : AppColors.surfaceAlt,
        borderRadius: AppRadius.lgAll,
        border: Border.all(
          color: highlighted ? AppColors.primary : AppColors.border,
          width: highlighted ? 1.5 : 1,
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Expanded(
                child: Text(title, style: Theme.of(context).textTheme.titleMedium),
              ),
              AppStatusChip(label: badgeLabel, color: badgeColor),
            ],
          ),
          if (pickupAddress != null || dropoffAddress != null) ...[
            const SizedBox(height: AppSpacing.md),
            if (pickupAddress != null)
              _MiniAddressRow(icon: Icons.location_on, color: AppColors.primary, address: pickupAddress!),
            if (pickupAddress != null && dropoffAddress != null)
              const SizedBox(height: AppSpacing.xs),
            if (dropoffAddress != null)
              _MiniAddressRow(icon: Icons.flag, color: AppColors.error, address: dropoffAddress!),
          ],
          if (fareLabel != null) ...[
            const SizedBox(height: AppSpacing.sm),
            Text(
              'Thu nhập dự kiến: $fareLabel',
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ],
          const SizedBox(height: AppSpacing.md),
          AppButton.primary(label: ctaLabel, onPressed: onCta, expand: true),
        ],
      ),
    );
  }
}

class _MiniAddressRow extends StatelessWidget {
  const _MiniAddressRow({required this.icon, required this.color, required this.address});

  final IconData icon;
  final Color color;
  final String address;

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, size: AppIconSize.sm, color: color),
        const SizedBox(width: 6),
        Expanded(
          child: Text(
            address,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppColors.textPrimary),
          ),
        ),
      ],
    );
  }
}

// ─── Completed (brief transitional flash) ──────────────────────────────────────

class _CompletedRow extends StatelessWidget {
  const _CompletedRow();

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Container(
          padding: const EdgeInsets.all(8),
          decoration: const BoxDecoration(color: AppColors.primaryLight, shape: BoxShape.circle),
          child: const Icon(Icons.check_circle, color: AppColors.primary, size: AppIconSize.lg),
        ),
        const SizedBox(width: AppSpacing.md),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Chuyến đi hoàn thành!', style: Theme.of(context).textTheme.titleMedium),
              Text(
                'Đang quay lại hàng đợi nhận chuyến…',
                style: Theme.of(context).textTheme.bodySmall,
              ),
            ],
          ),
        ),
      ],
    );
  }
}
