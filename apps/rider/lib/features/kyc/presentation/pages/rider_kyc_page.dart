import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';

import '../../../../core/network/api_client.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_radius.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_bottom_sheet.dart';
import '../../../../shared/widgets/app_button.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../../../shared/widgets/app_loading_view.dart';
import '../../../../shared/widgets/app_status_chip.dart';
import '../../data/kyc_repository.dart';
import '../../domain/models/rider_verification.dart';

/// Rider KYC — the whole "Onboarding" gate the Auth spec requires before a
/// new rider can reach Home: Họ tên, Ngày sinh, CCCD number, and the two
/// CCCD photos, all on one screen (this app's version of Grab/Be's
/// single-screen ID verification). Approval is manual (no Admin dashboard
/// in this phase — see plan's Known Gaps), so after a valid submit the
/// rider sees "Đang chờ duyệt" here until an admin approves via the
/// RequireAdmin-gated API.
class RiderKycPage extends StatefulWidget {
  const RiderKycPage({super.key, required this.apiClient});

  final ApiClient apiClient;

  @override
  State<RiderKycPage> createState() => _RiderKycPageState();
}

class _RiderKycPageState extends State<RiderKycPage> {
  late final KYCRepository _repo = KYCRepository(widget.apiClient);

  final _nameCtrl = TextEditingController();
  final _idCtrl = TextEditingController();
  DateTime? _dob;

