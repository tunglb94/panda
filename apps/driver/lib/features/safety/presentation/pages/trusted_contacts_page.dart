import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_bottom_sheet.dart';
import '../../../../shared/widgets/app_button.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../../../shared/widgets/app_dialog.dart';
import '../../../../shared/widgets/app_empty_state.dart';
import '../../../../shared/widgets/app_snackbar.dart';
import '../../domain/models/trusted_contact.dart';

/// Trusted Contact — UI only, exactly as scoped. There is no
/// trusted-contact backend, so the list starts empty and every add/edit/
/// delete only mutates in-memory state for this page's lifetime; nothing is
/// synced or persisted, and a fresh app launch starts from empty again.
class TrustedContactsPage extends StatefulWidget {
  const TrustedContactsPage({super.key});

  @override
  State<TrustedContactsPage> createState() => _TrustedContactsPageState();
}

class _TrustedContactsPageState extends State<TrustedContactsPage> {
  final List<TrustedContact> _contacts = [];
  int _nextId = 1;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Liên hệ tin cậy'),
        actions: [
          IconButton(
            icon: const Icon(Icons.person_add_alt_outlined),
            tooltip: 'Thêm liên hệ',
            onPressed: () => _openForm(context),
          ),
        ],
      ),
      body: _contacts.isEmpty
          ? AppEmptyState(
              icon: Icons.contacts_outlined,
              title: 'Chưa có liên hệ tin cậy',
              subtitle: 'Thêm người thân hoặc bạn bè để họ có thể được liên hệ '
                  'khi cần trong tình huống khẩn cấp.',
              actionLabel: 'Thêm liên hệ',
              onAction: () => _openForm(context),
            )
          : ListView.separated(
              padding: const EdgeInsets.all(AppSpacing.lg),
              itemCount: _contacts.length,
              separatorBuilder: (_, _) => const SizedBox(height: AppSpacing.sm),
              itemBuilder: (context, i) {
                final contact = _contacts[i];
                return AppCard(
                  padding: const EdgeInsets.all(AppSpacing.md),
                  onTap: () => _openForm(context, existing: contact),
                  child: Row(
                    children: [
                      Container(
                        width: 44,
                        height: 44,
                        decoration: const BoxDecoration(
                          color: AppColors.primaryLight,
                          shape: BoxShape.circle,
                        ),
                        child: const Icon(Icons.person, color: AppColors.primary, size: AppIconSize.lg),
                      ),
                      const SizedBox(width: AppSpacing.md),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(contact.name, style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                                  fontWeight: FontWeight.w700,
                                )),
                            Text(
                              '${contact.relationship} · ${contact.phone}',
                              style: Theme.of(context).textTheme.bodySmall,
                            ),
                          ],
                        ),
                      ),
                      IconButton(
                        icon: const Icon(Icons.delete_outline, color: AppColors.error),
                        tooltip: 'Xóa',
                        onPressed: () => _confirmDelete(context, contact),
                      ),
                    ],
                  ),
                );
              },
            ),
    );
  }

  Future<void> _confirmDelete(BuildContext context, TrustedContact contact) async {
    final confirmed = await AppDialog.confirm(
      context,
      title: 'Xóa liên hệ này?',
      message: '${contact.name} sẽ bị xóa khỏi danh sách liên hệ tin cậy.',
      confirmLabel: 'Xóa',
      isDestructive: true,
    );
    if (confirmed && mounted) {
      setState(() => _contacts.removeWhere((c) => c.id == contact.id));
      if (context.mounted) AppSnackbar.show(context, 'Đã xóa liên hệ.');
    }
  }

  Future<void> _openForm(BuildContext context, {TrustedContact? existing}) async {
    final result = await AppBottomSheet.show<TrustedContact>(
      context,
      title: existing == null ? 'Thêm liên hệ tin cậy' : 'Sửa liên hệ',
      isScrollControlled: true,
      builder: (sheetContext) => _ContactForm(existing: existing),
    );
    if (result == null || !mounted) return;
    setState(() {
      if (existing == null) {
        _contacts.add(TrustedContact(
          id: '${_nextId++}',
          name: result.name,
          phone: result.phone,
          relationship: result.relationship,
        ));
      } else {
        final i = _contacts.indexWhere((c) => c.id == existing.id);
        if (i != -1) {
          _contacts[i] = existing.copyWith(
            name: result.name,
            phone: result.phone,
            relationship: result.relationship,
          );
        }
      }
    });
  }
}

class _ContactForm extends StatefulWidget {
  const _ContactForm({this.existing});

  final TrustedContact? existing;

  @override
  State<_ContactForm> createState() => _ContactFormState();
}

class _ContactFormState extends State<_ContactForm> {
  late final _nameController = TextEditingController(text: widget.existing?.name);
  late final _phoneController = TextEditingController(text: widget.existing?.phone);
  late final _relationshipController = TextEditingController(text: widget.existing?.relationship);

  @override
  void dispose() {
    _nameController.dispose();
    _phoneController.dispose();
    _relationshipController.dispose();
    super.dispose();
  }

  bool get _isValid => _nameController.text.trim().isNotEmpty && _phoneController.text.trim().isNotEmpty;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        TextField(
          controller: _nameController,
          decoration: const InputDecoration(labelText: 'Họ tên'),
          onChanged: (_) => setState(() {}),
        ),
        const SizedBox(height: AppSpacing.md),
        TextField(
          controller: _phoneController,
          keyboardType: TextInputType.phone,
          decoration: const InputDecoration(labelText: 'Số điện thoại'),
          onChanged: (_) => setState(() {}),
        ),
        const SizedBox(height: AppSpacing.md),
        TextField(
          controller: _relationshipController,
          decoration: const InputDecoration(labelText: 'Mối quan hệ (VD: Vợ, Anh trai…)'),
        ),
        const SizedBox(height: AppSpacing.lg),
        AppButton.primary(
          label: widget.existing == null ? 'Thêm liên hệ' : 'Lưu thay đổi',
          onPressed: _isValid
              ? () => Navigator.pop(
                    context,
                    TrustedContact(
                      id: widget.existing?.id ?? '',
                      name: _nameController.text.trim(),
                      phone: _phoneController.text.trim(),
                      relationship: _relationshipController.text.trim().isEmpty
                          ? 'Chưa cập nhật'
                          : _relationshipController.text.trim(),
                    ),
                  )
              : null,
        ),
      ],
    );
  }
}
