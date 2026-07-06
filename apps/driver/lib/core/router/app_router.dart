import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../../features/earnings/presentation/pages/earnings_page.dart';
import '../../features/home/presentation/pages/home_page.dart';
import '../../features/profile/presentation/pages/profile_page.dart';
import '../../features/trip/presentation/pages/trip_page.dart';
import '../../shared/widgets/scaffold_with_nav.dart';

abstract final class AppRoutes {
  static const home = '/';
  static const trip = '/trip';
  static const earnings = '/earnings';
  static const profile = '/profile';
}

abstract final class AppRouter {
  static final router = GoRouter(
    initialLocation: AppRoutes.home,
    routes: [
      StatefulShellRoute.indexedStack(
        builder: (context, state, shell) => ScaffoldWithNav(shell: shell),
        branches: [
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: AppRoutes.home,
                builder: (context, state) => const HomePage(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: AppRoutes.trip,
                builder: (context, state) => const TripPage(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: AppRoutes.earnings,
                builder: (context, state) => const EarningsPage(),
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
