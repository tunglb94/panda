import 'package:go_router/go_router.dart';

import 'package:driver/features/earnings/presentation/pages/earnings_page.dart';
import 'package:driver/features/home/presentation/pages/home_page.dart';
import 'package:driver/features/notifications/presentation/pages/notifications_page.dart';
import 'package:driver/features/profile/presentation/pages/profile_page.dart';
import 'package:driver/features/trips/presentation/pages/trips_page.dart';
import 'package:driver/shared/widgets/scaffold_with_nav.dart';

abstract final class AppRoutes {
  static const home = '/';
  static const trips = '/trips';
  static const earnings = '/earnings';
  static const notifications = '/notifications';
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
                name: 'home',
                builder: (context, state) => const HomePage(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: AppRoutes.trips,
                name: 'trips',
                builder: (context, state) => const TripsPage(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: AppRoutes.earnings,
                name: 'earnings',
                builder: (context, state) => const EarningsPage(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: AppRoutes.notifications,
                name: 'notifications',
                builder: (context, state) => const NotificationsPage(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: AppRoutes.profile,
                name: 'profile',
                builder: (context, state) => const ProfilePage(),
              ),
            ],
          ),
        ],
      ),
    ],
  );
}
