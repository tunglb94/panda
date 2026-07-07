import 'package:flutter/material.dart';

/// Self-ticking mm:ss waiting-time counter for the Arrived-at-Pickup screen.
/// No package — a recursive `Future.delayed(1s)` chain, same self-scheduling
/// shape as `_RouteProgressTicker` (Phase D-05), stopped via a `_stopped`
/// flag in [dispose] so it never fires `setState` after the widget is gone.
/// [initialSeconds] lets the live flow start at 0 and the Arrival Preview
/// seed a fixed starting point (00:00 / 03:00 / 08:00) without waiting in
/// real time.
class WaitingTimer extends StatefulWidget {
  const WaitingTimer({super.key, this.initialSeconds = 0, this.onMinutePassed});

  final int initialSeconds;

  /// Fires once per whole minute reached (mock only — no real billing).
  final ValueChanged<int>? onMinutePassed;

  @override
  State<WaitingTimer> createState() => _WaitingTimerState();
}

class _WaitingTimerState extends State<WaitingTimer> {
  late int _seconds;
  bool _stopped = false;

  @override
  void initState() {
    super.initState();
    _seconds = widget.initialSeconds;
    _scheduleTick();
  }

  Future<void> _scheduleTick() async {
    await Future.delayed(const Duration(seconds: 1));
    if (_stopped || !mounted) return;
    setState(() => _seconds++);
    if (_seconds % 60 == 0) {
      widget.onMinutePassed?.call(_seconds ~/ 60);
    }
    _scheduleTick();
  }

  @override
  void dispose() {
    _stopped = true;
    super.dispose();
  }

  String get _formatted {
    final minutes = (_seconds ~/ 60).toString().padLeft(2, '0');
    final seconds = (_seconds % 60).toString().padLeft(2, '0');
    return '$minutes:$seconds';
  }

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(Icons.hourglass_bottom, size: 16, color: Colors.grey.shade600),
        const SizedBox(width: 6),
        Text(
          'Waiting time',
          style: TextStyle(fontSize: 12, color: Colors.grey.shade600),
        ),
        const SizedBox(width: 8),
        Text(_formatted, style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 14)),
      ],
    );
  }
}
