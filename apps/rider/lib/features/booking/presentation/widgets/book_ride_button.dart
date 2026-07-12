import 'package:flutter/material.dart';

import 'package:rider/shared/widgets/app_button.dart';

enum _BookState { idle, loading, success }

/// Primary CTA for the Booking UI.
///
/// This performs no network call. Pressing it runs [onConfirm] (a caller-
/// supplied mock delay) and animates through loading → success states, which
/// is what the real booking submission
/// (`docs/project/MVP_DEVELOPMENT_PLAN.md` Rider App Roadmap stage R4) will
/// eventually drive for real. Built on `AppButton`'s own loading/success
/// morph rather than reimplementing the same `AnimatedSwitcher` state
/// machine a second time — this widget now only owns the state transition
/// timing, not the visuals.
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
    return AppButton.primary(
      label: _state == _BookState.success ? 'Đã đặt xe' : widget.label,
      isLoading: _state == _BookState.loading,
      isSuccess: _state == _BookState.success,
      onPressed: _state == _BookState.idle ? _handlePress : null,
    );
  }
}
