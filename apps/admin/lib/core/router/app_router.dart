import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../../features/auth/presentation/pages/login_page.dart';
import '../../features/kyc/data/kyc_repository.dart';
import '../../features/kyc/presentation/pages/driver_verifications_page.dart';
import '../../features/payout/data/payout_repository.dart';
import '../../features/payout/presentation/pages/payouts_page.dart';
import '../../features/promotion/data/promotion_repository.dart';
import '../../features/promotion/presentation/pages/vouchers_page.dart';
import '../auth/auth_state.dart';
import '../layout/admin_shell.dart';
import '../network/api_client.dart';
import '../storage/token_storage.dart';

abstract final class AppRoutes {
  static const login = '/login';
  static const driverVerifications = '/admin/driver-verifications';
  static const vouchers = '/admin/vouchers';
  static const payouts = '/admin/payouts';
}

const _navItems = [
  AdminNavItem(path: AppRoutes.driverVerifications, label: 'KYC', icon: Icons.badge_outlined),
  AdminNavItem(path: AppRoutes.vouchers, label: 'Voucher', icon: Icons.local_offer_outlined),
  AdminNavItem(path: AppRoutes.payouts, label: 'Payout', icon: Icons.account_balance_wallet_outlined),
];

abstract final class AppRouter {
  static GoRouter create({
    required AuthState authState,
    required TokenStorage tokenStorage,
    required ApiClient apiClient,
  }) {
    final kycRepository = KYCRepository(apiClient);
    final promotionRepository = PromotionRepository(apiClient);
    final payoutRepository = PayoutRepository(apiClient);
    return GoRouter(
      initialLocation: AppRoutes.driverVerifications,
      refreshListenable: authState,
      redirect: (context, state) {
        final isLoggedIn = authState.isLoggedIn;
        final isOnLogin = state.matchedLocation == AppRoutes.login;
        if (!isLoggedIn && !isOnLogin) return AppRoutes.login;
        if (isLoggedIn && isOnLogin) return AppRoutes.driverVerifications;
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
        ShellRoute(
          builder: (context, state, child) => AdminShell(
            currentPath: state.matchedLocation,
            items: _navItems,
            onNavigate: (path) => context.go(path),
            authState: authState,
            tokenStorage: tokenStorage,
            child: child,
          ),
          routes: [
            GoRoute(
              path: AppRoutes.driverVerifications,
              builder: (context, state) => DriverVerificationsPage(
                repository: kycRepository,
                authState: authState,
                tokenStorage: tokenStorage,
              ),
            ),
            GoRoute(
              path: AppRoutes.vouchers,
              builder: (context, state) => VouchersPage(
                repository: promotionRepository,
                authState: authState,
                tokenStorage: tokenStorage,
              ),
            ),
            GoRoute(
              path: AppRoutes.payouts,
              builder: (context, state) => PayoutsPage(
                repository: payoutRepository,
                authState: authState,
                tokenStorage: tokenStorage,
              ),
            ),
          ],
        ),
      ],
    );
  }
}
