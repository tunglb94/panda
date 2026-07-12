import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_radius.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_card.dart';

class _ChecklistGroup {
  const _ChecklistGroup(this.icon, this.title, this.items);
  final IconData icon;
  final String title;
  final List<String> items;
}

const _groups = [
  _ChecklistGroup(Icons.nightlight_outlined, 'Ban đêm', [
    'Bật đèn xe đầy đủ, kiểm tra đèn hậu còn hoạt động',
    'Đỗ xe nơi có ánh sáng khi chờ khách',
    'Tránh nhận cuốc ở khu vực không quen thuộc, vắng người',
  ]),
  _ChecklistGroup(Icons.cloud_outlined, 'Thời tiết', [
    'Giảm tốc độ khi trời mưa hoặc đường trơn trượt',
    'Kiểm tra gạt mưa và lốp xe trước khi xuất phát',
    'Bật đèn sương mù khi tầm nhìn hạn chế',
  ]),
  _ChecklistGroup(Icons.person_pin_circle_outlined, 'Đón khách', [
    'Xác nhận đúng tên khách trước khi cho lên xe',
    'Đón/trả khách ở nơi an toàn, tránh làn đường ngược chiều',
    'Giữ khoảng cách lịch sự, chuyên nghiệp với khách hàng',
  ]),
  _ChecklistGroup(Icons.payments_outlined, 'Thanh toán', [
    'Xác nhận số tiền trên ứng dụng trước khi kết thúc chuyến',
    'Không giao dịch tiền mặt lớn khi đang di chuyển',
    'Báo cáo ngay nếu phát hiện thanh toán bất thường',
  ]),
];

/// Safety Tips checklist — static, app-authored safety guidance grouped by
/// scenario. This is app content (like the Driver Level tier benefits
/// shown in the Earnings sprint), not driver-specific data, so it's shown
/// in full rather than as a placeholder. Checkmarks are local UI state only
/// (not persisted) — ticking items off is just a personal reading aid.
class SafetyChecklist extends StatefulWidget {
  const SafetyChecklist({super.key});

  @override
  State<SafetyChecklist> createState() => _SafetyChecklistState();
}

class _SafetyChecklistState extends State<SafetyChecklist> {
  final Set<String> _checked = {};

  @override
  Widget build(BuildContext context) {
    return AppCard(
      padding: const EdgeInsets.all(AppSpacing.lg),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.checklist_outlined, color: AppColors.primary, size: AppIconSize.lg),
              const SizedBox(width: AppSpacing.sm),
              Expanded(
                child: Text('Cẩm nang an toàn', style: Theme.of(context).textTheme.titleMedium),
              ),
            ],
          ),
          const SizedBox(height: AppSpacing.md),
          for (var g = 0; g < _groups.length; g++) ...[
            if (g != 0) const Divider(height: AppSpacing.xl),
            _GroupSection(
              group: _groups[g],
              checked: _checked,
              onToggle: (key) => setState(() {
                if (!_checked.remove(key)) _checked.add(key);
              }),
            ),
          ],
        ],
      ),
    );
  }
}

class _GroupSection extends StatelessWidget {
  const _GroupSection({required this.group, required this.checked, required this.onToggle});

  final _ChecklistGroup group;
  final Set<String> checked;
  final ValueChanged<String> onToggle;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(group.icon, size: AppIconSize.md, color: AppColors.textSecondary),
            const SizedBox(width: AppSpacing.sm),
            Text(
              group.title,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w700),
            ),
          ],
        ),
        const SizedBox(height: AppSpacing.sm),
        for (final item in group.items)
          _ChecklistItem(
            label: item,
            checked: checked.contains('${group.title}|$item'),
            onTap: () => onToggle('${group.title}|$item'),
          ),
      ],
    );
  }
}

class _ChecklistItem extends StatelessWidget {
  const _ChecklistItem({required this.label, required this.checked, required this.onTap});

  final String label;
  final bool checked;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: AppRadius.smAll,
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 6),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              AnimatedContainer(
                duration: const Duration(milliseconds: 180),
                width: 20,
                height: 20,
                margin: const EdgeInsets.only(top: 1),
                decoration: BoxDecoration(
                  color: checked ? AppColors.primary : Colors.transparent,
                  border: Border.all(
                    color: checked ? AppColors.primary : AppColors.border,
                    width: 1.6,
                  ),
                  borderRadius: AppRadius.smAll,
                ),
                child: checked
                    ? const Icon(Icons.check, size: 14, color: AppColors.textOnPrimary)
                    : null,
              ),
              const SizedBox(width: AppSpacing.sm),
              Expanded(
                child: Text(
                  label,
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: checked ? AppColors.textTertiary : AppColors.textPrimary,
                        decoration: checked ? TextDecoration.lineThrough : null,
                      ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
