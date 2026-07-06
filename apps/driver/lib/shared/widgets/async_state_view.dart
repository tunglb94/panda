import 'package:flutter/material.dart';

/// Generic Loading / Success / Empty / Error wrapper around a [Future].
///
/// Mirrors `apps/rider`'s `AsyncStateView` (Profile module, Phase R-03)
/// exactly — the two apps are separate Flutter projects with no shared
/// package (decided in Phase D-01), so this is hand-mirrored rather than
/// imported. Kept in `shared/widgets/` (not a single feature) since it is
/// intended to be reused across every Driver feature that fetches data,
/// starting with the Home dashboard (Phase D-02).
class AsyncStateView<T> extends StatelessWidget {
  const AsyncStateView({
    super.key,
    required this.future,
    required this.successBuilder,
    this.isEmpty,
    this.emptyBuilder,
    this.loadingBuilder,
    this.errorBuilder,
  });

  final Future<T> future;
  final Widget Function(BuildContext context, T data) successBuilder;

  /// Returns true when [data] should be treated as the Empty state instead
  /// of Success (e.g. a driver with no vehicle assigned yet). Omit for data
  /// types with no meaningful "empty" concept.
  final bool Function(T data)? isEmpty;

  final WidgetBuilder? emptyBuilder;
  final WidgetBuilder? loadingBuilder;
  final Widget Function(BuildContext context, Object error)? errorBuilder;

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<T>(
      future: future,
      builder: (context, snapshot) {
        final Widget child;
        final String stateKey;

        if (snapshot.connectionState != ConnectionState.done) {
          stateKey = 'loading';
          child = (loadingBuilder ?? _defaultLoading)(context);
        } else if (snapshot.hasError) {
          stateKey = 'error';
          child = (errorBuilder ?? _defaultError)(context, snapshot.error!);
        } else {
          final data = snapshot.data as T;
          if (isEmpty != null && isEmpty!(data)) {
            stateKey = 'empty';
            child = (emptyBuilder ?? _defaultEmpty)(context);
          } else {
            stateKey = 'success';
            child = successBuilder(context, data);
          }
        }

        return AnimatedSwitcher(
          duration: const Duration(milliseconds: 300),
          child: KeyedSubtree(key: ValueKey(stateKey), child: child),
        );
      },
    );
  }

  static Widget _defaultLoading(BuildContext context) => const Padding(
        padding: EdgeInsets.symmetric(vertical: 48),
        child: Center(child: CircularProgressIndicator()),
      );

  static Widget _defaultError(BuildContext context, Object error) => Padding(
        padding: const EdgeInsets.symmetric(vertical: 48),
        child: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.error_outline, size: 40, color: Colors.red.shade400),
              const SizedBox(height: 12),
              const Text('Something went wrong.'),
            ],
          ),
        ),
      );

  static Widget _defaultEmpty(BuildContext context) => const Padding(
        padding: EdgeInsets.symmetric(vertical: 48),
        child: Center(child: Text('Nothing here yet.')),
      );
}
