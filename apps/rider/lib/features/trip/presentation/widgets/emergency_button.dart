import 'package:flutter/material.dart';

import 'package:rider/shared/widgets/app_button.dart';
import 'package:rider/shared/widgets/app_snackbar.dart';

/// The emergency/SOS action is not wired up yet — tapping it only shows a
/// message explaining that, so the button's presence never misleads the
/// rider into thinking help was actually contacted.
class EmergencyButton extends StatelessWidget {
  const EmergencyButton({super.key});

  @override
  Widget build(BuildContext context) {
    return AppButton.danger(
      label: 'Khẩn cấp',
      icon: Icons.emergency_outlined,
      onPressed: () => AppSnackbar.warning(
        context,
        'Tính năng hỗ trợ khẩn cấp sẽ sớm ra mắt.',
      ),
    );
  }
}
