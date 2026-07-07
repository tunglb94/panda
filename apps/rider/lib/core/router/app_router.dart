import 'package:go_router/go_router.dart';
import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/routing/route_service.dart';
import 'package:rider/features/booking/presentation/pages/booking_page.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';
import 'package:rider/features/map/presentation/pages/map_page.dart';
import 'package:rider/features/profile/presentation/pages/profile_page.dart';
import 'package:rider/shared/widgets/scaffold_with_nav.dart';

abstract final class AppRoutes {
  static const home = '/';
  static const booking = '/booking';
  static const profile = '/profile';
}

class AppRouter {
  AppRouter._();

  static GoRouter create({
    required ApiClient apiClient,
    required RouteService routeService,
  }) {
    return GoRouter(
      initialLocation: AppRoutes.home,
      routes: [
        StatefulShellRoute.indexedStack(
          builder: (context, state, shell) => ScaffoldWithNav(shell: shell),
          branches: [
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: AppRoutes.home,
                  builder: (context, state) => MapPage(
                    apiClient: apiClient,
                    routeService: routeService,
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
                  ),
                ),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: AppRoutes.profile,
                  builder: (context, state) => const ProfilePage(),
                ),
              ],
            ),
          ],
        ),
      ],
    );
  }
}
