import 'package:flutter/material.dart';

import '../../../../core/network/api_client.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/utils/currency_format.dart';
import '../../../../shared/widgets/app_button.dart';
import '../../data/wallet_repository.dart';
import '../../domain/models/wallet_summary.dart';

/// Phần 3/5/8 — "Driver nhấn Rút tiền. Nhập số tiền." A bottom sheet, not a
/// full page, since it's a single-field form. All validation messages come
/// straight from the backend (Phần 5 — "Thông báo tiếng Việt") rather than
/// being duplicated client-side, so the rules stay in exactly one place.
class PayoutRequestSheet extends StatefulWidget {
  const PayoutRequestSheet({super.key, required this.repository, required this.summary});

  final WalletRepository repository;
  final WalletSummary summary;

  @override
  State<PayoutRequestSheet> createState() => _PayoutRequestSheetState();
}

class _PayoutRequestSheetState extends State<PayoutRequestSheet> {
  final _amountCtrl = TextEditingController();
  bool _submitting = false;
  String? _error;

  @override
  void dispose() {
    _amountCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final amount = int.tryParse(_amountCtrl.text.replaceAll('.', '').trim());
    if (amount == null || amount <= 0) {
      setState(() => _error = 'Vui lòng nhập số tiền hợp lệ');
      return;
    }
    setState(() {
      _submitting = true;
      _error = null;
    });
    try {
      await widget.repository.createPayoutRequest(amount);
      if (mounted) Navigator.of(context).pop(true);
    } on ApiException catch (e) {
      if (mounted) {
        setState(() {
          _submitting = false;
          _error = e.message;
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() {
          _submitting = false;
          _error = 'Không thể gửi yêu cầu rút tiền. Vui lòng thử lại.';
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Rút tiền', style: Theme.of(context).textTheme.titleLarge),
        const SizedBox(height: AppSpacing.sm),
        Text(
          'Số dư khả dụng: ${formatMoney(widget.summary.netCents, widget.summary.currency)}',
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary),
        ),
        const SizedBox(height: AppSpacing.lg),
        TextField(
          controller: _amountCtrl,
          keyboardType: TextInputType.number,
          autofocus: true,
          decoration: const InputDecoration(labelText: 'Số tiền muốn rút', suffixText: 'đ'),
        ),
        if (_error != null) ...[
          const SizedBox(height: AppSpacing.sm),
          Text(_error!, style: const TextStyle(color: AppColors.error, fontSize: 13)),
        ],
        const SizedBox(height: AppSpacing.lg),
        AppButton.primary(label: 'Xác nhận rút tiền', isLoading: _submitting, onPressed: _submit),
      ],
    );
  }
}
