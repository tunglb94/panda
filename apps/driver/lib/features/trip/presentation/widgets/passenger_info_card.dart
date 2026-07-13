import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../../core/network/api_client.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../features/chat/presentation/pages/chat_page.dart';
import '../../../../features/contact/data/contact_repository.dart';
import '../../../../features/contact/domain/models/contact_info.dart';
import '../../../../shared/widgets/app_badge.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../../../shared/widgets/app_snackbar.dart';
import '../../../../shared/widgets/pressable_scale.dart';

/// Passenger identity row shown on the offer, active-trip, and
/// awaiting-payment cards — mirrors Uber Driver's rider-info row (avatar +
/// name + quick contact actions).
///
/// Real Call/Chat (Part 1 & 2) only activate when both [tripId] and
/// [apiClient] are supplied — the backend requires the trip to already have
/// a driver assigned (`trip.driver_id != ""`) before Contact/Chat/Call are
/// valid, which is only true once this driver has accepted the offer. The
/// pre-accept `_OfferCard` therefore keeps passing neither, and this card
/// falls back to its original "coming soon" placeholder there — not a
/// regression, since there was never a real participant relationship to
/// call/chat with at that stage anyway.
class PassengerInfoCard extends StatefulWidget {
  const PassengerInfoCard({
    super.key,
    this.tripId,
    this.apiClient,
    this.unreadCount = 0,
  });

  final String? tripId;
  final ApiClient? apiClient;
  final int unreadCount;

  bool get _isLive => tripId != null && apiClient != null;

  @override
  State<PassengerInfoCard> createState() => _PassengerInfoCardState();
}

class _PassengerInfoCardState extends State<PassengerInfoCard> {
  ContactInfo? _contact;
  bool _calling = false;

  @override
  void initState() {
    super.initState();
    if (widget._isLive) _fetchContact();
  }

  Future<void> _fetchContact() async {
    try {
      final contact = await ContactRepository(widget.apiClient!).getContact(widget.tripId!);
      if (mounted) setState(() => _contact = contact);
    } catch (_) {
      // Non-fatal — card just shows the generic "Hành khách" label.
    }
  }

  Future<void> _call() async {
    if (!widget._isLive) {
      _showPlaceholder('gọi điện');
      return;
    }
    if (_calling) return;
    setState(() => _calling = true);
    try {
      final phone = await ContactRepository(widget.apiClient!).call(widget.tripId!);
      if (phone.isEmpty) {
        if (mounted) AppSnackbar.error(context, 'Không thể lấy số điện thoại hành khách.');
        return;
      }
      final uri = Uri(scheme: 'tel', path: phone);
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri);
      } else if (mounted) {
        AppSnackbar.error(context, 'Không thể mở ứng dụng gọi điện trên thiết bị này.');
      }
    } on ApiException catch (e) {
      if (mounted) {
        AppSnackbar.error(context, e.statusCode == 0 ? e.message : 'Không thể gọi hành khách lúc này.');
      }
    } catch (_) {
      if (mounted) AppSnackbar.error(context, 'Không thể gọi hành khách lúc này.');
    } finally {
      if (mounted) setState(() => _calling = false);
    }
  }

  void _openChat() {
    if (!widget._isLive) {
      _showPlaceholder('nhắn tin');
      return;
    }
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => ChatPage(tripId: widget.tripId!, apiClient: widget.apiClient!),
      ),
    );
  }

  void _showPlaceholder(String action) {
    AppSnackbar.show(
      context,
      'Tính năng $action chỉ khả dụng sau khi bạn đã nhận chuyến.',
    );
  }

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
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  _contact?.name.isNotEmpty == true ? _contact!.name : 'Hành khách',
                  style: Theme.of(context).textTheme.titleSmall,
                ),
                if (_contact != null && _contact!.hasRating)
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(Icons.star, size: 12, color: Colors.amber),
                      const SizedBox(width: 2),
                      Text(
                        _contact!.rating.toStringAsFixed(1),
                        style: Theme.of(context).textTheme.labelSmall,
                      ),
                    ],
                  ),
              ],
            ),
          ),
          _ContactIconButton(
            icon: Icons.call,
            tooltip: 'Gọi hành khách',
            isLoading: _calling,
            onTap: _call,
          ),
          const SizedBox(width: AppSpacing.sm),
          Stack(
            clipBehavior: Clip.none,
            children: [
              _ContactIconButton(
                icon: Icons.chat_bubble_outline,
                tooltip: 'Nhắn tin',
                onTap: _openChat,
              ),
              if (widget.unreadCount > 0)
                Positioned(
                  right: -2,
                  top: -2,
                  child: AppBadge(count: widget.unreadCount),
                ),
            ],
          ),
        ],
      ),
    );
  }
}

class _ContactIconButton extends StatefulWidget {
  const _ContactIconButton({
    required this.icon,
    required this.tooltip,
    required this.onTap,
    this.isLoading = false,
  });

  final IconData icon;
  final String tooltip;
  final VoidCallback onTap;
  final bool isLoading;

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
        onTap: widget.isLoading ? null : widget.onTap,
        customBorder: const CircleBorder(),
        onHighlightChanged: (v) => setState(() => _pressed = v),
        child: Tooltip(
          message: widget.tooltip,
          // 48dp minimum touch target (14 padding + 20 icon + 14 padding).
          child: PressableScale(
            pressed: _pressed,
            child: Padding(
              padding: const EdgeInsets.all(14),
              child: widget.isLoading
                  ? const SizedBox(
                      width: AppIconSize.md,
                      height: AppIconSize.md,
                      child: CircularProgressIndicator(strokeWidth: 2, color: AppColors.primary),
                    )
                  : Icon(widget.icon, size: AppIconSize.md, color: AppColors.primary),
            ),
          ),
        ),
      ),
    );
  }
}
