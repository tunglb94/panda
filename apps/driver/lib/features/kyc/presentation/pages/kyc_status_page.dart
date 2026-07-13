import 'package:flutter/material.dart';

import '../../../../core/network/api_client.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_button.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../../../shared/widgets/app_loading_view.dart';
import '../../../../shared/widgets/app_status_chip.dart';
import '../../data/kyc_repository.dart';
import '../../domain/models/document_type.dart';
import '../../domain/models/driver_verification.dart';
import '../../domain/models/kyc_status.dart';
import '../../domain/models/vehicle_verification.dart';
import 'become_driver_page.dart';

/// Phần 6/11 — Status UI: shows Progress, the Driver + Vehicle Verification
/// status (Pending/Under Review/Approved/Rejected/Expired, with the admin's
/// reason when Rejected), and a Documents checklist with version + expiry
/// banners (yellow within 30 days, red once expired).
class KYCStatusPage extends StatefulWidget {
  const KYCStatusPage({super.key, required this.apiClient});

  final ApiClient apiClient;

  @override
  State<KYCStatusPage> createState() => _KYCStatusPageState();
}

class _KYCStatusPageState extends State<KYCStatusPage> {
  late final KYCRepository _repo = KYCRepository(widget.apiClient);
  bool _loading = true;
  DriverVerification? _driver;
  VehicleVerification? _vehicle;
  List<DocumentChecklistItem> _documents = [];
  String? _error;

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
      final driver = await _repo.getDriverVerification();
      final vehicle = await _repo.getVehicleVerification();
      List<DocumentChecklistItem> documents = [];
      try {
        documents = await _repo.listDocuments();
      } catch (_) {
        // Best-effort — the page still works without the documents list.
      }
      if (mounted) {
        setState(() {
          _driver = driver;
          _vehicle = vehicle;
          _documents = documents;
          _loading = false;
        });
      }
    } on ApiException catch (e) {
      if (mounted) {
        setState(() {
          _loading = false;
          _error = e.statusCode == 0 ? e.message : 'Không thể tải trạng thái xác minh.';
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() {
          _loading = false;
          _error = 'Không thể tải trạng thái xác minh.';
        });
      }
    }
  }

  void _openWizard() {
    Navigator.of(context)
        .push(MaterialPageRoute(builder: (_) => BecomeDriverPage(apiClient: widget.apiClient)))
        .then((_) => _load());
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Xác minh tài xế')),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: _buildBody(),
          ),
        ),
      ),
    );
  }

  Widget _buildBody() {
    if (_loading) return const AppLoadingView(label: 'Đang tải…');
    if (_error != null) {
      return Padding(
        padding: const EdgeInsets.all(AppSpacing.lg),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(_error!, textAlign: TextAlign.center),
            const SizedBox(height: AppSpacing.md),
            AppButton.outline(label: 'Thử lại', onPressed: _load),
          ],
        ),
      );
    }

    final driver = _driver;
    final vehicle = _vehicle;

    if (driver == null && vehicle == null) {
      return Padding(
        padding: const EdgeInsets.all(AppSpacing.lg),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.badge_outlined, size: 64, color: AppColors.textTertiary),
            const SizedBox(height: AppSpacing.md),
            Text(
              'Bạn chưa gửi hồ sơ xác minh',
              style: Theme.of(context).textTheme.titleMedium,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: AppSpacing.sm),
            Text(
              'Hoàn tất KYC và đăng ký xe để bắt đầu nhận chuyến.',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: AppSpacing.xl),
            AppButton.primary(label: 'Bắt đầu xác minh', onPressed: _openWizard),
          ],
        ),
      );
    }

    final approvedCount = (driver?.status.isApproved ?? false ? 1 : 0) + (vehicle?.status.isApproved ?? false ? 1 : 0);

    return RefreshIndicator(
      onRefresh: _load,
      child: ListView(
        padding: const EdgeInsets.all(AppSpacing.lg),
        children: [
          _ProgressSummary(approvedCount: approvedCount, total: 2),
          const SizedBox(height: AppSpacing.md),
          if (driver != null) _StatusSection(title: 'Hồ sơ cá nhân (KYC)', status: driver.status, rejectReason: driver.rejectReason),
          const SizedBox(height: AppSpacing.md),
          if (vehicle != null) _StatusSection(title: 'Đăng ký xe', status: vehicle.status, rejectReason: vehicle.rejectReason),
          if (_documents.isNotEmpty) ...[
            const SizedBox(height: AppSpacing.lg),
            Text('Giấy tờ', style: Theme.of(context).textTheme.titleSmall),
            const SizedBox(height: AppSpacing.sm),
            _DocumentsSection(documents: _documents),
          ],
          const SizedBox(height: AppSpacing.xl),
          if (driver == null || vehicle == null || driver.status.isEditable || vehicle.status.isEditable)
            AppButton.primary(
              label: (driver == null || vehicle == null) ? 'Tiếp tục hồ sơ' : 'Chỉnh sửa & gửi lại',
              onPressed: _openWizard,
            ),
        ],
      ),
    );
  }
}

