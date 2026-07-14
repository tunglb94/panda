import 'package:flutter/foundation.dart';
import 'package:go_router/go_router.dart';

import '../../features/auth/presentation/pages/login_page.dart';
import '../../features/earnings/presentation/pages/earnings_page.dart';
import '../../features/kyc/presentation/pages/kyc_status_page.dart';
import '../../features/location/services/location_upload_service.dart';
import '../../features/map/presentation/pages/map_page.dart';
import '../../features/notifications/presentation/pages/notifications_page.dart';
import '../../features/profile/presentation/pages/profile_page.dart';
import '../../features/trip/presentation/pages/trip_page.dart';
import '../../shared/widgets/scaffold_with_nav.dart';
import '../auth/auth_state.dart';
import '../kyc/kyc_gate.dart';
import '../network/api_client.dart';
import '../splash/splash_page.dart';
import '../storage/token_storage.dart';

abstract final class AppRoutes {
  static const splash = '/splash';
  static const login = '/login';
  static const kycStatus = '/kyc-status';
  static const home = '/';
  static const trips = '/trips';
  static const earnings = '/earnings';
  static const notifications = '/notifications';
  static const profile = '/profile';
}

abstract final class AppRouter {
  static GoRouter create({
    required AuthState authState,
    required KycGate kycGate,
    required TokenStorage tokenStorage,
    required ApiClient apiClient,
    required LocationUploadService uploadService,
  }) {
    return GoRouter(
      initialLocation: AppRoutes.splash,
      refreshListenable: Listenable.merge([authState, kycGate]),
      // APP STARTUP state machine: Splash → Check Token (AuthState.initialize,
      // already resolved before runApp) → Refresh Token (transparent, inside
      // ApiClient) → Load Profile / Check KYC (KycGate.refresh, triggered by
      // SplashPage) → route. kycGate.approved == null means "still resolving".
      redirect: (context, state) {
        final loc = state.matchedLocation;
        if (!authState.isLoggedIn) {
          return loc == AppRoutes.login ? null : AppRoutes.login;
        }
        if (loc == AppRoutes.login || loc == AppRoutes.splash) {
          final approved = kycGate.approved;
          if (approved == null) return AppRoutes.splash;
          return approved ? AppRoutes.home : AppRoutes.kycStatus;
        }
        if (kycGate.approved == false && loc != AppRoutes.kycStatus) {
          return AppRoutes.kycStatus;
        }
        return null;
      },
      routes: [
        GoRoute(
          path: AppRoutes.splash,
          builder: (context, state) => SplashPage(
            authState: authState,
            kycGate: kycGate,
            apiClient: apiClient,
          ),
        ),
        GoRoute(
          path: AppRoutes.login,
          builder: (context, state) => LoginPage(
            authState: authState,
            kycGate: kycGate,
            tokenStorage: tokenStorage,
            apiClient: apiClient,
          ),
        ),
        GoRoute(
          path: AppRoutes.kycStatus,
          builder: (context, state) => KYCStatusPage(apiClient: apiClient),
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
