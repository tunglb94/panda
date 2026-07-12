import 'package:go_router/go_router.dart';

import '../../features/auth/presentation/pages/login_page.dart';
import '../../features/earnings/presentation/pages/earnings_page.dart';
import '../../features/location/services/location_upload_service.dart';
import '../../features/map/presentation/pages/map_page.dart';
import '../../features/notifications/presentation/pages/notifications_page.dart';
import '../../features/profile/presentation/pages/profile_page.dart';
import '../../features/trip/presentation/pages/trip_page.dart';
import '../../shared/widgets/scaffold_with_nav.dart';
import '../auth/auth_state.dart';
import '../network/api_client.dart';
import '../storage/token_storage.dart';

abstract final class AppRoutes {
  static const login = '/login';
  static const home = '/';
  static const trips = '/trips';
  static const earnings = '/earnings';
  static const notifications = '/notifications';
  static const profile = '/profile';
}

abstract final class AppRouter {
  static GoRouter create({
    required AuthState authState,
    required TokenStorage tokenStorage,
    required ApiClient apiClient,
    required LocationUploadService uploadService,
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
                    authState: authState,
                    tokenStorage: tokenStorage,
                    apiClient: apiClient,
                    uploadService: uploadService,
                  ),
                ),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: AppRoutes.trips,
                  builder: (context, state) => TripPage(
                    apiClient: apiClient,
                    locationStream: uploadService.locationStream,
                  ),
                ),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: AppRoutes.earnings,
                  builder: (context, state) => EarningsPage(apiClient: apiClient),
                ),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: AppRoutes.notifications,
                  builder: (context, state) => NotificationsPage(apiClient: apiClient),
                ),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: AppRoutes.profile,
                  builder: (context, state) => ProfilePage(
                    apiClient: apiClient,
                    authState: authState,
                    tokenStorage: tokenStorage,
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
