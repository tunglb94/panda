import 'package:go_router/go_router.dart';
import 'package:rider/features/booking/presentation/pages/booking_page.dart';
import 'package:rider/features/map/presentation/pages/map_page.dart';
import 'package:rider/features/profile/presentation/pages/profile_page.dart';
import 'package:rider/shared/widgets/scaffold_with_nav.dart';

abstract final class AppRoutes {
  static const home = '/';
  static const booking = '/booking';
  static const profile = '/profile';
}

class AppRouter {
  static final GoRouter router = GoRouter(
    initialLocation: AppRoutes.home,
    routes: [
      StatefulShellRoute.indexedStack(
        builder: (context, state, shell) => ScaffoldWithNav(shell: shell),
        branches: [
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: AppRoutes.home,
                builder: (context, state) => const MapPage(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: AppRoutes.booking,
                builder: (context, state) => const BookingPage(),
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
