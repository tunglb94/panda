import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../../../../core/auth/auth_state.dart';
import '../../../../core/storage/token_storage.dart';
import '../../data/kyc_repository.dart';
import '../../domain/models/driver_verification_row.dart';
import '../../domain/models/kyc_summary.dart';
import '../widgets/avatar_thumbnail.dart';
import '../widgets/status_badge.dart';
import '../widgets/vehicle_badge.dart';
import '../widgets/verification_detail_dialog.dart';

final _dateFmt = DateFormat('dd/MM/yyyy HH:mm');

/// Combines Phần 1 (the driver-verifications table) and Phần 10 (the 4
/// dashboard summary cards) into a single page — the cards double as status
/// filters, so approving/rejecting from the table immediately updates both
/// without a separate navigation. Matches Phần 12's "ưu tiên tốc độ".
class DriverVerificationsPage extends StatefulWidget {
  const DriverVerificationsPage({
    super.key,
    required this.repository,
    required this.authState,
    required this.tokenStorage,
  });

  final KYCRepository repository;
  final AuthState authState;
  final TokenStorage tokenStorage;

  @override
  State<DriverVerificationsPage> createState() => _DriverVerificationsPageState();
}

class _DriverVerificationsPageState extends State<DriverVerificationsPage> {
  String _status = 'pending';
  String _query = '';
  bool _sortAsc = false;
  final _searchCtrl = TextEditingController();

