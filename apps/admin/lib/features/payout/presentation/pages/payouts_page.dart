import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../../../../core/auth/auth_state.dart';
import '../../../../core/storage/token_storage.dart';
import '../../data/payout_repository.dart';
import '../../domain/models/payout_request.dart';
import '../widgets/payout_detail_dialog.dart';

final _dateFmt = DateFormat('dd/MM/yyyy HH:mm');
final _dayFmt = DateFormat('dd/MM/yyyy');
final _moneyFmt = NumberFormat.decimalPattern('vi');

const _statuses = ['pending', 'approved', 'paid', 'rejected'];
const _statusLabels = {
  'pending': 'Chờ duyệt',
  'approved': 'Đã duyệt',
  'paid': 'Đã trả',
  'rejected': 'Từ chối',
};
const _statusColors = {
  'pending': Color(0xFFF59E0B),
  'approved': Color(0xFF2563EB),
  'paid': Color(0xFF16A34A),
  'rejected': Color(0xFFDC2626),
};

/// Driver Payout review — real-money withdrawal requests. Status is
/// server-filtered (ListByFilter); driver-id search and requested-date
/// range are applied client-side on the current status tab's page (the
/// repository has no date-range or partial-driver-id support — see
/// PayoutRepository's doc comment), same tradeoff as the Voucher list.
class PayoutsPage extends StatefulWidget {
  const PayoutsPage({
    super.key,
    required this.repository,
    required this.authState,
    required this.tokenStorage,
  });

  final PayoutRepository repository;
  final AuthState authState;
  final TokenStorage tokenStorage;

  @override
  State<PayoutsPage> createState() => _PayoutsPageState();
}

class _PayoutsPageState extends State<PayoutsPage> {
  String _status = 'pending';
  String _query = '';
  DateTimeRange? _dateRange;

  bool _loading = true;
  String? _error;
  List<PayoutRequest> _payouts = const [];

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
      final payouts = await widget.repository.listPayouts(status: _status);
      if (!mounted) return;
      setState(() {
        _payouts = payouts;
        _loading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = 'Không tải được danh sách yêu cầu rút tiền: $e';
      });
    }
  }

  void _setStatus(String status) {
    if (status == _status) return;
    setState(() => _status = status);
    _load();
  }

  List<PayoutRequest> get _filtered {
    var rows = _payouts;
    if (_query.isNotEmpty) {
      final q = _query.toLowerCase();
      rows = rows.where((p) => p.driverId.toLowerCase().contains(q)).toList();
    }
    final range = _dateRange;
    if (range != null) {
      rows = rows.where((p) {
        if (p.requestedAt.isEmpty) return false;
        final d = DateTime.parse(p.requestedAt).toLocal();
        final day = DateTime(d.year, d.month, d.day);
        return !day.isBefore(range.start) && !day.isAfter(range.end);
      }).toList();
    }
    return rows;
  }

  Future<void> _pickDateRange() async {
    final now = DateTime.now();
    final picked = await showDateRangePicker(
      context: context,
      firstDate: DateTime(now.year - 2),
      lastDate: DateTime(now.year + 1),
      initialDateRange: _dateRange,
    );
    if (picked != null) setState(() => _dateRange = picked);
  }

  void _openDetail(PayoutRequest p) async {
    await PayoutDetailDialog.show(context, repository: widget.repository, payout: p);
    _load();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Driver Payout')),
      body: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            _StatusTabs(selected: _status, onSelect: _setStatus),
            const SizedBox(height: 12),
            if (!_loading && _error == null) _AggregateSummary(status: _status, payouts: _payouts),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    onChanged: (q) => setState(() => _query = q),
                    decoration: const InputDecoration(
                      hintText: 'Tìm theo mã tài xế (driver ID)',
                      prefixIcon: Icon(Icons.search),
                    ),
                  ),
                ),
                const SizedBox(width: 12),
                OutlinedButton.icon(
                  onPressed: _pickDateRange,
                  icon: const Icon(Icons.date_range),
                  label: Text(_dateRange == null
                      ? 'Lọc theo ngày'
                      : '${_dayFmt.format(_dateRange!.start)} - ${_dayFmt.format(_dateRange!.end)}'),
                ),
                if (_dateRange != null)
                  IconButton(tooltip: 'Bỏ lọc ngày', onPressed: () => setState(() => _dateRange = null), icon: const Icon(Icons.close)),
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
      return const Center(child: Text('Không có yêu cầu nào', style: TextStyle(color: Color(0xFF9CA3AF))));
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
              itemBuilder: (context, i) => _TableDataRow(payout: rows[i], onView: () => _openDetail(rows[i])),
            ),
          ),
        ],
      ),
    );
  }
}

