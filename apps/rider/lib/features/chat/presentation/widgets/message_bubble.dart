import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_radius.dart';
import 'package:rider/core/theme/app_spacing.dart';

import '../../domain/models/chat_message.dart';

/// One chat bubble. [isMine] controls alignment/color; [showRead] renders
/// "Đã xem" under only the single most recent message from me that the
/// other party has actually read (the caller decides which one that is).
class MessageBubble extends StatelessWidget {
  const MessageBubble({
    super.key,
    required this.message,
    required this.isMine,
    this.showRead = false,
  });

  final ChatMessage message;
  final bool isMine;
  final bool showRead;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final bubbleColor = isMine ? AppColors.primary : AppColors.surfaceAlt;
    final textColor = isMine ? AppColors.textOnPrimary : AppColors.textPrimary;

    return Semantics(
      label: '${isMine ? "Bạn" : "Tài xế"}: ${message.body}',
      child: Align(
        alignment: isMine ? Alignment.centerRight : Alignment.centerLeft,
        child: Column(
          crossAxisAlignment: isMine ? CrossAxisAlignment.end : CrossAxisAlignment.start,
          children: [
            ConstrainedBox(
              constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.75),
              child: AnimatedContainer(
                duration: const Duration(milliseconds: 180),
                curve: Curves.easeOut,
                margin: const EdgeInsets.symmetric(vertical: 3),
                padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: AppSpacing.sm),
                decoration: BoxDecoration(
                  color: bubbleColor,
                  borderRadius: BorderRadius.only(
                    topLeft: Radius.circular(AppRadius.lg),
                    topRight: Radius.circular(AppRadius.lg),
                    bottomLeft: Radius.circular(isMine ? AppRadius.lg : 0),
                    bottomRight: Radius.circular(isMine ? 0 : AppRadius.lg),
                  ),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Flexible(
                      child: Text(
                        message.body,
                        style: theme.textTheme.bodyMedium?.copyWith(color: textColor),
                      ),
                    ),
                    if (message.pending) ...[
                      const SizedBox(width: 6),
                      Icon(Icons.access_time, size: 12, color: textColor.withValues(alpha: 0.7)),
                    ],
                  ],
                ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 4),
              child: Text(
                showRead ? 'Đã xem' : _timeLabel(message.createdAt),
                style: theme.textTheme.labelSmall?.copyWith(color: AppColors.textTertiary, fontSize: 10),
              ),
            ),
          ],
        ),
      ),
    );
  }

  static String _timeLabel(DateTime dt) {
    final h = dt.hour.toString().padLeft(2, '0');
    final m = dt.minute.toString().padLeft(2, '0');
    return '$h:$m';
  }
}
