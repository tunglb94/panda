import 'package:flutter/foundation.dart';
import 'package:go_router/go_router.dart';
import 'package:rider/core/auth/auth_state.dart';
import 'package:rider/core/kyc/kyc_gate.dart';
import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/routing/route_provider.dart';
import 'package:rider/core/splash/splash_page.dart';
import 'package:rider/core/storage/token_storage.dart';
import 'package:rider/features/auth/presentation/pages/login_page.dart';
import 'package:rider/features/home/presentation/pages/home_hub_page.dart';
import 'package:rider/features/kyc/presentation/pages/rider_kyc_page.dart';
import 'package:rider/features/map/presentation/pages/map_page.dart';
import 'package:rider/features/profile/presentation/pages/profile_page.dart';
import 'package:rider/features/wallet/presentation/pages/wallet_page.dart';
import 'package:rider/shared/widgets/scaffold_with_nav.dart';

abstract final class AppRoutes {
  static const splash = '/splash';
  static const login = '/login';
  static const kyc = '/kyc';
  static const home = '/';
  static const booking = '/booking';
  static const wallet = '/wallet';
  static const profile = '/profile';
}

class AppRouter {
  AppRouter._();

  static GoRouter create({
    required AuthState authState,
    required KycGate kycGate,
    required TokenStorage tokenStorage,
    required ApiClient apiClient,
    required RouteProvider routeProvider,
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
          return approved ? AppRoutes.home : AppRoutes.kyc;
        }
        if (kycGate.approved == false && loc != AppRoutes.kyc) {
          return AppRoutes.kyc;
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
          path: AppRoutes.kyc,
          builder: (context, state) => RiderKycPage(apiClient: apiClient),
        ),
        StatefulShellRoute.indexedStack(
          builder: (context, state, shell) => ScaffoldWithNav(shell: shell),
          branches: [
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: AppRoutes.home,
                  builder: (context, state) => HomeHubPage(
                    apiClient: apiClient,
                    routeProvider: routeProvider,
                  ),
                ),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(
                  // The full real pickup/destination selection flow (current
                  // location default, search, drag-to-adjust, confirm) — the
                  // same MapPage Home's "Đặt xe" card pushes, so there is
                  // only one real booking entry point instead of this tab
                  // showing a non-editable sample trip.
                  path: AppRoutes.booking,
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
                  path: AppRoutes.wallet,
                  builder: (context, state) => const WalletPage(),
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
