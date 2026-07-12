import 'package:flutter/material.dart';

import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../data/mock_app_info_repository.dart';
import '../../domain/models/app_info.dart';
import '../widgets/info_row.dart';

/// Development Utilities screen — accessible only from the Profile tab.
/// Shows app version (mock), build mode (real), and Flutter
/// version/environment (placeholders). No HTTP, no backend.
class DeveloperPage extends StatelessWidget {
  const DeveloperPage({super.key});

  static const _repository = MockAppInfoRepository();

  @override
  Widget build(BuildContext context) {
    final info = _repository.current();
    return Scaffold(
      appBar: AppBar(title: const Text('Nhà phát triển')),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: Padding(
              padding: const EdgeInsets.all(AppSpacing.lg),
              child: AppCard(
                padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg),
                child: Column(
                  children: [
                    InfoRow(label: 'Phiên bản ứng dụng', value: info.appVersion),
                    const Divider(height: 1),
                    InfoRow(label: 'Chế độ build', value: info.buildMode.label),
                    const Divider(height: 1),
                    InfoRow(
                      label: 'Phiên bản Flutter',
                      value: info.flutterVersionPlaceholder,
                    ),
                    const Divider(height: 1),
                    InfoRow(
                      label: 'Môi trường',
                      value: info.environmentPlaceholder,
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