  bool _loading = true;
  bool _submitting = false;
  bool _frontUploaded = false;
  bool _backUploaded = false;
  RiderVerification? _existing;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _nameCtrl.dispose();
    _idCtrl.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final v = await _repo.getVerification();
      if (mounted) {
        setState(() {
          _existing = v;
          _frontUploaded = v?.cccdFrontUploaded ?? false;
          _backUploaded = v?.cccdBackUploaded ?? false;
          if (v != null && v.status != RiderKYCStatus.pending) {
            _nameCtrl.text = v.fullName;
            _idCtrl.text = v.nationalIdNumber;
            _dob = v.dateOfBirth;
          }
          _loading = false;
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

  bool get _isEditable => _existing == null || _existing!.status == RiderKYCStatus.rejected;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Xác minh danh tính')),
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

    final existing = _existing;
    if (existing != null && existing.status == RiderKYCStatus.pending) {
      return _PendingView(onRefresh: _load);
    }
    if (existing != null && existing.status == RiderKYCStatus.approved) {
      return const _ApprovedView();
    }

    // Not submitted yet, or Rejected (editable) — show the form.
    return ListView(
      padding: const EdgeInsets.all(AppSpacing.lg),
      children: [
        if (existing != null && existing.status == RiderKYCStatus.rejected)
          Padding(
            padding: const EdgeInsets.only(bottom: AppSpacing.md),
            child: _RejectedBanner(reason: existing.rejectReason),
          ),
        Text(
          'Vui lòng cung cấp thông tin để xác minh danh tính trước khi đặt chuyến.',
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary),
        ),
        const SizedBox(height: AppSpacing.lg),
        TextField(
          controller: _nameCtrl,
          enabled: _isEditable && !_submitting,
          textCapitalization: TextCapitalization.words,
          decoration: const InputDecoration(
            labelText: 'Họ và tên',
            border: OutlineInputBorder(borderRadius: AppRadius.mdAll),
          ),
        ),
        const SizedBox(height: AppSpacing.md),
        _DateOfBirthField(
          value: _dob,
          enabled: _isEditable && !_submitting,
          onChanged: (d) => setState(() => _dob = d),
        ),
        const SizedBox(height: AppSpacing.md),
        TextField(
          controller: _idCtrl,
          enabled: _isEditable && !_submitting,
          keyboardType: TextInputType.number,
          decoration: const InputDecoration(
            labelText: 'Số CCCD',
            border: OutlineInputBorder(borderRadius: AppRadius.mdAll),
          ),
        ),
        const SizedBox(height: AppSpacing.lg),
        _PhotoTile(
          label: 'Ảnh CCCD mặt trước',
          documentType: 'cccd_front',
          repo: _repo,
          uploaded: _frontUploaded,
          enabled: _isEditable && !_submitting,
          onUploaded: () => setState(() => _frontUploaded = true),
        ),
        const SizedBox(height: AppSpacing.sm),
        _PhotoTile(
          label: 'Ảnh CCCD mặt sau',
          documentType: 'cccd_back',
          repo: _repo,
          uploaded: _backUploaded,
          enabled: _isEditable && !_submitting,
          onUploaded: () => setState(() => _backUploaded = true),
        ),
        if (_error != null) ...[
          const SizedBox(height: AppSpacing.md),
          Text(_error!, style: const TextStyle(color: AppColors.error), textAlign: TextAlign.center),
        ],
        const SizedBox(height: AppSpacing.xl),
        if (_isEditable)
          AppButton.primary(
            label: 'Gửi xác minh',
            isLoading: _submitting,
            onPressed: _submit,
          ),
      ],
    );
  }

  Future<void> _submit() async {
    if (_submitting) return;
    final name = _nameCtrl.text.trim();
    final id = _idCtrl.text.trim();
    if (name.isEmpty || id.isEmpty || _dob == null) {
      setState(() => _error = 'Vui lòng điền đầy đủ thông tin.');
      return;
    }
    if (!_frontUploaded || !_backUploaded) {
      setState(() => _error = 'Vui lòng tải lên cả hai ảnh CCCD.');
      return;
    }

    setState(() {
      _submitting = true;
      _error = null;
    });
    try {
      final v = await _repo.submit(fullName: name, dateOfBirth: _dob!, nationalIdNumber: id);
      if (mounted) setState(() => _existing = v);
    } on ApiException catch (e) {
      if (mounted) {
        setState(() => _error = e.statusCode == 0 ? e.message : e.message);
      }
    } catch (_) {
      if (mounted) {
        setState(() => _error = 'Gửi xác minh thất bại. Vui lòng thử lại.');
      }
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }
}

class _DateOfBirthField extends StatelessWidget {
  const _DateOfBirthField({required this.value, required this.enabled, required this.onChanged});

  final DateTime? value;
  final bool enabled;
  final ValueChanged<DateTime?> onChanged;

  String _text(DateTime d) => '${d.day.toString().padLeft(2, '0')}/${d.month.toString().padLeft(2, '0')}/${d.year}';

  @override
  Widget build(BuildContext context) {
    return InkWell(
      borderRadius: AppRadius.mdAll,
      onTap: !enabled
          ? null
          : () async {
              final now = DateTime.now();
              final picked = await showDatePicker(
                context: context,
                initialDate: value ?? DateTime(now.year - 20),
                firstDate: DateTime(now.year - 100),
                lastDate: now,
                helpText: 'Ngày sinh',
              );
              if (picked != null) onChanged(picked);
            },
      child: InputDecorator(
        decoration: const InputDecoration(
          labelText: 'Ngày sinh',
          border: OutlineInputBorder(borderRadius: AppRadius.mdAll),
          suffixIcon: Icon(Icons.calendar_today_outlined),
        ),
        child: Text(value == null ? 'Chọn ngày sinh' : _text(value!)),
      ),
    );
  }
}

class _PhotoTile extends StatefulWidget {
  const _PhotoTile({
    required this.label,
    required this.documentType,
    required this.repo,
    required this.uploaded,
    required this.enabled,
    required this.onUploaded,
  });

  final String label;
  final String documentType;
  final KYCRepository repo;
  final bool uploaded;
  final bool enabled;
  final VoidCallback onUploaded;

  @override
  State<_PhotoTile> createState() => _PhotoTileState();
}

class _PhotoTileState extends State<_PhotoTile> {
  bool _uploading = false;
  String? _error;

  Future<void> _pickAndUpload(ImageSource source) async {
    final picker = ImagePicker();
    XFile? file;
    try {
      file = await picker.pickImage(source: source, imageQuality: 85);
    } catch (_) {
      if (mounted) setState(() => _error = 'Không thể mở camera/thư viện ảnh.');
      return;
    }
    if (file == null) return;

    setState(() {
      _uploading = true;
      _error = null;
    });
    try {
      final bytes = await file.readAsBytes();
      await widget.repo.uploadDocument(documentType: widget.documentType, bytes: bytes, filename: file.name);
      if (mounted) {
        setState(() => _uploading = false);
        widget.onUploaded();
      }
    } on ApiException catch (e) {
      if (mounted) {
        setState(() {
          _uploading = false;
          _error = e.statusCode == 0 ? e.message : 'Tải lên thất bại. Vui lòng thử lại.';
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() {
          _uploading = false;
          _error = 'Tải lên thất bại. Vui lòng thử lại.';
        });
      }
    }
  }

  void _showSourceSheet() {
    AppBottomSheet.show<void>(
      context,
      title: widget.label,
      builder: (sheetContext) => Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          ListTile(
            leading: const Icon(Icons.photo_camera_outlined),
            title: const Text('Chụp ảnh'),
            onTap: () {
              Navigator.pop(sheetContext);
              _pickAndUpload(ImageSource.camera);
            },
          ),
          ListTile(
            leading: const Icon(Icons.photo_library_outlined),
            title: const Text('Chọn từ thư viện'),
            onTap: () {
              Navigator.pop(sheetContext);
              _pickAndUpload(ImageSource.gallery);
            },
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final uploaded = widget.uploaded;
    return Semantics(
      label: '${widget.label}, ${uploaded ? "đã tải lên" : "chưa tải lên"}',
      button: true,
      child: AppCard(
        animateIn: false,
        padding: const EdgeInsets.all(AppSpacing.md),
        onTap: (!widget.enabled || _uploading) ? null : _showSourceSheet,
        child: Row(
          children: [
            Container(
              width: 44,
              height: 44,
              decoration: BoxDecoration(
                color: uploaded ? AppColors.primaryLight : AppColors.surfaceAlt,
                shape: BoxShape.circle,
              ),
              child: Icon(
                uploaded ? Icons.check_circle : Icons.upload_file_outlined,
                color: uploaded ? AppColors.primary : AppColors.textSecondary,
              ),
            ),
            const SizedBox(width: AppSpacing.md),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(widget.label, style: Theme.of(context).textTheme.bodyMedium),
                  if (_error != null)
                    Padding(
                      padding: const EdgeInsets.only(top: 2),
                      child: Text(_error!, style: const TextStyle(color: AppColors.error, fontSize: 12)),
                    ),
                ],
              ),
            ),
            if (_uploading)
              const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2))
            else
              Icon(uploaded ? Icons.check : Icons.chevron_right, color: uploaded ? AppColors.primary : AppColors.textTertiary),
          ],
        ),
      ),
    );
  }
}

