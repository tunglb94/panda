import 'package:flutter/material.dart';

enum _BookState { idle, loading, success }

/// Primary CTA for the Booking UI.
///
/// This performs no network call. Pressing it runs [onConfirm] (a caller-
/// supplied mock delay) and animates through loading → success states, which
/// is what the real booking submission
/// (`docs/project/MVP_DEVELOPMENT_PLAN.md` Rider App Roadmap stage R4) will
/// eventually drive for real.
class BookRideButton extends StatefulWidget {
  const BookRideButton({super.key, required this.label, required this.onConfirm});

  final String label;
  final Future<void> Function() onConfirm;

  @override
  State<BookRideButton> createState() => _BookRideButtonState();
}

class _BookRideButtonState extends State<BookRideButton> {
  _BookState _state = _BookState.idle;

  Future<void> _handlePress() async {
    if (_state != _BookState.idle) return;
    setState(() => _state = _BookState.loading);
    try {
      await widget.onConfirm();
      if (!mounted) return;
      setState(() => _state = _BookState.success);
      await Future.delayed(const Duration(milliseconds: 900));
      if (mounted) setState(() => _state = _BookState.idle);
    } catch (_) {
      // onConfirm signalled failure — reset so the user can retry.
      if (mounted) setState(() => _state = _BookState.idle);
    }
  }

  @override
  Widget build(BuildContext context) {
    return FilledButton(
      onPressed: _state == _BookState.idle ? _handlePress : null,
      child: AnimatedSwitcher(
        duration: const Duration(milliseconds: 220),
        transitionBuilder: (child, animation) =>
            ScaleTransition(scale: animation, child: child),
        child: switch (_state) {
          _BookState.idle => Text(widget.label, key: const ValueKey('idle')),
          _BookState.loading => const SizedBox(
              key: ValueKey('loading'),
              height: 20,
              width: 20,
              child: CircularProgressIndicator(
                strokeWidth: 2.4,
                color: Colors.white,
              ),
            ),
          _BookState.success => const Row(
              key: ValueKey('success'),
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(Icons.check_circle, color: Colors.white, size: 20),
                SizedBox(width: 8),
                Text('Requested'),
              ],
            ),
        },
      ),
    );
  }
}
