import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';

import '../../../../core/network/api_client.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_bottom_sheet.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../data/kyc_repository.dart';

/// One document-upload row (CCCD front/back, selfie, GPLX, vehicle
/// registration, insurance, đăng kiểm) — tap to choose Camera or Gallery,
/// uploads immediately via [KYCRepository.uploadDocument]. Shows a
/// checkmark once uploaded; re-tapping re-uploads as a NEW version (Phần 4
/// — the previous version and file are never overwritten, only superseded).
///
/// When [requiresExpiry] is true (GPLX/registration/insurance/inspection —
/// Phần 2), picking a file also prompts for its expiry date before
/// uploading, and the tile shows an expiry banner: yellow within 30 days of
/// expiring, red once expired (Phần 11).
class DocumentUploadTile extends StatefulWidget {
  const DocumentUploadTile({
    super.key,
    required this.label,
    required this.documentType,
    required this.repository,
    required this.uploaded,
    this.requiresExpiry = false,
    this.expiresAt,
    this.expired = false,
    this.expiringSoon = false,
    this.version = 0,
    this.onUploaded,
  });

  final String label;
  final String documentType;
  final KYCRepository repository;
  final bool uploaded;
  final bool requiresExpiry;
  final DateTime? expiresAt;
  final bool expired;
  final bool expiringSoon;
  final int version;
  final VoidCallback? onUploaded;

  @override
  State<DocumentUploadTile> createState() => _DocumentUploadTileState();
}

class _DocumentUploadTileState extends State<DocumentUploadTile> {
  bool _uploading = false;
  late bool _uploaded = widget.uploaded;
  String? _error;

  @override
  void didUpdateWidget(covariant DocumentUploadTile oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.uploaded != oldWidget.uploaded) {
      _uploaded = widget.uploaded;
    }
  }

  Future<DateTime?> _pickExpiryDate() async {
    final now = DateTime.now();
    return showDatePicker(
      context: context,
      initialDate: widget.expiresAt ?? now.add(const Duration(days: 365)),
      firstDate: now,
      lastDate: DateTime(now.year + 20),
      helpText: 'Ngày hết hạn',
    );
  }

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

    DateTime? expiresAt;
    if (widget.requiresExpiry) {
      expiresAt = await _pickExpiryDate();
      if (expiresAt == null) return; // user cancelled — abort, don't upload
    }

    setState(() {
      _uploading = true;
      _error = null;
    });
    try {
      final bytes = await file.readAsBytes();
      await widget.repository.uploadDocument(
        documentType: widget.documentType,
        bytes: bytes,
        filename: file.name,
        expiresAt: expiresAt,
      );
      if (mounted) {
        setState(() {
          _uploaded = true;
          _uploading = false;
        });
        widget.onUploaded?.call();
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

  String get _expiryText {
    final d = widget.expiresAt;
    if (d == null) return '';
    return '${d.day.toString().padLeft(2, '0')}/${d.month.toString().padLeft(2, '0')}/${d.year}';
  }

  @override
  Widget build(BuildContext context) {
    return Semantics(
      label: '${widget.label}, ${_uploaded ? "đã tải lên" : "chưa tải lên"}'
          '${widget.expired ? ", đã hết hạn" : widget.expiringSoon ? ", sắp hết hạn" : ""}',
      button: true,
      child: AppCard(
        animateIn: false,
        padding: const EdgeInsets.all(AppSpacing.md),
        onTap: _uploading ? null : _showSourceSheet,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  width: 44,
                  height: 44,
                  decoration: BoxDecoration(
                    color: _uploaded ? AppColors.primaryLight : AppColors.surfaceAlt,
                    shape: BoxShape.circle,
                  ),
                  child: Icon(
                    _uploaded ? Icons.check_circle : Icons.upload_file_outlined,
                    color: _uploaded ? AppColors.primary : AppColors.textSecondary,
                    size: AppIconSize.md,
                  ),
                ),
                const SizedBox(width: AppSpacing.md),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Text(widget.label, style: Theme.of(context).textTheme.bodyMedium),
                          if (_uploaded && widget.version > 1) ...[
                            const SizedBox(width: 6),
                            Text('v${widget.version}',
                                style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppColors.textTertiary)),
                          ],
                        ],
                      ),
                      if (_error != null)
                        Padding(
                          padding: const EdgeInsets.only(top: 2),
                          child: Text(_error!, style: const TextStyle(color: AppColors.error, fontSize: 12)),
                        ),
                    ],
                  ),
                ),
                if (_uploading)
                  const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                else
                  Tooltip(
                    message: _uploaded ? 'Đã tải lên — nhấn để thay ảnh khác' : 'Nhấn để tải lên',
                    child: Icon(
                      _uploaded ? Icons.check : Icons.chevron_right,
                      color: _uploaded ? AppColors.primary : AppColors.textTertiary,
                    ),
                  ),
              ],
            ),
            if (widget.expiresAt != null) ...[
              const SizedBox(height: AppSpacing.sm),
              _ExpiryBanner(expired: widget.expired, expiringSoon: widget.expiringSoon, expiryText: _expiryText),
            ],
          ],
        ),
      ),
    );
  }
}

/// Phần 11 — red once expired, yellow within 30 days, nothing otherwise.
class _ExpiryBanner extends StatelessWidget {
  const _ExpiryBanner({required this.expired, required this.expiringSoon, required this.expiryText});

  final bool expired;
  final bool expiringSoon;
  final String expiryText;

  @override
  Widget build(BuildContext context) {
    if (!expired && !expiringSoon) {
      return Text(
        'Hết hạn: $expiryText',
        style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppColors.textSecondary),
      );
    }
    final color = expired ? AppColors.error : AppColors.warning;
    final label = expired ? 'Đã hết hạn: $expiryText' : 'Sắp hết hạn: $expiryText';
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: AppSpacing.sm, vertical: 6),
      decoration: BoxDecoration(color: color.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(6)),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(expired ? Icons.error_outline : Icons.warning_amber_rounded, size: 14, color: color),
          const SizedBox(width: 4),
          Flexible(child: Text(label, style: TextStyle(color: color, fontSize: 12, fontWeight: FontWeight.w600))),
        ],
      ),
    );
  }
}
