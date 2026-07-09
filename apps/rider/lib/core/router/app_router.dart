import 'package:go_router/go_router.dart';
import 'package:rider/core/auth/auth_state.dart';
import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/routing/route_provider.dart';
import 'package:rider/core/storage/token_storage.dart';
import 'package:rider/features/auth/presentation/pages/login_page.dart';
import 'package:rider/features/booking/presentation/pages/booking_page.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';
import 'package:rider/features/map/presentation/pages/map_page.dart';
import 'package:rider/features/profile/presentation/pages/profile_page.dart';
import 'package:rider/shared/widgets/scaffold_with_nav.dart';

abstract final class AppRoutes {
  static const login = '/login';
  static const home = '/';
  static const booking = '/booking';
  static const profile = '/profile';
}

class AppRouter {
  AppRouter._();

  static GoRouter create({
    required AuthState authState,
    required TokenStorage tokenStorage,
    required ApiClient apiClient,
    required RouteProvider routeProvider,
  }) {
    return GoRouter(
      initialLocation: AppRoutes.home,
      refreshListenable: authState,
      redirect: (context, state) {
        final isLoggedIn = authState.isLoggedIn;
        final isOnLogin = state.matchedLocation == AppRoutes.login;
        if (!isLoggedIn && !isOnLogin) return AppRoutes.login;
        if (isLoggedIn && isOnLogin) return AppRoutes.home;
        return null;
      },
      routes: [
        GoRoute(
          path: AppRoutes.login,
          builder: (context, state) => LoginPage(
            authState: authState,
            tokenStorage: tokenStorage,
            apiClient: apiClient,
          ),
        ),
        StatefulShellRoute.indexedStack(
          builder: (context, state, shell) => ScaffoldWithNav(shell: shell),
          branches: [
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: AppRoutes.home,
                  builder: (context, state) => MapPage(
                    apiClient: apiClient,
                    routeProvider: routeProvider,
                  ),
                ),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: AppRoutes.booking,
                  builder: (context, state) => BookingPage(
                    tripSelection: state.extra is TripSelection
                        ? state.extra as TripSelection
                        : null,
                    apiClient: apiClient,
                  ),
                ),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: AppRoutes.profile,
                  builder: (context, state) => ProfilePage(
                    authState: authState,
                    tokenStorage: tokenStorage,
                    apiClient: apiClient,
                  ),
                ),
              ],
            ),
          ],
        ),
      ],
    );
  }
}
