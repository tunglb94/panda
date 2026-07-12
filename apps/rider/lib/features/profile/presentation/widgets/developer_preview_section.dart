import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/features/trip/presentation/pages/trip_preview_menu_page.dart';

import 'settings_section.dart';
import '../../domain/models/settings_entry.dart';

/// Carries over the "Trip UI Preview (dev)" entry point added in Phase R-02
/// (it previously lived directly on `ProfilePage`, which R-03 rewrote).
/// Kept separate from the required settings groups so it's clearly marked
/// as a development aid, not a real user-facing setting.
class DeveloperPreviewSection extends StatelessWidget {
  const DeveloperPreviewSection({super.key, required this.apiClient});

  final ApiClient apiClient;

  @override
  Widget build(BuildContext context) {
    return SettingsSection(
      title: 'Nhà phát triển',
      entries: const [
        SettingsEntry(
          action: SettingsAction.developerPreview,
          icon: Icons.bug_report_outlined,
          label: 'Xem trước giao diện chuyến đi (dev)',
        ),
      ],
      onTapEntry: (_) => Navigator.of(context).push(
        MaterialPageRoute(builder: (_) => TripPreviewMenuPage(apiClient: apiClient)),
      ),
    );
  }
}
