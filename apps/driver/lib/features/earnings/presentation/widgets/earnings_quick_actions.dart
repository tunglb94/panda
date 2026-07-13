import 'package:flutter/material.dart';

import '../../../../core/network/api_client.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_snackbar.dart';
import '../../../../shared/widgets/pressable_scale.dart';
import '../../../wallet/presentation/pages/wallet_page.dart';

/// Quick action shortcuts. "Lịch sử" scrolls to the real Transaction
/// History section further down this same page (real data, no new route).
/// "Rút tiền"/"Ngân hàng" open the real Wallet screen (Settlement Engine —
/// Phần 13), where the Rút tiền button and bank-account form actually live.
/// "Sao kê"/"Hỗ trợ" remain honest placeholders — there is no statement
/// export or support backend today — tapping shows a clear "not
/// implemented yet" snackbar rather than pretending to do something.
class EarningsQuickActions extends StatelessWidget {
  const EarningsQuickActions({super.key, required this.apiClient, required this.onViewHistory});

  final ApiClient apiClient;
  final VoidCallback onViewHistory;

  void _openWallet(BuildContext context) {
    Navigator.of(context).push(MaterialPageRoute(builder: (_) => WalletPage(apiClient: apiClient)));
  }

  @override
  Widget build(BuildContext context) {
    final actions = [
      (Icons.arrow_circle_up_outlined, 'Rút tiền', () => _openWallet(context)),
      (Icons.receipt_long_outlined, 'Lịch sử', onViewHistory),
      (Icons.description_outlined, 'Sao kê', () => _placeholder(context, 'Xuất sao kê')),
      (Icons.account_balance_outlined, 'Ngân hàng', () => _openWallet(context)),
      (Icons.support_agent_outlined, 'Hỗ trợ', () => _placeholder(context, 'Hỗ trợ')),
    ];

    return Row(
      children: actions
          .map((a) => Expanded(child: _QuickActionButton(icon: a.$1, label: a.$2, onTap: a.$3)))
          .toList(),
    );
  }

  static void _placeholder(BuildContext context, String feature) {
    AppSnackbar.show(context, '$feature chưa khả dụng — sẽ ra mắt trong giai đoạn tiếp theo.');
  }
}

class _QuickActionButton extends StatefulWidget {
  const _QuickActionButton({required this.icon, required this.label, required this.onTap});

  final IconData icon;
  final String label;
  final VoidCallback onTap;

  @override
  State<_QuickActionButton> createState() => _QuickActionButtonState();
}

class _QuickActionButtonState extends State<_QuickActionButton> {
  bool _pressed = false;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: widget.onTap,
        onHighlightChanged: (v) => setState(() => _pressed = v),
        borderRadius: BorderRadius.circular(14),
        child: PressableScale(
          pressed: _pressed,
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: AppSpacing.sm),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Container(
                  width: 48,
                  height: 48,
                  decoration: BoxDecoration(
                    color: AppColors.primaryLight,
                    shape: BoxShape.circle,
                  ),
                  child: Icon(widget.icon, color: AppColors.primary, size: 22),
                ),
                const SizedBox(height: 6),
                Text(
                  widget.label,
                  textAlign: TextAlign.center,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(fontSize: 11, fontWeight: FontWeight.w600),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
