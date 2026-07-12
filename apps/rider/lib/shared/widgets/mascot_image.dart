import 'package:flutter/material.dart';

/// Size tiers per the Mascot Design Rule (see `docs/design/MASCOT_CATALOG.md`):
/// [large] 120–180dp for hero moments (Empty State, Celebration), [medium]
/// 72–96dp for supporting moments (inline success/reward), [small] 32–48dp
/// for compact inline accents. Values below are the mid-point of each band.
enum MascotSize { large, medium, small }

enum MascotAnimation { fade, scale, slide, bounce, none }

/// The one mascot widget for Panda. Renders a PNG from `assets/mascot/`
/// with a single, deliberately restrained entrance animation — mascots are
/// reserved for emotional moments (Empty State, Success, Celebration,
/// Achievement, Welcome, Reward — see the catalog), never for navigation,
/// map, trip controls, fare, or form elements, so this widget is used
/// sparingly by design, not dropped in everywhere.
///
/// Uses only implicit Flutter animations (`AnimatedOpacity`/`AnimatedScale`/
/// `AnimatedSlide`), no `AnimationController` and no new package — mirrors
/// the same one-shot entrance pattern already used by `AppCard`. The
/// animation plays once per mount; a mascot that merely rebuilds does not
/// replay it.
///
/// Callers are responsible for the ≥24dp surrounding whitespace required by
/// the design rule (e.g. wrap in `SizedBox`/`Padding` using `AppSpacing.xxl`)
/// — this widget only sizes the image itself, not its margins, since the
/// right margin depends on the surrounding layout.
class MascotImage extends StatefulWidget {
  const MascotImage({
    super.key,
    required this.asset,
    this.size = MascotSize.medium,
    this.animation = MascotAnimation.fade,
    this.semanticLabel,
  });

  /// File name under `assets/mascot/`, e.g. `mascot_celebration.png`.
  final String asset;
  final MascotSize size;
  final MascotAnimation animation;
  final String? semanticLabel;

  double get _dimension => switch (size) {
        MascotSize.large => 150,
        MascotSize.medium => 88,
        MascotSize.small => 40,
      };

  @override
  State<MascotImage> createState() => _MascotImageState();
}

class _MascotImageState extends State<MascotImage> {
  bool _visible = false;

  @override
  void initState() {
    super.initState();
    if (widget.animation == MascotAnimation.none) {
      _visible = true;
      return;
    }
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) setState(() => _visible = true);
    });
  }

  @override
  Widget build(BuildContext context) {
    final image = Image.asset(
      'assets/mascot/${widget.asset}',
      width: widget._dimension,
      height: widget._dimension,
      fit: BoxFit.contain,
      semanticLabel: widget.semanticLabel,
    );

    return switch (widget.animation) {
      MascotAnimation.none => image,
      MascotAnimation.fade => AnimatedOpacity(
          opacity: _visible ? 1 : 0,
          duration: const Duration(milliseconds: 320),
          curve: Curves.easeOut,
          child: image,
        ),
      MascotAnimation.scale => AnimatedScale(
          scale: _visible ? 1 : 0.7,
          duration: const Duration(milliseconds: 320),
          curve: Curves.easeOut,
          child: AnimatedOpacity(
            opacity: _visible ? 1 : 0,
            duration: const Duration(milliseconds: 320),
            child: image,
          ),
        ),
      MascotAnimation.slide => AnimatedSlide(
          offset: _visible ? Offset.zero : const Offset(0, 0.08),
          duration: const Duration(milliseconds: 320),
          curve: Curves.easeOut,
          child: AnimatedOpacity(
            opacity: _visible ? 1 : 0,
            duration: const Duration(milliseconds: 320),
            child: image,
          ),
        ),
      // A gentle overshoot via Curves.easeOutBack reads as a light "bounce"
      // without a physics-simulation package or AnimationController.
      MascotAnimation.bounce => AnimatedScale(
          scale: _visible ? 1 : 0.5,
          duration: const Duration(milliseconds: 480),
          curve: Curves.easeOutBack,
          child: AnimatedOpacity(
            opacity: _visible ? 1 : 0,
            duration: const Duration(milliseconds: 240),
            child: image,
          ),
        ),
    };
  }
}
