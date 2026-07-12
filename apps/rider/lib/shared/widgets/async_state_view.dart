import 'package:flutter/material.dart';

import 'app_empty_state.dart';
import 'app_loading_view.dart';

/// Generic Loading / Success / Empty / Error wrapper around a [Future].
///
/// Promoted from `features/profile/presentation/widgets/async_state_view.dart`
/// (Phase R-03) to `shared/widgets/` as part of the design-system sync, so
/// every feature that fetches data reuses one implementation instead of
/// each screen (e.g. `trip_history_page.dart`) hand-rolling its own
/// Loading/Error/Empty `FutureBuilder` branching. Mirrors `apps/driver`'s
/// `AsyncStateView`.
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
  /// of Success (e.g. an empty list). Omit for data types with no
  /// meaningful "empty" concept (e.g. a single profile object).
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
        child: AppLoadingView(),
      );

  static Widget _defaultError(BuildContext context, Object error) =>
      const Padding(
        padding: EdgeInsets.symmetric(vertical: 24),
        child: AppEmptyState.error(),
      );

  static Widget _defaultEmpty(BuildContext context) => const Padding(
        padding: EdgeInsets.symmetric(vertical: 24),
        child: AppEmptyState(
          icon: Icons.inbox_outlined,
          title: 'Chưa có gì ở đây',
        ),
      );
}