  KYCSummary? _summary;
  List<DriverVerificationRow>? _rows;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _searchCtrl.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() => _error = null);
    try {
      final summary = await widget.repository.getSummary();
      final rows = await widget.repository.listDriverVerifications(
        status: _status,
        query: _query,
        sortAsc: _sortAsc,
      );
      if (mounted) {
        setState(() {
          _summary = summary;
          _rows = rows;
        });
      }
    } catch (e) {
      if (mounted) setState(() => _error = '$e');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('KYC Review')),
      body: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            _SummaryCards(summary: _summary, selected: _status, onSelect: _setStatus),
            const SizedBox(height: 20),
            _Toolbar(
              searchCtrl: _searchCtrl,
              sortAsc: _sortAsc,
              onSearch: (q) {
                _query = q;
                _load();
              },
              onSortToggle: () {
                setState(() => _sortAsc = !_sortAsc);
                _load();
              },
              onRefresh: _load,
            ),
            const SizedBox(height: 12),
            Expanded(child: _buildTable(context)),
          ],
        ),
      ),
    );
  }

  void _setStatus(String status) {
    setState(() => _status = status);
    _load();
  }

  Widget _buildTable(BuildContext context) {
    if (_error != null) {
      return Center(child: Text('Lỗi tải danh sách: $_error'));
    }
    final rows = _rows;
    if (rows == null) {
      return const Center(child: CircularProgressIndicator());
    }
    if (rows.isEmpty) {
      return const Center(child: Text('Không có hồ sơ nào', style: TextStyle(color: Color(0xFF9CA3AF))));
    }
    return Card(
      child: Column(
        children: [
          _TableHeaderRow(),
          const Divider(height: 1),
          Expanded(
            child: ListView.separated(
              itemCount: rows.length,
              separatorBuilder: (_, _) => const Divider(height: 1),
              itemBuilder: (context, i) => _TableDataRow(
                row: rows[i],
                repository: widget.repository,
                onView: () async {
                  final changed = await VerificationDetailDialog.show(
                    context,
                    repository: widget.repository,
                    driverId: rows[i].driverId,
                  );
                  if (changed == true) _load();
                },
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _SummaryCards extends StatelessWidget {
  const _SummaryCards({required this.summary, required this.selected, required this.onSelect});

  final KYCSummary? summary;
  final String selected;
  final ValueChanged<String> onSelect;

  @override
  Widget build(BuildContext context) {
    final s = summary;
    return Row(
      children: [
        Expanded(child: _card('Chờ duyệt', s?.pending, 'pending', const Color(0xFFF59E0B))),
        const SizedBox(width: 16),
        Expanded(child: _card('Đã duyệt', s?.approved, 'approved', const Color(0xFF16A34A))),
        const SizedBox(width: 16),
        Expanded(child: _card('Từ chối', s?.rejected, 'rejected', const Color(0xFFDC2626))),
        const SizedBox(width: 16),
        Expanded(child: _card('Hết hạn', s?.expired, 'expired', const Color(0xFF6B7280))),
      ],
    );
  }

  Widget _card(String label, int? value, String status, Color color) {
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
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(label, style: TextStyle(color: color, fontWeight: FontWeight.w600)),
              const SizedBox(height: 8),
              Text(
                value?.toString() ?? '—',
                style: const TextStyle(fontSize: 28, fontWeight: FontWeight.bold),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _Toolbar extends StatelessWidget {
  const _Toolbar({
    required this.searchCtrl,
    required this.sortAsc,
    required this.onSearch,
    required this.onSortToggle,
    required this.onRefresh,
  });

  final TextEditingController searchCtrl;
  final bool sortAsc;
  final ValueChanged<String> onSearch;
  final VoidCallback onSortToggle;
  final VoidCallback onRefresh;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(
          child: TextField(
            controller: searchCtrl,
            onSubmitted: onSearch,
            decoration: InputDecoration(
              hintText: 'Tìm theo Tên / SĐT / CCCD',
              prefixIcon: const Icon(Icons.search),
              suffixIcon: IconButton(icon: const Icon(Icons.arrow_forward), onPressed: () => onSearch(searchCtrl.text)),
            ),
          ),
        ),
        const SizedBox(width: 12),
        OutlinedButton.icon(
          onPressed: onSortToggle,
          icon: Icon(sortAsc ? Icons.arrow_upward : Icons.arrow_downward),
          label: Text(sortAsc ? 'Cũ nhất' : 'Mới nhất'),
        ),
        const SizedBox(width: 12),
        IconButton(tooltip: 'Làm mới', onPressed: onRefresh, icon: const Icon(Icons.refresh)),
      ],
    );
  }
}

class _TableHeaderRow extends StatelessWidget {
  const _TableHeaderRow();

  static const _style = TextStyle(fontWeight: FontWeight.w600, fontSize: 12, color: Color(0xFF6B7280));

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      child: Row(
        children: const [
          SizedBox(width: 44),
          Expanded(flex: 3, child: Text('Tên', style: _style)),
          Expanded(flex: 2, child: Text('SĐT', style: _style)),
          Expanded(flex: 2, child: Text('Loại xe / Service', style: _style)),
          Expanded(flex: 2, child: Text('Ngày gửi', style: _style)),
          Expanded(flex: 2, child: Text('Trạng thái', style: _style)),
          SizedBox(width: 90, child: Text('Action', style: _style)),
        ],
      ),
    );
  }
}

class _TableDataRow extends StatelessWidget {
  const _TableDataRow({required this.row, required this.repository, required this.onView});

  final DriverVerificationRow row;
  final KYCRepository repository;
  final VoidCallback onView;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        children: [
          SizedBox(width: 44, child: AvatarThumbnail(repository: repository, driverId: row.driverId)),
          Expanded(flex: 3, child: Text(row.fullName.isEmpty ? '—' : row.fullName)),
          Expanded(flex: 2, child: Text(row.phone.isEmpty ? '—' : row.phone)),
          Expanded(flex: 2, child: VehicleBadge(serviceType: row.serviceType)),
          Expanded(flex: 2, child: Text(_dateFmt.format(row.submittedAt))),
          Expanded(flex: 2, child: StatusBadge(status: row.status)),
          SizedBox(
            width: 90,
            child: TextButton(onPressed: onView, child: const Text('Xem')),
          ),
        ],
      ),
    );
  }
}
