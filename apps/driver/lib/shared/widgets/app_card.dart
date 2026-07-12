import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';
import '../../core/theme/app_radius.dart';
import '../../core/theme/app_shadows.dart';
import '../../core/theme/app_spacing.dart';
import 'pressable_scale.dart';

/// The one card surface for PandaDriver: white background, [AppRadius.lg]
/// corners, a soft [AppShadows.card] shadow instead of a hard border.
///
/// Plays a one-shot fade + slide-up + shadow-in entrance animation the
/// first time it mounts (purely implicit — `AnimatedOpacity`/
/// `AnimatedSlide`/`AnimatedContainer`, no `AnimationController`), so a card
/// appearing on screen reads as a deliberate arrival rather than a jump-cut.
/// Because the animation lives in `State.initState`, it plays exactly once
/// per `State` object — a card that merely rebuilds (e.g. its child content
/// changes while it stays in the same list position) does not replay it;
/// only a genuinely new card instance does. Set [animateIn] to false to
/// opt out (e.g. if the card is already inside an outer transition that
/// would otherwise double up).
class AppCard extends StatefulWidget {
  const AppCard({
    super.key,
    required this.child,
    this.padding = const EdgeInsets.all(AppSpacing.lg),
    this.color = AppColors.surface,
    this.onTap,
    this.animateIn = true,
  });

  final Widget child;
  final EdgeInsetsGeometry padding;
  final Color color;
  final VoidCallback? onTap;
  final bool animateIn;

  @override
  State<AppCard> createState() => _AppCardState();
}

class _AppCardState extends State<AppCard> {
  bool _visible = false;
  bool _pressed = false;

  @override
  void initState() {
    super.initState();
    if (!widget.animateIn) {
      _visible = true;
      return;
    }
    // One frame delay so the implicit animations actually animate from
    // their initial (hidden) value instead of snapping straight to visible.
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) setState(() => _visible = true);
    });
  }

  @override
  Widget build(BuildContext context) {
    final surface = AnimatedContainer(
      duration: const Duration(milliseconds: 260),
      curve: Curves.easeOut,
      padding: widget.padding,
      decoration: BoxDecoration(
        color: widget.color,
        borderRadius: AppRadius.lgAll,
        boxShadow: _visible ? AppShadows.card : AppShadows.none,
      ),
      child: widget.child,
    );

    final content = widget.onTap == null
        ? surface
        : Material(
            color: Colors.transparent,
            borderRadius: AppRadius.lgAll,
            child: InkWell(
              onTap: widget.onTap,
              borderRadius: AppRadius.lgAll,
              onHighlightChanged: (v) => setState(() => _pressed = v),
              child: PressableScale(pressed: _pressed, scale: 0.98, child: surface),
            ),
          );

    return AnimatedSlide(
      offset: _visible ? Offset.zero : const Offset(0, 0.03),
      duration: const Duration(milliseconds: 260),
      curve: Curves.easeOut,
      child: AnimatedOpacity(
        opacity: _visible ? 1 : 0,
        duration: const Duration(milliseconds: 260),
        curve: Curves.easeOut,
        child: content,
      ),
    );
  }
}
