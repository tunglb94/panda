import 'package:flutter/material.dart';

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
      appBar: AppBar(title: const Text('Developer')),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: Colors.grey.shade200),
                ),
                child: Column(
                  children: [
                    InfoRow(label: 'App version', value: info.appVersion),
                    const Divider(height: 1),
                    InfoRow(label: 'Build mode', value: info.buildMode.label),
                    const Divider(height: 1),
                    InfoRow(
                      label: 'Flutter version',
                      value: info.flutterVersionPlaceholder,
                    ),
                    const Divider(height: 1),
                    InfoRow(
                      label: 'Environment',
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