class _ProgressSummary extends StatelessWidget {
  const _ProgressSummary({required this.approvedCount, required this.total});

  final int approvedCount;
  final int total;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text('Tiến độ xác minh', style: Theme.of(context).textTheme.titleSmall),
            Text('$approvedCount/$total đã duyệt', style: Theme.of(context).textTheme.bodySmall),
          ],
        ),
        const SizedBox(height: 6),
        ClipRRect(
          borderRadius: BorderRadius.circular(4),
          child: LinearProgressIndicator(
            value: total == 0 ? 0 : approvedCount / total,
            minHeight: 6,
            backgroundColor: AppColors.divider,
            valueColor: const AlwaysStoppedAnimation(AppColors.primary),
          ),
        ),
      ],
    );
  }
}

class _StatusSection extends StatelessWidget {
  const _StatusSection({required this.title, required this.status, required this.rejectReason});

  final String title;
  final KYCStatus status;
  final String rejectReason;

  Color get _color => switch (status) {
        KYCStatus.approved => AppColors.primary,
        KYCStatus.rejected => AppColors.error,
        KYCStatus.underReview => AppColors.info,
        KYCStatus.expired => AppColors.error,
        KYCStatus.pending => AppColors.warning,
      };

  IconData get _icon => switch (status) {
        KYCStatus.approved => Icons.verified,
        KYCStatus.rejected => Icons.cancel_outlined,
        KYCStatus.underReview => Icons.search,
        KYCStatus.expired => Icons.event_busy_outlined,
        KYCStatus.pending => Icons.hourglass_empty,
      };

  @override
  Widget build(BuildContext context) {
    return AppCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Expanded(child: Text(title, style: Theme.of(context).textTheme.titleSmall)),
              AppStatusChip(label: status.label, color: _color, icon: _icon),
            ],
          ),
          if ((status == KYCStatus.rejected || status == KYCStatus.expired) && rejectReason.isNotEmpty) ...[
            const SizedBox(height: AppSpacing.sm),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(AppSpacing.sm),
              decoration: BoxDecoration(
                color: AppColors.error.withValues(alpha: 0.08),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                'Lý do: $rejectReason',
                style: const TextStyle(color: AppColors.error, fontSize: 13),
              ),
            ),
          ],
        ],
      ),
    );
  }
}

/// Phần 11 — Documents checklist: version + expiry, yellow banner within 30
/// days of expiring, red banner once expired.
class _DocumentsSection extends StatelessWidget {
  const _DocumentsSection({required this.documents});

  final List<DocumentChecklistItem> documents;

  @override
  Widget build(BuildContext context) {
    final uploadedDocs = documents.where((d) => d.uploaded).toList();
    if (uploadedDocs.isEmpty) return const SizedBox.shrink();
    return AppCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          for (var i = 0; i < uploadedDocs.length; i++) ...[
            if (i > 0) const Divider(height: AppSpacing.lg),
            _DocumentRow(item: uploadedDocs[i]),
          ],
        ],
      ),
    );
  }
}

class _DocumentRow extends StatelessWidget {
  const _DocumentRow({required this.item});

  final DocumentChecklistItem item;

  String _dateText(DateTime d) => '${d.day.toString().padLeft(2, '0')}/${d.month.toString().padLeft(2, '0')}/${d.year}';

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            const Icon(Icons.check_circle, size: 16, color: AppColors.primary),
            const SizedBox(width: 6),
            Expanded(child: Text(documentTypeLabel(item.documentType), style: Theme.of(context).textTheme.bodyMedium)),
            if (item.version > 1)
              Text('v${item.version}', style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppColors.textTertiary)),
          ],
        ),
        if (item.expiresAt != null) ...[
          const SizedBox(height: 6),
          if (item.expired || item.expiringSoon)
            Container(
              padding: const EdgeInsets.symmetric(horizontal: AppSpacing.sm, vertical: 6),
              decoration: BoxDecoration(
                color: (item.expired ? AppColors.error : AppColors.warning).withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(6),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(
                    item.expired ? Icons.error_outline : Icons.warning_amber_rounded,
                    size: 14,
                    color: item.expired ? AppColors.error : AppColors.warning,
                  ),
                  const SizedBox(width: 4),
                  Text(
                    item.expired ? 'Đã hết hạn: ${_dateText(item.expiresAt!)}' : 'Sắp hết hạn: ${_dateText(item.expiresAt!)}',
                    style: TextStyle(
                      color: item.expired ? AppColors.error : AppColors.warning,
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
              ),
            )
          else
            Text(
              'Hết hạn: ${_dateText(item.expiresAt!)}',
              style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppColors.textSecondary),
            ),
        ],
      ],
    );
  }
}
