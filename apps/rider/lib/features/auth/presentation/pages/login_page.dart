import 'package:flutter/material.dart';
import '../../../../core/auth/auth_state.dart';
import '../../../../core/network/api_client.dart';
import '../../../../core/storage/token_storage.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_radius.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_button.dart';
import '../../../../shared/widgets/mascot_image.dart';
import '../../data/auth_repository.dart';

class LoginPage extends StatefulWidget {
  const LoginPage({
    super.key,
    required this.authState,
    required this.tokenStorage,
    required this.apiClient,
  });

  final AuthState authState;
  final TokenStorage tokenStorage;
  final ApiClient apiClient;

  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  final _phoneCtrl = TextEditingController();
  bool _isLoading = false;
  String? _error;

  @override
  void dispose() {
    _phoneCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.symmetric(horizontal: 32),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                _Logo(cs: cs),
                const SizedBox(height: AppSpacing.xxxl + AppSpacing.lg),
                TextField(
                  controller: _phoneCtrl,
                  keyboardType: TextInputType.phone,
                  textInputAction: TextInputAction.done,
                  enabled: !_isLoading,
                  onSubmitted: (_) => _login(),
                  decoration: const InputDecoration(
                    labelText: 'Số điện thoại',
                    hintText: '090 123 4567',
                    border: OutlineInputBorder(borderRadius: AppRadius.mdAll),
                    prefixIcon: Icon(Icons.phone_outlined),
                  ),
                ),
                const SizedBox(height: AppSpacing.md),
                if (_error != null)
                  Text(
                    _error!,
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppColors.error),
                    textAlign: TextAlign.center,
                  ),
                const SizedBox(height: AppSpacing.xl),
                AppButton.primary(
                  label: 'Đăng nhập',
                  isLoading: _isLoading,
                  onPressed: _isLoading ? null : _login,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Future<void> _login() async {
    if (_isLoading) return;
    final rawPhone = _phoneCtrl.text.trim();
    if (rawPhone.isEmpty) {
      setState(() => _error = 'Vui lòng nhập số điện thoại');
      return;
    }
    final phone = _normalizeVietnamesePhone(rawPhone);

    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final repo = AuthRepository(widget.apiClient);
      final result = await repo.loginRider(phone);
      await widget.authState.login(
        accessToken: result.accessToken,
        riderId: result.riderId,
        storage: widget.tokenStorage,
      );
      // GoRouter's refreshListenable redirects to home automatically.
    } on ApiException catch (e) {
      if (mounted) {
        setState(() {
          _error = e.statusCode == 0 ? e.message : 'Đăng nhập thất bại. Vui lòng thử lại.';
          _isLoading = false;
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() {
          _error = 'Đăng nhập thất bại. Vui lòng thử lại.';
          _isLoading = false;
        });
      }
    }
  }
}

/// Accepts the Vietnamese-style local input riders actually type (e.g.
/// `090 123 4567`) and normalizes it to the `+84...` form the backend
/// stores phone numbers in. Panda only serves Vietnam today, so the app
/// doesn't ask riders to type a country code. Input that already has a `+`
/// is passed through unchanged (covers dev/test accounts).
String _normalizeVietnamesePhone(String raw) {
  final digitsOnly = raw.replaceAll(RegExp(r'[^0-9+]'), '');
  if (digitsOnly.startsWith('+')) return digitsOnly;
  if (digitsOnly.startsWith('84')) return '+$digitsOnly';
  if (digitsOnly.startsWith('0')) return '+84${digitsOnly.substring(1)}';
  return '+84$digitsOnly';
}

class _Logo extends StatelessWidget {
  const _Logo({required this.cs});

  final ColorScheme cs;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        const MascotImage(
          asset: 'mascot_welcome.png',
          size: MascotSize.large,
          animation: MascotAnimation.bounce,
          semanticLabel: 'Chào mừng đến với Panda',
        ),
        const SizedBox(height: 16),
        Text(
          'Panda',
          style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                fontWeight: FontWeight.bold,
                color: cs.primary,
              ),
        ),
        const SizedBox(height: 4),
        Text(
          'Ứng dụng Hành khách',
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                color: cs.onSurfaceVariant,
              ),
        ),
      ],
    );
  }
}
