import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../../data/promotion_repository.dart';
import '../../domain/models/voucher.dart';

final _dateFmt = DateFormat('dd/MM/yyyy');
final _moneyFmt = NumberFormat.decimalPattern('vi');

/// Read-only voucher detail — full configuration, budget usage, eligible
/// services, and current stats. Reuses the existing `GET
/// /api/v1/admin/vouchers/{id}` endpoint (fresh stats on every open),
/// matching KYC's VerificationDetailDialog show-as-dialog pattern.
class VoucherDetailDialog extends StatelessWidget {
  const VoucherDetailDialog({super.key, required this.voucher});

  final Voucher voucher;

  static Future<void> show(BuildContext context, {required PromotionRepository repository, required String id}) async {
    final voucher = await showDialog<Voucher>(
      context: context,
      builder: (_) => _LoadingDialog(repository: repository, id: id),
    );
    if (voucher == null || !context.mounted) return;
  }

  @override
  Widget build(BuildContext context) {
    final v = voucher;
    final discountLabel = v.type == 'fixed' ? '${_moneyFmt.format(v.value)}đ' : '${v.value}%';
    return AlertDialog(
      title: Text('${v.title}  ·  ${v.code.isEmpty ? "(auto)" : v.code}'),
      content: SizedBox(
        width: 420,
        child: SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _row('Campaign', v.campaign.isEmpty ? '-' : v.campaign),
              _row('Dịch vụ', v.serviceType.isEmpty ? 'Tất cả' : v.serviceType.join(', ')),
              _row('Giảm giá', discountLabel),
              if (v.maxDiscount > 0) _row('Giảm tối đa', '${_moneyFmt.format(v.maxDiscount)}đ'),
              if (v.minOrder > 0) _row('Đơn tối thiểu', '${_moneyFmt.format(v.minOrder)}đ'),
              const Divider(height: 24),
              _row('Ngân sách', '${_moneyFmt.format(v.budget)}đ'),
              _row('Đã dùng', v.statsRedeemed?.toString() ?? '${v.usageCount}'),
              _row('Còn lại', '${_moneyFmt.format(v.remainingBudget)}đ'),
              if (v.statsIssued != null) _row('Đã phát', '${v.statsIssued}'),
              if (v.statsExpired != null) _row('Hết hạn', '${v.statsExpired}'),
              const Divider(height: 24),
              _row('Ngày bắt đầu', v.start.isEmpty ? '-' : _dateFmt.format(DateTime.parse(v.start).toLocal())),
              _row('Ngày kết thúc', v.end.isEmpty ? '-' : _dateFmt.format(DateTime.parse(v.end).toLocal())),
              _row('Trạng thái', v.state),
            ],
          ),
        ),
      ),
      actions: [
        TextButton(onPressed: () => Navigator.pop(context), child: const Text('Đóng')),
      ],
    );
  }

  Widget _row(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(width: 130, child: Text(label, style: const TextStyle(color: Color(0xFF6B7280)))),
          Expanded(child: Text(value)),
        ],
      ),
    );
  }
}

class _LoadingDialog extends StatelessWidget {
  const _LoadingDialog({required this.repository, required this.id});

  final PromotionRepository repository;
  final String id;

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<Voucher>(
      future: repository.getVoucher(id),
      builder: (context, snapshot) {
        if (snapshot.connectionState != ConnectionState.done) {
          return const AlertDialog(content: SizedBox(height: 80, child: Center(child: CircularProgressIndicator())));
        }
        if (snapshot.hasError) {
          return AlertDialog(
            title: const Text('Lỗi'),
            content: Text('Không tải được chi tiết voucher: ${snapshot.error}'),
            actions: [TextButton(onPressed: () => Navigator.pop(context), child: const Text('Đóng'))],
          );
        }
        return VoucherDetailDialog(voucher: snapshot.data!);
      },
    );
  }
}
