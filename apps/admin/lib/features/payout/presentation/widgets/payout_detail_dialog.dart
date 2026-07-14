import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../../data/payout_repository.dart';
import '../../domain/models/payout_request.dart';

final _dateFmt = DateFormat('dd/MM/yyyy HH:mm');
final _moneyFmt = NumberFormat.decimalPattern('vi');

/// Payout detail + review actions (Approve/Reject/Mark Paid) — status-gated,
/// same show-as-dialog / pop(true)-to-reload pattern as
/// VerificationDetailDialog and VoucherDetailDialog.
class PayoutDetailDialog extends StatefulWidget {
  const PayoutDetailDialog({super.key, required this.repository, required this.payout});

  final PayoutRepository repository;
  final PayoutRequest payout;

  static Future<void> show(BuildContext context, {required PayoutRepository repository, required PayoutRequest payout}) async {
    final changed = await showDialog<bool>(
      context: context,
      builder: (_) => PayoutDetailDialog(repository: repository, payout: payout),
    );
    if (changed == true) return;
  }

  @override
  State<PayoutDetailDialog> createState() => _PayoutDetailDialogState();
}

class _PayoutDetailDialogState extends State<PayoutDetailDialog> {
  bool _busy = false;
  String? _error;

  Future<void> _approve() => _run(() => widget.repository.approve(widget.payout.id));

  Future<void> _markPaid() => _run(() => widget.repository.markPaid(widget.payout.id));

  Future<void> _reject() async {
    final reason = await showDialog<String>(context: context, builder: (_) => const _RejectReasonDialog());
    if (reason == null || reason.trim().isEmpty) return;
    await _run(() => widget.repository.reject(widget.payout.id, reason.trim()));
  }

  Future<void> _run(Future<void> Function() action) async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      await action();
      if (mounted) Navigator.pop(context, true);
    } catch (e) {
      if (mounted) setState(() => _error = '$e');
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final p = widget.payout;
    return AlertDialog(
      title: Text('Yêu cầu rút tiền · ${p.driverId}'),
      content: SizedBox(
        width: 420,
        child: SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _row('Tài xế', p.driverId),
              _row('Số tiền', '${_moneyFmt.format(p.amount)} ${p.currency}'),
              _row('Ngân hàng', p.bankName.isEmpty ? '-' : p.bankName),
              _row('Số tài khoản', p.maskedAccountNumber.isEmpty ? '-' : p.maskedAccountNumber),
              _row('Ngày yêu cầu', p.requestedAt.isEmpty ? '-' : _dateFmt.format(DateTime.parse(p.requestedAt).toLocal())),
              _row('Trạng thái', p.status),
              if (p.reviewedAt != null) _row('Ngày duyệt', _dateFmt.format(DateTime.parse(p.reviewedAt!).toLocal())),
              if (p.rejectReason != null && p.rejectReason!.isNotEmpty) _row('Lý do từ chối', p.rejectReason!),
              if (p.paidAt != null) _row('Ngày trả', _dateFmt.format(DateTime.parse(p.paidAt!).toLocal())),
              const Divider(height: 24),
              Text('Lịch sử rút của tài xế này', style: Theme.of(context).textTheme.titleSmall),
              const SizedBox(height: 8),
              _DriverHistory(repository: widget.repository, driverId: p.driverId, excludeId: p.id),
              if (_error != null) ...[
                const SizedBox(height: 8),
                Text('Lỗi: $_error', style: const TextStyle(color: Colors.red)),
              ],
            ],
          ),
        ),
      ),
      actions: [
        TextButton(onPressed: _busy ? null : () => Navigator.pop(context), child: const Text('Đóng')),
        if (p.status == 'pending') ...[
          TextButton(onPressed: _busy ? null : _reject, child: const Text('Từ chối', style: TextStyle(color: Colors.red))),
          FilledButton(onPressed: _busy ? null : _approve, child: const Text('Duyệt')),
        ],
        if (p.status == 'approved')
          FilledButton(onPressed: _busy ? null : _markPaid, child: const Text('Đánh dấu đã trả')),
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

class _DriverHistory extends StatelessWidget {
  const _DriverHistory({required this.repository, required this.driverId, required this.excludeId});

  final PayoutRepository repository;
  final String driverId;
  final String excludeId;

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<List<PayoutRequest>>(
      future: repository.listByDriver(driverId),
      builder: (context, snapshot) {
        if (snapshot.connectionState != ConnectionState.done) {
          return const Padding(padding: EdgeInsets.symmetric(vertical: 8), child: LinearProgressIndicator());
        }
        if (snapshot.hasError) {
          return Text('Không tải được lịch sử: ${snapshot.error}', style: const TextStyle(color: Colors.red, fontSize: 12));
        }
        final others = (snapshot.data ?? []).where((r) => r.id != excludeId).toList();
        if (others.isEmpty) {
          return const Text('Chưa có yêu cầu nào khác', style: TextStyle(color: Color(0xFF9CA3AF), fontSize: 12));
        }
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            for (final r in others)
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 2),
                child: Text(
                  '${r.requestedAt.isEmpty ? "-" : _dateFmt.format(DateTime.parse(r.requestedAt).toLocal())} · '
                  '${_moneyFmt.format(r.amount)} ${r.currency} · ${r.status}',
                  style: const TextStyle(fontSize: 12),
                ),
              ),
          ],
        );
      },
    );
  }
}

class _RejectReasonDialog extends StatefulWidget {
  const _RejectReasonDialog();

  @override
  State<_RejectReasonDialog> createState() => _RejectReasonDialogState();
}

class _RejectReasonDialogState extends State<_RejectReasonDialog> {
  final _ctrl = TextEditingController();

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Lý do từ chối'),
      content: TextField(
        controller: _ctrl,
        autofocus: true,
        maxLines: 3,
        decoration: const InputDecoration(hintText: 'Nhập lý do từ chối yêu cầu rút tiền'),
      ),
      actions: [
        TextButton(onPressed: () => Navigator.pop(context), child: const Text('Huỷ')),
        FilledButton(onPressed: () => Navigator.pop(context, _ctrl.text), child: const Text('Xác nhận')),
      ],
    );
  }
}
