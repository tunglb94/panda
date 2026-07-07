import 'package:flutter/material.dart';

/// Replaces `TripActionButtons` the instant Accept is pressed: a single
/// full-width, disabled button with an inline spinner. There is no Reject
/// affordance here at all — satisfies "cannot reject anymore" by omission,
/// not by disabling a still-visible button.
class AcceptLoadingButton extends StatelessWidget {
  const AcceptLoadingButton({super.key});

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: double.infinity,
      child: FilledButton(
        onPressed: null,
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            SizedBox(
              width: 16,
              height: 16,
              child: CircularProgressIndicator(
                strokeWidth: 2,
                valueColor: AlwaysStoppedAnimation(Colors.white.withValues(alpha: 0.9)),
              ),
            ),
            const SizedBox(width: 10),
            const Text('Accepting…'),
          ],
        ),
      ),
    );
  }
}
