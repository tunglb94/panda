import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/features/chat/presentation/pages/chat_page.dart';
import 'package:rider/features/contact/data/contact_repository.dart';
import 'package:rider/shared/widgets/app_badge.dart';
import 'package:rider/shared/widgets/app_button.dart';
import 'package:rider/shared/widgets/app_snackbar.dart';

/// Real Call + Chat actions for the assigned driver (Part 1 & 2) —
/// replaces the former "coming soon" placeholder. Only [tripId] is needed:
/// the backend resolves which participant is calling/chatting from the JWT
/// + trip lookup, so no driverId has to be threaded down to this widget.
/// [unreadCount] (Part 5) shows a badge on the Chat button when non-zero —
/// a point-in-time snapshot fetched by the parent page, not live-updated
/// while this button is on screen (opening the chat itself always shows
/// the true state).
class ContactDriverButton extends StatefulWidget {
  const ContactDriverButton({
    super.key,
    required this.tripId,
    required this.apiClient,
    this.unreadCount = 0,
  });

  final String tripId;
  final ApiClient apiClient;
  final int unreadCount;

  @override
  State<ContactDriverButton> createState() => _ContactDriverButtonState();
}

class _ContactDriverButtonState extends State<ContactDriverButton> {
  bool _calling = false;

  Future<void> _call() async {
    if (_calling) return;
    setState(() => _calling = true);
    try {
      final phone = await ContactRepository(widget.apiClient).call(widget.tripId);
      if (phone.isEmpty) {
        if (mounted) AppSnackbar.error(context, 'Không thể lấy số điện thoại tài xế.');
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
        AppSnackbar.error(context, e.statusCode == 0 ? e.message : 'Không thể gọi tài xế lúc này.');
      }
    } catch (_) {
      if (mounted) AppSnackbar.error(context, 'Không thể gọi tài xế lúc này.');
    } finally {
      if (mounted) setState(() => _calling = false);
    }
  }

  void _openChat() {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => ChatPage(tripId: widget.tripId, apiClient: widget.apiClient),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(
          child: AppButton.outline(
            label: 'Gọi',
            icon: Icons.call,
            isLoading: _calling,
            onPressed: _calling ? null : _call,
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Stack(
            clipBehavior: Clip.none,
            children: [
              AppButton.outline(
                label: 'Nhắn tin',
                icon: Icons.chat_bubble_outline,
                onPressed: _openChat,
              ),
              if (widget.unreadCount > 0)
                Positioned(
                  right: 4,
                  top: -4,
                  child: AppBadge(count: widget.unreadCount),
                ),
            ],
          ),
        ),
      ],
    );
  }
}
