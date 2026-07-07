import 'package:flutter/material.dart';

import '../../domain/models/notification_item.dart';

/// Single Notification Center row. Unread items get a filled icon
/// background, a bold title, and a small dot indicator; tapping animates
/// the transition to the read state via [onTap].
class NotificationTile extends StatelessWidget {
  const NotificationTile({super.key, required this.item, required this.onTap});

  final NotificationItem item;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    final unread = !item.isRead;

    return InkWell(
      borderRadius: BorderRadius.circular(12),
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 300),
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: unread ? primary.withValues(alpha: 0.05) : Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: unread ? primary.withValues(alpha: 0.3) : Colors.grey.shade200,
          ),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: unread ? primary.withValues(alpha: 0.15) : Colors.grey.shade100,
                shape: BoxShape.circle,
              ),
              child: Icon(
                item.type.icon,
                size: 18,
                color: unread ? primary : Colors.grey.shade500,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          item.title,
                          style: TextStyle(
                            fontWeight: unread ? FontWeight.w700 : FontWeight.w500,
                          ),
                        ),
                      ),
                      if (unread)
                        Container(
                          width: 8,
                          height: 8,
                          margin: const EdgeInsets.only(left: 6, top: 4),
                          decoration: BoxDecoration(
                            color: primary,
                            shape: BoxShape.circle,
                          ),
                        ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Text(
                    item.body,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(fontSize: 13, color: Colors.grey.shade600),
                  ),
                  const SizedBox(height: 6),
                  Text(
                    item.relativeTimeLabel,
                    style: TextStyle(fontSize: 11, color: Colors.grey.shade400),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
