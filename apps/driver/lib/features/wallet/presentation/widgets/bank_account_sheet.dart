import 'package:flutter/material.dart';

import '../../../../core/network/api_client.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_button.dart';
import '../../data/wallet_repository.dart';
import '../../domain/models/bank_account.dart';

/// Phần 6 — "Driver thêm Ngân hàng: Tên chủ TK, Số tài khoản, Chi nhánh
/// (optional)". Only ever one account (adding replaces the existing one —
/// the backend upserts by driver).
class BankAccountSheet extends StatefulWidget {
  const BankAccountSheet({super.key, required this.repository, this.existing});

  final WalletRepository repository;
  final BankAccount? existing;

  @override
  State<BankAccountSheet> createState() => _BankAccountSheetState();
}

class _BankAccountSheetState extends State<BankAccountSheet> {
  late final _bankNameCtrl = TextEditingController(text: widget.existing?.bankName ?? '');
  final _holderNameCtrl = TextEditingController();
  final _accountNumberCtrl = TextEditingController();
  late final _branchCtrl = TextEditingController(text: widget.existing?.branchName ?? '');
  bool _submitting = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _holderNameCtrl.text = widget.existing?.accountHolderName ?? '';
  }

  @override
  void dispose() {
    _bankNameCtrl.dispose();
    _holderNameCtrl.dispose();
    _accountNumberCtrl.dispose();
    _branchCtrl.dispose();
    super.dispose();
  }

  bool get _canSubmit =>
      _bankNameCtrl.text.trim().isNotEmpty &&
      _holderNameCtrl.text.trim().isNotEmpty &&
      _accountNumberCtrl.text.trim().isNotEmpty;

  Future<void> _submit() async {
    if (!_canSubmit) {
      setState(() => _error = 'Vui lòng nhập đủ Ngân hàng, Tên chủ TK và Số tài khoản');
      return;
    }
    setState(() {
      _submitting = true;
      _error = null;
    });
    try {
      await widget.repository.setBankAccount(
        bankName: _bankNameCtrl.text.trim(),
        accountHolderName: _holderNameCtrl.text.trim(),
        accountNumber: _accountNumberCtrl.text.trim(),
        branchName: _branchCtrl.text.trim(),
      );
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
          _error = 'Không thể lưu tài khoản ngân hàng. Vui lòng thử lại.';
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(widget.existing == null ? 'Thêm ngân hàng' : 'Cập nhật ngân hàng', style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: AppSpacing.sm),
          Text(
            'Chỉ lưu 1 tài khoản mặc định. Thêm mới sẽ thay thế tài khoản hiện tại.',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppColors.textSecondary),
          ),
          const SizedBox(height: AppSpacing.lg),
          TextField(controller: _bankNameCtrl, decoration: const InputDecoration(labelText: 'Ngân hàng')),
          const SizedBox(height: AppSpacing.md),
          TextField(controller: _holderNameCtrl, decoration: const InputDecoration(labelText: 'Tên chủ tài khoản')),
          const SizedBox(height: AppSpacing.md),
          TextField(
            controller: _accountNumberCtrl,
            keyboardType: TextInputType.number,
            decoration: InputDecoration(labelText: 'Số tài khoản', hintText: widget.existing?.maskedAccountNumber),
          ),
          const SizedBox(height: AppSpacing.md),
          TextField(controller: _branchCtrl, decoration: const InputDecoration(labelText: 'Chi nhánh (không bắt buộc)')),
          if (_error != null) ...[
            const SizedBox(height: AppSpacing.sm),
            Text(_error!, style: const TextStyle(color: AppColors.error, fontSize: 13)),
          ],
          const SizedBox(height: AppSpacing.lg),
          AppButton.primary(label: 'Lưu', isLoading: _submitting, onPressed: _submit),
        ],
      ),
    );
  }
}
