import 'package:go_router/go_router.dart';

import '../../features/auth/presentation/pages/login_page.dart';
import '../../features/kyc/data/kyc_repository.dart';
import '../../features/kyc/presentation/pages/driver_verifications_page.dart';
import '../auth/auth_state.dart';
import '../network/api_client.dart';
import '../storage/token_storage.dart';

abstract final class AppRoutes {
  static const login = '/login';
  static const driverVerifications = '/admin/driver-verifications';
}

abstract final class AppRouter {
  static GoRouter create({
    required AuthState authState,
    required TokenStorage tokenStorage,
    required ApiClient apiClient,
  }) {
    final kycRepository = KYCRepository(apiClient);
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
        GoRoute(
          path: AppRoutes.driverVerifications,
          builder: (context, state) => DriverVerificationsPage(
            repository: kycRepository,
            authState: authState,
            tokenStorage: tokenStorage,
          ),
        ),
      ],
    );
  }
}