class _PendingView extends StatelessWidget {
  const _PendingView({required this.onRefresh});

  final Future<void> Function() onRefresh;

  @override
  Widget build(BuildContext context) {
    return RefreshIndicator(
      onRefresh: onRefresh,
      child: ListView(
        padding: const EdgeInsets.all(AppSpacing.lg),
        children: [
          const SizedBox(height: AppSpacing.xxxl),
          const Icon(Icons.hourglass_empty, size: 64, color: AppColors.warning),
          const SizedBox(height: AppSpacing.md),
          Center(
            child: AppStatusChip(label: 'Đang chờ duyệt', color: AppColors.warning, icon: Icons.hourglass_empty),
          ),
          const SizedBox(height: AppSpacing.md),
          Text(
            'Hồ sơ của bạn đã được gửi và đang chờ quản trị viên duyệt. Vui lòng quay lại sau.',
            textAlign: TextAlign.center,
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary),
          ),
          const SizedBox(height: AppSpacing.lg),
          Center(child: AppButton.outline(label: 'Kiểm tra lại', expand: false, onPressed: onRefresh)),
        ],
      ),
    );
  }
}

class _ApprovedView extends StatelessWidget {
  const _ApprovedView();

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.lg),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.verified, size: 64, color: AppColors.primary),
            const SizedBox(height: AppSpacing.md),
            Text('Đã xác minh', style: Theme.of(context).textTheme.titleMedium),
          ],
        ),
      ),
    );
  }
}

class _RejectedBanner extends StatelessWidget {
  const _RejectedBanner({required this.reason});

  final String reason;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(AppSpacing.md),
      decoration: BoxDecoration(
        color: AppColors.error.withValues(alpha: 0.08),
        borderRadius: AppRadius.mdAll,
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Icon(Icons.error_outline, color: AppColors.error, size: 18),
          const SizedBox(width: AppSpacing.sm),
          Expanded(
            child: Text(
              reason.isEmpty ? 'Hồ sơ bị từ chối. Vui lòng gửi lại.' : 'Bị từ chối: $reason',
              style: const TextStyle(color: AppColors.error, fontSize: 13),
            ),
          ),
        ],
      ),
    );
  }
}