class _AggregateSummary extends StatelessWidget {
  const _AggregateSummary({required this.status, required this.payouts});

  final String status;
  final List<PayoutRequest> payouts;

  @override
  Widget build(BuildContext context) {
    final total = payouts.fold<int>(0, (sum, p) => sum + p.amount);
    final label = _statusLabels[status] ?? status;
    return Text(
      '$label: ${_moneyFmt.format(total)} đ · ${payouts.length} yêu cầu',
      style: TextStyle(fontWeight: FontWeight.w600, fontSize: 15, color: _statusColors[status]),
    );
  }
}

class _StatusTabs extends StatelessWidget {
  const _StatusTabs({required this.selected, required this.onSelect});

  final String selected;
  final ValueChanged<String> onSelect;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        for (final status in _statuses) ...[
          Expanded(child: _tab(status)),
          if (status != _statuses.last) const SizedBox(width: 16),
        ],
      ],
    );
  }

  Widget _tab(String status) {
    final color = _statusColors[status]!;
    final isSelected = selected == status;
    return Card(
      color: isSelected ? color.withValues(alpha: 0.08) : null,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: isSelected ? color : const Color(0xFFE5E7EB), width: isSelected ? 2 : 1),
      ),
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: () => onSelect(status),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Text(_statusLabels[status]!, style: TextStyle(color: color, fontWeight: FontWeight.w600, fontSize: 16)),
        ),
      ),
    );
  }
}

class _TableHeaderRow extends StatelessWidget {
  const _TableHeaderRow();

  static const _style = TextStyle(fontWeight: FontWeight.w600, fontSize: 12, color: Color(0xFF6B7280));

  @override
  Widget build(BuildContext context) {
    return const Padding(
      padding: EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      child: Row(
        children: [
          Expanded(flex: 3, child: Text('Tài xế', style: _style)),
          Expanded(flex: 2, child: Text('Số tiền', style: _style)),
          Expanded(flex: 2, child: Text('Ngân hàng', style: _style)),
          Expanded(flex: 2, child: Text('Số TK', style: _style)),
          Expanded(flex: 3, child: Text('Ngày yêu cầu', style: _style)),
          Expanded(flex: 2, child: Text('Trạng thái', style: _style)),
          SizedBox(width: 80, child: Text('Action', style: _style)),
        ],
      ),
    );
  }
}

class _TableDataRow extends StatelessWidget {
  const _TableDataRow({required this.payout, required this.onView});

  final PayoutRequest payout;
  final VoidCallback onView;

  @override
  Widget build(BuildContext context) {
    final p = payout;
    return InkWell(
      onTap: onView,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
        child: Row(
          children: [
            Expanded(flex: 3, child: Text(p.driverId, overflow: TextOverflow.ellipsis)),
            Expanded(flex: 2, child: Text('${_moneyFmt.format(p.amount)} ${p.currency}')),
            Expanded(flex: 2, child: Text(p.bankName.isEmpty ? '-' : p.bankName)),
            Expanded(flex: 2, child: Text(p.maskedAccountNumber.isEmpty ? '-' : p.maskedAccountNumber)),
            Expanded(flex: 3, child: Text(p.requestedAt.isEmpty ? '-' : _dateFmt.format(DateTime.parse(p.requestedAt).toLocal()))),
            Expanded(
              flex: 2,
              child: Text(
                _statusLabels[p.status] ?? p.status,
                style: TextStyle(color: _statusColors[p.status] ?? Colors.grey, fontWeight: FontWeight.w600),
              ),
            ),
            SizedBox(width: 80, child: TextButton(onPressed: onView, child: const Text('Xem'))),
          ],
        ),
      ),
    );
  }
}
