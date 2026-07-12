import 'package:flutter/material.dart';

import 'package:rider/shared/widgets/app_button.dart';
import 'package:rider/shared/widgets/app_dialog.dart';

/// Cancel Ride action, shown only before the trip has actually started
/// (searching / assigned / arriving — see `RiderTripStatus.isCancellable`).
///
/// Confirms with the rider first. [onCancel] is a plain callback — in the
/// real wiring (`trip_lifecycle_page.dart`) it calls the real
/// `TripRepository.cancelRide(tripId)`, i.e. a real `POST /rides/{id}/cancel`
/// request. Fixed during the Closed Beta polish pass: this widget's
/// confirmation dialog used to tell the rider "this is just a UI mock, no
/// request is sent" — inaccurate and actively misleading once the real
/// cancel flow was wired up, since a rider reading that text would not
/// realize confirming actually cancels their ride and notifies the driver.
class CancelRideButton extends StatelessWidget {
  const CancelRideButton({super.key, required this.onCancel});

  final VoidCallback onCancel;

  @override
  Widget build(BuildContext context) {
    return AppButton.danger(
      label: 'Hủy chuyến',
      icon: Icons.close,
      onPressed: () => _confirmCancel(context),
    );
  }

  Future<void> _confirmCancel(BuildContext context) async {
    final confirmed = await AppDialog.confirm(
      context,
      title: 'Hủy chuyến xe này?',
      message: 'Chuyến đi sẽ được hủy và tài xế sẽ được thông báo ngay lập tức.',
      confirmLabel: 'Hủy chuyến',
      cancelLabel: 'Giữ chuyến',
      isDestructive: true,
    );
    if (confirmed) onCancel();
  }
}
