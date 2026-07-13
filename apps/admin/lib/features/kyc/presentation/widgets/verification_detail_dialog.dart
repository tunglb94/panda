import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../../data/kyc_repository.dart';
import '../../domain/models/audit_log_entry.dart';
import '../../domain/models/kyc_detail.dart';
import '../../domain/models/kyc_document_item.dart';
import 'document_thumbnail.dart';
import 'reject_reason_dialog.dart';
import 'status_badge.dart';
import 'vehicle_badge.dart';
import 'web_download_stub.dart';

final _dateFmt = DateFormat('dd/MM/yyyy HH:mm');

/// The review Drawer/Dialog (Phần 2/3/8/9) — opened by "Xem" from the list
/// table. Not a separate route/page: `showDialog` keeps the admin's list
/// scroll position and filters intact underneath.
class VerificationDetailDialog extends StatefulWidget {
  const VerificationDetailDialog({super.key, required this.repository, required this.driverId});

  final KYCRepository repository;
  final String driverId;

  /// Returns true if an approve/reject action was taken, so the caller can
  /// refresh the list.
  static Future<bool?> show(BuildContext context, {required KYCRepository repository, required String driverId}) {
    return showDialog<bool>(
      context: context,
      builder: (_) => VerificationDetailDialog(repository: repository, driverId: driverId),
    );
  }

  @override
  State<VerificationDetailDialog> createState() => _VerificationDetailDialogState();
}

class _VerificationDetailDialogState extends State<VerificationDetailDialog> {
  late Future<KYCDetail> _detailFuture;
  bool _busy = false;
  bool _changed = false;

  @override
  void initState() {
    super.initState();
    _detailFuture = widget.repository.getDetail(widget.driverId);
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      insetPadding: const EdgeInsets.all(24),
      child: SizedBox(
        width: 960,
        height: 720,
        child: FutureBuilder<KYCDetail>(
          future: _detailFuture,
          builder: (context, snapshot) {
            if (snapshot.connectionState != ConnectionState.done) {
              return const Center(child: CircularProgressIndicator());
            }
            if (snapshot.hasError) {
              return Center(child: Text('Không tải được hồ sơ: ${snapshot.error}'));
            }
            return _buildContent(context, snapshot.data!);
          },
        ),
      ),
    );
  }

  Widget _buildContent(BuildContext context, KYCDetail detail) {
    final dv = detail.driverVerification;
    final vv = detail.vehicleVerification;
    final submitCount = detail.auditLog.where((a) => a.action == 'submit').length;

    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(20, 16, 12, 8),
          child: Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(dv?.fullName ?? '(chưa có hồ sơ cá nhân)',
                        style: Theme.of(context).textTheme.titleLarge),
                    const SizedBox(height: 4),
                    Row(
                      children: [
                        if (dv != null) StatusBadge(status: dv.status),
                        const SizedBox(width: 8),
                        VehicleBadge(serviceType: vv?.serviceType),
                        const SizedBox(width: 8),
                        Text('SĐT: ${detail.phone}', style: Theme.of(context).textTheme.bodySmall),
                        const SizedBox(width: 8),
                        if (submitCount > 0)
                          Text('• Đã gửi $submitCount lần', style: Theme.of(context).textTheme.bodySmall),
                      ],
                    ),
                  ],
                ),
              ),
              IconButton(onPressed: () => Navigator.of(context).pop(_changed), icon: const Icon(Icons.close)),
            ],
          ),
        ),
        const Divider(height: 1),
        Expanded(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(20),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(flex: 4, child: _InfoPanel(dv: dv, vv: vv)),
                const SizedBox(width: 24),
                Expanded(flex: 5, child: _DocumentGrid(repository: widget.repository, documents: detail.documents)),
                const SizedBox(width: 24),
                Expanded(flex: 4, child: _AuditTimeline(entries: detail.auditLog)),
              ],
            ),
          ),
        ),
        const Divider(height: 1),
        Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.end,
            children: [
              OutlinedButton.icon(
                onPressed: _busy ? null : () => _downloadZip(context),
                icon: const Icon(Icons.download_outlined),
                label: const Text('Download ZIP'),
              ),
              const SizedBox(width: 12),
              OutlinedButton.icon(
                onPressed: _busy ? null : () => _reject(context),
                icon: const Icon(Icons.close),
                label: const Text('Reject'),
                style: OutlinedButton.styleFrom(foregroundColor: const Color(0xFFDC2626)),
              ),
              const SizedBox(width: 12),
              FilledButton.icon(
                onPressed: _busy ? null : () => _approve(context),
                icon: const Icon(Icons.check),
                label: const Text('Approve'),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Future<void> _approve(BuildContext context) async {
    setState(() => _busy = true);
    try {
      await widget.repository.approveDriver(widget.driverId);
      try {
        await widget.repository.approveVehicle(widget.driverId);
      } catch (_) {
        // Vehicle verification may not exist yet — driver approval alone
        // still stands; GoOnline requires both, admin can approve the
        // vehicle separately once it's submitted.
      }
      _changed = true;
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Đã duyệt hồ sơ')));
        Navigator.of(context).pop(true);
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Duyệt thất bại: $e')));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _reject(BuildContext context) async {
    final reason = await RejectReasonDialog.show(context);
    if (reason == null || !context.mounted) return;
    setState(() => _busy = true);
    try {
      await widget.repository.rejectDriver(widget.driverId, reason);
      try {
        await widget.repository.rejectVehicle(widget.driverId, reason);
      } catch (_) {
        // Same best-effort rationale as approve — vehicle record optional.
      }
      _changed = true;
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Đã từ chối hồ sơ')));
        Navigator.of(context).pop(true);
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Từ chối thất bại: $e')));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _downloadZip(BuildContext context) async {
    setState(() => _busy = true);
    try {
      final bytes = await widget.repository.getDocumentsZip(widget.driverId);
      downloadBytes(bytes, '${widget.driverId}-kyc-documents.zip');
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Tải ZIP thất bại: $e')));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }
}

class _InfoPanel extends StatelessWidget {
  const _InfoPanel({required this.dv, required this.vv});

  final DriverVerificationDetail? dv;
  final VehicleVerificationDetail? vv;

  @override
  Widget build(BuildContext context) {
    final dv = this.dv;
    final vv = this.vv;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Thông tin đối chiếu', style: Theme.of(context).textTheme.titleSmall),
        const SizedBox(height: 12),
        if (dv != null) ...[
          _row('Họ tên', dv.fullName),
          _row('Ngày sinh', dv.dateOfBirth),
          _row('Địa chỉ', dv.address),
          _row('Số CCCD', dv.nationalIdNumber),
          _row('Số GPLX', dv.licenseNumber),
          if (dv.rejectReason.isNotEmpty) _row('Lý do từ chối trước', dv.rejectReason, isWarning: true),
          const SizedBox(height: 12),
        ] else
          const Text('Chưa gửi hồ sơ cá nhân', style: TextStyle(color: Color(0xFF9CA3AF))),
        const Divider(height: 24),
        if (vv != null) ...[
          _row('Hãng / Dòng xe', '${vv.brand} ${vv.model} (${vv.year})'),
          _row('Màu xe', vv.color),
          _row('Biển số', vv.plateNumber),
          _row('Số VIN', vv.vin),
          _row('Số máy', vv.engineNumber),
          _row('Số khung', vv.chassisNumber),
          _row('Hạng GPLX', vv.licenseClass),
          if (vv.expiredAt != null) _row('Ngày hết hạn', _dateFmt.format(vv.expiredAt!)),
        ] else
          const Text('Chưa gửi hồ sơ xe', style: TextStyle(color: Color(0xFF9CA3AF))),
      ],
    );
  }

  Widget _row(String label, String value, {bool isWarning = false}) {
    if (value.isEmpty) return const SizedBox.shrink();
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF))),
          Text(value, style: TextStyle(color: isWarning ? const Color(0xFFDC2626) : null)),
        ],
      ),
    );
  }
}

