import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../../../../core/auth/auth_state.dart';
import '../../../../core/storage/token_storage.dart';
import '../../data/promotion_repository.dart';
import '../../domain/models/voucher.dart';
import '../widgets/voucher_detail_dialog.dart';
import '../widgets/voucher_form_dialog.dart';

final _dateFmt = DateFormat('dd/MM/yy');
final _moneyFmt = NumberFormat.decimalPattern('vi');

/// Voucher list — MVP scope: 4 summary cards + one table (search-free,
/// backend has no pagination and voucher counts are small — see
/// PromotionRepository.listVouchers) + Create/Edit/Enable-Disable. No
/// Clone/Schedule/Delete, no charts (kept out of this pass on purpose).
class VouchersPage extends StatefulWidget {
  const VouchersPage({
    super.key,
    required this.repository,
    required this.authState,
    required this.tokenStorage,
  });

  final PromotionRepository repository;
  final AuthState authState;
  final TokenStorage tokenStorage;

  @override
  State<VouchersPage> createState() => _VouchersPageState();
}

class _VouchersPageState extends State<VouchersPage> {
  bool _loading = true;
  String? _error;
  List<Voucher> _vouchers = const [];
  String _query = '';

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final vouchers = await widget.repository.listVouchers();
      if (!mounted) return;
      setState(() {
        _vouchers = vouchers;
        _loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = 'Không tải được danh sách voucher: $e';
      });
    }
  }

  List<Voucher> get _filtered {
    if (_query.isEmpty) return _vouchers;
    final q = _query.toLowerCase();
    return _vouchers.where((v) => v.title.toLowerCase().contains(q) || v.code.toLowerCase().contains(q)).toList();
  }

  Future<void> _openForm({Voucher? existing}) async {
    final saved = await showDialog<bool>(
      context: context,
      builder: (_) => VoucherFormDialog(repository: widget.repository, existing: existing),
    );
    if (saved == true) _load();
  }

  Future<void> _toggle(Voucher v) async {
    try {
      if (v.enabled) {
        await widget.repository.disableVoucher(v.id);
      } else {
        await widget.repository.enableVoucher(v.id);
      }
      _load();
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Lỗi: $e')));
      }
    }
  }

  void _openDetail(Voucher v) {
    VoucherDetailDialog.show(context, repository: widget.repository, id: v.id);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Voucher')),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _openForm(),
        icon: const Icon(Icons.add),
        label: const Text('Tạo voucher'),
      ),
      body: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            _SummaryCards(vouchers: _vouchers),
            const SizedBox(height: 20),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    onChanged: (q) => setState(() => _query = q),
                    decoration: const InputDecoration(
                      hintText: 'Tìm theo tên / mã voucher',
                      prefixIcon: Icon(Icons.search),
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                IconButton(tooltip: 'Làm mới', onPressed: _load, icon: const Icon(Icons.refresh)),
              ],
            ),
            const SizedBox(height: 12),
            Expanded(child: _buildTable(context)),
          ],
        ),
      ),
    );
  }

  Widget _buildTable(BuildContext context) {
    if (_error != null) {
      return Center(child: Text(_error!));
    }
    if (_loading) {
      return const Center(child: CircularProgressIndicator());
    }
    final rows = _filtered;
    if (rows.isEmpty) {
      return const Center(child: Text('Không có voucher nào', style: TextStyle(color: Color(0xFF9CA3AF))));
    }
    return Card(
      child: Column(
        children: [
          const _TableHeaderRow(),
          const Divider(height: 1),
          Expanded(
            child: ListView.separated(
              itemCount: rows.length,
              separatorBuilder: (_, _) => const Divider(height: 1),
              itemBuilder: (context, i) => _TableDataRow(
                voucher: rows[i],
                onView: () => _openDetail(rows[i]),
                onEdit: () => _openForm(existing: rows[i]),
                onToggle: () => _toggle(rows[i]),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _SummaryCards extends StatelessWidget {
  const _SummaryCards({required this.vouchers});

  final List<Voucher> vouchers;

  @override
  Widget build(BuildContext context) {
    final active = vouchers.where((v) => v.state == 'active').length;
    final disabled = vouchers.where((v) => v.state == 'disabled').length;
    final redeemed = vouchers.fold<int>(0, (sum, v) => sum + (v.statsRedeemed ?? 0));
    final remaining = vouchers.fold<int>(0, (sum, v) => sum + v.remainingBudget);
    return Row(
      children: [
        Expanded(child: _card('Voucher Active', '$active', const Color(0xFF16A34A))),
        const SizedBox(width: 16),
        Expanded(child: _card('Voucher Disabled', '$disabled', const Color(0xFF6B7280))),
        const SizedBox(width: 16),
        Expanded(child: _card('Redeemed', '$redeemed', const Color(0xFF2563EB))),
        const SizedBox(width: 16),
        Expanded(child: _card('Remaining', '${_moneyFmt.format(remaining)}đ', const Color(0xFFF59E0B))),
      ],
    );
  }

  Widget _card(String label, String value, Color color) {
    return Card(
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: const BorderSide(color: Color(0xFFE5E7EB)),
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(label, style: TextStyle(color: color, fontWeight: FontWeight.w600)),
            const SizedBox(height: 8),
            Text(value, style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold)),
          ],
        ),
      ),
    );
  }
}

Color _stateColor(String state) => switch (state) {
      'active' => Colors.green,
      'expired' => Colors.orange,
      'exhausted' => Colors.red,
      _ => Colors.grey,
    };

class _TableHeaderRow extends StatelessWidget {
  const _TableHeaderRow();

  static const _style = TextStyle(fontWeight: FontWeight.w600, fontSize: 12, color: Color(0xFF6B7280));

  @override
  Widget build(BuildContext context) {
    return const Padding(
      padding: EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      child: Row(
        children: [
          Expanded(flex: 3, child: Text('Tên', style: _style)),
          Expanded(flex: 2, child: Text('Code', style: _style)),
          Expanded(flex: 2, child: Text('Loại', style: _style)),
          Expanded(flex: 2, child: Text('Giảm', style: _style)),
          Expanded(flex: 1, child: Text('Đã phát', style: _style)),
          Expanded(flex: 1, child: Text('Đã dùng', style: _style)),
          Expanded(flex: 2, child: Text('Còn lại', style: _style)),
          Expanded(flex: 3, child: Text('Hiệu lực', style: _style)),
          Expanded(flex: 2, child: Text('Status', style: _style)),
          SizedBox(width: 96, child: Text('Action', style: _style)),
        ],
      ),
    );
  }
}

class _TableDataRow extends StatelessWidget {
  const _TableDataRow({
    required this.voucher,
    required this.onView,
    required this.onEdit,
    required this.onToggle,
  });

  final Voucher voucher;
  final VoidCallback onView;
  final VoidCallback onEdit;
  final VoidCallback onToggle;

  @override
  Widget build(BuildContext context) {
    final v = voucher;
    final discountLabel = v.type == 'fixed' ? '${_moneyFmt.format(v.value)}đ' : '${v.value}%';
    final validity = v.start.isEmpty || v.end.isEmpty
        ? '-'
        : '${_dateFmt.format(DateTime.parse(v.start).toLocal())} - ${_dateFmt.format(DateTime.parse(v.end).toLocal())}';
    return InkWell(
      onTap: onView,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
        child: Row(
          children: [
            Expanded(flex: 3, child: Text(v.title)),
            Expanded(flex: 2, child: Text(v.code.isEmpty ? '(auto)' : v.code)),
            Expanded(flex: 2, child: Text(v.type == 'fixed' ? 'Cố định' : 'Phần trăm')),
            Expanded(flex: 2, child: Text(discountLabel)),
            Expanded(flex: 1, child: Text('${v.statsIssued ?? '-'}')),
            Expanded(flex: 1, child: Text('${v.statsRedeemed ?? '-'}')),
            Expanded(flex: 2, child: Text('${_moneyFmt.format(v.remainingBudget)}đ')),
            Expanded(flex: 3, child: Text(validity)),
            Expanded(
              flex: 2,
              child: Text(v.state, style: TextStyle(color: _stateColor(v.state), fontWeight: FontWeight.w600)),
            ),
            SizedBox(
              width: 96,
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  IconButton(tooltip: 'Sửa', icon: const Icon(Icons.edit, size: 20), onPressed: onEdit),
                  Switch(value: v.enabled, onChanged: (_) => onToggle()),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
