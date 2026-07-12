import 'package:flutter/material.dart';

/// Purely presentational press-down scale (1.0 → [scale]) driven by an
/// externally-supplied [pressed] flag — this widget never listens for
/// pointer/tap events itself, so it can safely wrap *any* interactive
/// child (a `FilledButton`, an `InkWell`) without competing with that
/// child's own gesture recognizer. Mirrors `apps/driver`'s `PressableScale`.
class PressableScale extends StatelessWidget {
  const PressableScale({
    super.key,
    required this.child,
    required this.pressed,
    this.scale = 0.96,
  });

  final Widget child;
  final bool pressed;
  final double scale;

  @override
  Widget build(BuildContext context) {
    return AnimatedScale(
      scale: pressed ? scale : 1.0,
      duration: const Duration(milliseconds: 120),
      curve: Curves.easeOut,
      child: child,
    );
  }
}

/// Wraps [child] with a [Listener] that tracks raw pointer down/up/cancel
/// state and feeds it into [PressableScale] — the ready-to-use version for
/// any widget that doesn't already expose a highlight/press callback of its
/// own (Material buttons, plain icons, etc). Safe to use around an
/// already-interactive child: `Listener` observes the pointer stream
/// without entering the tap gesture arena, so it never blocks or
/// double-fires the child's own `onPressed`/`onTap`.
class PressScaleObserver extends StatefulWidget {
  const PressScaleObserver({super.key, required this.child, this.scale = 0.96});

  final Widget child;
  final double scale;

  @override
  State<PressScaleObserver> createState() => _PressScaleObserverState();
}

class _PressScaleObserverState extends State<PressScaleObserver> {
  bool _pressed = false;

  void _setPressed(bool value) {
    if (_pressed != value) setState(() => _pressed = value);
  }

  @override
  Widget build(BuildContext context) {
    return Listener(
      onPointerDown: (_) => _setPressed(true),
      onPointerUp: (_) => _setPressed(false),
      onPointerCancel: (_) => _setPressed(false),
      child: PressableScale(pressed: _pressed, scale: widget.scale, child: widget.child),
    );
  }
}