class _DocumentGrid extends StatelessWidget {
  const _DocumentGrid({required this.repository, required this.documents});

  final KYCRepository repository;
  final List<KYCDocumentItem> documents;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Giấy tờ', style: Theme.of(context).textTheme.titleSmall),
        const SizedBox(height: 12),
        GridView.count(
          crossAxisCount: 2,
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          mainAxisSpacing: 16,
          crossAxisSpacing: 16,
          childAspectRatio: 0.8,
          children: [
            for (final d in documents) DocumentThumbnail(repository: repository, item: d),
          ],
        ),
      ],
    );
  }
}

class _AuditTimeline extends StatelessWidget {
  const _AuditTimeline({required this.entries});

  final List<AuditLogEntry> entries;

  static const _actionLabels = {
    'submit': 'Gửi hồ sơ',
    'modify': 'Chỉnh sửa',
    'approve': 'Duyệt',
    'reject': 'Từ chối',
    'expire': 'Hết hạn',
  };

  static const _entityLabels = {
    'driver_verification': 'Hồ sơ cá nhân',
    'vehicle_verification': 'Hồ sơ xe',
    'kyc_document': 'Tài liệu',
  };

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Lịch sử (Audit)', style: Theme.of(context).textTheme.titleSmall),
        const SizedBox(height: 12),
        if (entries.isEmpty)
          const Text('Chưa có lịch sử', style: TextStyle(color: Color(0xFF9CA3AF)))
        else
          ...entries.map((e) => Padding(
                padding: const EdgeInsets.only(bottom: 12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      '${_actionLabels[e.action] ?? e.action} • ${_entityLabels[e.entityType] ?? e.entityType}',
                      style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
                    ),
                    Text(_dateFmt.format(e.createdAt), style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF))),
                    if (e.reason.isNotEmpty) Text(e.reason, style: const TextStyle(fontSize: 12)),
                  ],
                ),
              )),
      ],
    );
  }
}
