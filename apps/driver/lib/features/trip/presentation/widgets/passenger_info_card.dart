import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../../../shared/widgets/app_snackbar.dart';
import '../../../../shared/widgets/pressable_scale.dart';

/// Passenger identity row shown on the offer, active-trip, and
/// awaiting-payment cards — mirrors Uber Driver's rider-info row (avatar +
/// name + quick contact actions).
///
/// The backend's trip/offer payloads carry a rider ID but not a display
/// name, photo, or rating (see `docs/driver/DRIVER_APP_SPEC.md` §15.5 — the
/// richer `RiderInfo` model with name+rating only exists in the disconnected
/// preview module today). Rather than fabricate a name or a star rating,
/// this shows the honest generic label "Hành khách" with no invented data —
/// consistent with this codebase's established convention (see the rider
/// app's own placeholder actions) of never presenting mock data as if it
/// were real.
class PassengerInfoCard extends StatelessWidget {
  const PassengerInfoCard({super.key});

  @override
  Widget build(BuildContext context) {
    return AppCard(
      padding: const EdgeInsets.symmetric(
        horizontal: AppSpacing.lg,
        vertical: AppSpacing.md,
      ),
      child: Row(
        children: [
          const CircleAvatar(
            radius: 22,
            backgroundColor: AppColors.primaryLight,
            child: Icon(Icons.person, color: AppColors.primary, size: AppIconSize.lg),
          ),
          const SizedBox(width: AppSpacing.md),
          Expanded(
            child: Text('Hành khách', style: Theme.of(context).textTheme.titleSmall),
          ),
          _ContactIconButton(
            icon: Icons.call,
            tooltip: 'Gọi hành khách',
            onTap: () => _showPlaceholder(context, 'gọi điện'),
          ),
          const SizedBox(width: AppSpacing.sm),
          _ContactIconButton(
            icon: Icons.chat_bubble_outline,
            tooltip: 'Nhắn tin',
            onTap: () => _showPlaceholder(context, 'nhắn tin'),
          ),
        ],
      ),
    );
  }

  void _showPlaceholder(BuildContext context, String action) {
    AppSnackbar.show(
      context,
      'Tính năng $action chỉ là giao diện mẫu — chưa được kết nối với backend.',
    );
  }
}

class _ContactIconButton extends StatefulWidget {
  const _ContactIconButton({required this.icon, required this.tooltip, required this.onTap});

  final IconData icon;
  final String tooltip;
  final VoidCallback onTap;

  @override
  State<_ContactIconButton> createState() => _ContactIconButtonState();
}

class _ContactIconButtonState extends State<_ContactIconButton> {
  bool _pressed = false;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: AppColors.primaryLight,
      shape: const CircleBorder(),
      child: InkWell(
        onTap: widget.onTap,
        customBorder: const CircleBorder(),
        onHighlightChanged: (v) => setState(() => _pressed = v),
        child: Tooltip(
          message: widget.tooltip,
          // 48dp minimum touch target (14 padding + 20 icon + 14 padding).
          child: PressableScale(
            pressed: _pressed,
            child: Padding(
              padding: const EdgeInsets.all(14),
              child: Icon(widget.icon, size: AppIconSize.md, color: AppColors.primary),
            ),
          ),
        ),
      ),
    );
  }
}
