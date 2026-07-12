import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_radius.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../../../shared/widgets/charts/progress_ring.dart';

class _Tier {
  const _Tier(this.name, this.color, this.icon, this.benefits);
  final String name;
  final Color color;
  final IconData icon;
  final List<String> benefits;
}

const _tiers = [
  _Tier('Đồng', Color(0xFFB08D57), Icons.shield_outlined, [
    'Ưu tiên hỗ trợ tiêu chuẩn',
  ]),
  _Tier('Bạc', Color(0xFF9CA3AF), Icons.shield_outlined, [
    'Ưu tiên hỗ trợ tiêu chuẩn',
    'Ưu tiên nhận chuyến giờ cao điểm',
  ]),
  _Tier('Vàng', Color(0xFFD4A017), Icons.shield, [
    'Ưu tiên hỗ trợ tiêu chuẩn',
    'Ưu tiên nhận chuyến giờ cao điểm',
    'Thưởng thêm theo tuần',
  ]),
  _Tier('Kim Cương', Color(0xFF7C3AED), Icons.diamond_outlined, [
    'Ưu tiên hỗ trợ 24/7',
    'Ưu tiên nhận chuyến giờ cao điểm',
    'Thưởng thêm theo tuần',
    'Huy hiệu tài xế Kim Cương',
  ]),
];

/// Driver Level / tiers section — **UI structure only**, exactly as scoped.
/// There is no rank/tier computation anywhere in the backend, so this
/// deliberately never claims the driver is currently at any particular
/// tier — the progress ring shows the honest "no data" state (see
/// `ProgressRing`'s `value: null` path) and the banner says outright that
/// this is a preview. What's real: the tier *structure* itself (name,
/// color, benefit list) — that's app configuration, not personal data, so
/// showing it isn't fabrication.
class DriverLevelCard extends StatefulWidget {
  const DriverLevelCard({super.key});

  @override
  State<DriverLevelCard> createState() => _DriverLevelCardState();
}

class _DriverLevelCardState extends State<DriverLevelCard> {
  int _selectedTier = 0;

  @override
  Widget build(BuildContext context) {
    final tier = _tiers[_selectedTier];
    return AppCard(
      padding: const EdgeInsets.all(AppSpacing.xl),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('Hạng tài xế', style: Theme.of(context).textTheme.titleMedium),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                decoration: BoxDecoration(
                  color: AppColors.surfaceAlt,
                  borderRadius: AppRadius.pillAll,
                  border: Border.all(color: AppColors.border),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.visibility_outlined, size: 12, color: AppColors.textSecondary),
                    const SizedBox(width: 4),
                    Text(
                      'Xem trước tính năng',
                      style: TextStyle(fontSize: 10, color: AppColors.textSecondary, fontWeight: FontWeight.w600),
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: AppSpacing.lg),
          Row(
            children: [
              const ProgressRing(value: null, centerLabel: ''),
              const SizedBox(width: AppSpacing.lg),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Chưa có dữ liệu xếp hạng',
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w700),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      'Hệ thống xếp hạng sẽ tính dựa trên số chuyến, đánh giá và '
                      'thời gian hoạt động khi ra mắt.',
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: AppSpacing.xl),
          SizedBox(
            height: 84,
            child: ListView.separated(
              scrollDirection: Axis.horizontal,
              itemCount: _tiers.length,
              separatorBuilder: (_, _) => const SizedBox(width: AppSpacing.sm),
              itemBuilder: (context, i) {
                final t = _tiers[i];
                final isSelected = i == _selectedTier;
                return GestureDetector(
                  onTap: () => setState(() => _selectedTier = i),
                  child: AnimatedContainer(
                    duration: const Duration(milliseconds: 200),
                    width: 72,
                    padding: const EdgeInsets.symmetric(vertical: AppSpacing.sm),
                    decoration: BoxDecoration(
                      color: isSelected ? t.color.withValues(alpha: 0.12) : AppColors.surfaceAlt,
                      borderRadius: AppRadius.mdAll,
                      border: Border.all(color: isSelected ? t.color : AppColors.border),
                    ),
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(t.icon, color: t.color, size: 22),
                        const SizedBox(height: 4),
                        Text(
                          t.name,
                          style: TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.w700,
                            color: isSelected ? t.color : AppColors.textSecondary,
                          ),
                        ),
                      ],
                    ),
                  ),
                );
              },
            ),
          ),
          const SizedBox(height: AppSpacing.lg),
          Text('Quyền lợi hạng ${tier.name}', style: Theme.of(context).textTheme.labelLarge),
          const SizedBox(height: AppSpacing.sm),
          ...tier.benefits.map(
            (b) => Padding(
              padding: const EdgeInsets.only(bottom: 6),
              child: Row(
                children: [
                  Icon(Icons.check, size: 16, color: tier.color),
                  const SizedBox(width: AppSpacing.sm),
                  Expanded(child: Text(b, style: Theme.of(context).textTheme.bodySmall)),
                ],
              ),
            ),
          ),
          const SizedBox(height: AppSpacing.lg),
          const Divider(),
          const SizedBox(height: AppSpacing.md),
          Text('Nhiệm vụ tuần này', style: Theme.of(context).textTheme.labelLarge),
          const SizedBox(height: AppSpacing.xs),
          Text(
            'Loại nhiệm vụ minh họa — chưa có dữ liệu tiến độ thật của bạn.',
            style: Theme.of(context).textTheme.bodySmall,
          ),
          const SizedBox(height: AppSpacing.sm),
          const _QuestRow(
            icon: Icons.local_taxi_outlined,
            title: 'Hoàn thành chuyến trong tuần',
            rewardLabel: 'Thưởng thêm',
          ),
          const SizedBox(height: AppSpacing.sm),
          const _QuestRow(
            icon: Icons.star_outline,
            title: 'Duy trì đánh giá cao',
            rewardLabel: 'Huy hiệu',
          ),
        ],
      ),
    );
  }
}

/// One illustrative quest row — shows the *shape* of the future quest
/// feature (icon, title, reward, a progress bar) without a fabricated
/// completion percentage. The bar is rendered in a fixed, low, neutral-gray
/// fill (`value: null` semantics, same convention as `ProgressRing`) — not
/// a specific number that would imply real tracked progress.
class _QuestRow extends StatelessWidget {
  const _QuestRow({required this.icon, required this.title, required this.rewardLabel});

  final IconData icon;
  final String title;
  final String rewardLabel;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(AppSpacing.md),
      decoration: BoxDecoration(
        color: AppColors.surfaceAlt,
        borderRadius: AppRadius.mdAll,
      ),
      child: Row(
        children: [
          Icon(icon, color: AppColors.textSecondary, size: 20),
          const SizedBox(width: AppSpacing.sm),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(title, style: Theme.of(context).textTheme.bodySmall?.copyWith(fontWeight: FontWeight.w600)),
                const SizedBox(height: 4),
                ClipRRect(
                  borderRadius: BorderRadius.circular(3),
                  child: const LinearProgressIndicator(
                    value: 0.08,
                    minHeight: 5,
                    backgroundColor: AppColors.divider,
                    valueColor: AlwaysStoppedAnimation(AppColors.textTertiary),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(width: AppSpacing.sm),
          Text(
            rewardLabel,
            style: const TextStyle(fontSize: 10, color: AppColors.textTertiary, fontWeight: FontWeight.w600),
          ),
        ],
      ),
    );
  }
}
