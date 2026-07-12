import 'package:flutter/material.dart';

import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../../../shared/widgets/app_settings_tile.dart';
import '../../../../shared/widgets/app_snackbar.dart';

class _SupportEntry {
  const _SupportEntry(this.icon, this.label);
  final IconData icon;
  final String label;
}

const _entries = [
  _SupportEntry(Icons.quiz_outlined, 'Câu hỏi thường gặp (FAQ)'),
  _SupportEntry(Icons.call_outlined, 'Hotline hỗ trợ'),
  _SupportEntry(Icons.chat_bubble_outline, 'Chat với hỗ trợ'),
  _SupportEntry(Icons.email_outlined, 'Gửi email'),
  _SupportEntry(Icons.bug_report_outlined, 'Báo lỗi'),
  _SupportEntry(Icons.feedback_outlined, 'Góp ý sản phẩm'),
];

/// Support — every entry is an honest placeholder. There is no FAQ CMS, no
/// hotline number on file, no chat/email backend anywhere in this project;
/// tapping any row says so plainly rather than pretending to connect.
class SupportPage extends StatelessWidget {
  const SupportPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Hỗ trợ')),
      body: ListView(
        padding: const EdgeInsets.all(AppSpacing.lg),
        children: [
          AppCard(
            padding: EdgeInsets.zero,
            child: Column(
              children: [
                for (var i = 0; i < _entries.length; i++) ...[
                  AppSettingsTile(
                    icon: _entries[i].icon,
                    label: _entries[i].label,
                    onTap: () => AppSnackbar.show(
                      context,
                      '${_entries[i].label} chưa khả dụng — sẽ ra mắt trong giai đoạn tiếp theo.',
                    ),
                  ),
                  if (i != _entries.length - 1) const Divider(height: 1),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }
}
