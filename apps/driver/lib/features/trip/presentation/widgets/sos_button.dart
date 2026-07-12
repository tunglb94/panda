import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../shared/widgets/pressable_scale.dart';
import '../../../safety/presentation/pages/safety_center_page.dart';

/// Always-reachable emergency action, shown as an AppBar icon (not buried
/// inside the trip card) during Picking Up / Waiting / In Trip — the states
/// where a driver might actually need it — so it's a single tap away
/// without scrolling, matching the "thao tác một tay" requirement.
///
/// Opens the full Safety Center (emergency numbers, trusted contacts,
/// incident report, safety checklist) rather than a standalone mini-sheet —
/// consolidated so there is exactly one safety UI in the app, not a
/// duplicate/competing one live only during a trip.
class SosButton extends StatelessWidget {
  const SosButton({super.key});

  @override
  Widget build(BuildContext context) {
    return PressScaleObserver(
      child: IconButton(
        onPressed: () => Navigator.of(context).push(
          MaterialPageRoute(builder: (_) => const SafetyCenterPage()),
        ),
        icon: const Icon(Icons.emergency_outlined, color: AppColors.error),
        tooltip: 'Khẩn cấp',
      ),
    );
  }
}
