import 'dart:async';

import 'package:flutter/material.dart';
import 'package:google_sign_in/google_sign_in.dart';

import '../../../../core/auth/auth_state.dart';
import '../../../../core/config/app_config.dart';
import '../../../../core/kyc/kyc_gate.dart';
import '../../../../core/network/api_client.dart';
import '../../../../core/storage/token_storage.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_radius.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_button.dart';
import '../../../../shared/widgets/mascot_image.dart';
import '../../data/auth_repository.dart';

/// Login flow per the Auth + Onboarding spec: phone + OTP (self-registering
/// — a correct code on an unknown phone auto-creates the account, no office
/// visit) or Google Sign-In, both handled the same way once tokens come
/// back — GoRouter's redirect takes it from there via [AuthState].
class LoginPage extends StatefulWidget {
  const LoginPage({
    super.key,
    required this.authState,
    required this.kycGate,
    required this.tokenStorage,
    required this.apiClient,
  });

  final AuthState authState;
  final KycGate kycGate;
  final TokenStorage tokenStorage;
  final ApiClient apiClient;

  @override
  State<LoginPage> createState() => _LoginPageState();
}

enum _Step { phone, otp }

class _LoginPageState extends State<LoginPage> {
  late final AuthRepository _repo = AuthRepository(widget.apiClient);
  final _phoneCtrl = TextEditingController();
  final _codeCtrl = TextEditingController();

  _Step _step = _Step.phone;
  String _normalizedPhone = '';
  bool _isLoading = false;
  bool _isGoogleLoading = false;
  String? _error;
  String? _debugOtpCode;
  int _cooldownSeconds = 0;
  Timer? _cooldownTimer;

  @override
  void dispose() {
    _phoneCtrl.dispose();
    _codeCtrl.dispose();
    _cooldownTimer?.cancel();
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
                if (_step == _Step.phone) ..._buildPhoneStep() else ..._buildOtpStep(),
              ],
            ),
          ),
        ),
      ),
    );
  }

  List<Widget> _buildPhoneStep() {
    return [
      TextField(
        controller: _phoneCtrl,
        keyboardType: TextInputType.phone,
        textInputAction: TextInputAction.done,
        enabled: !_isLoading,
        onSubmitted: (_) => _requestOtp(),
        decoration: const InputDecoration(
          labelText: 'Số điện thoại',
          hintText: '090 123 4567',
          border: OutlineInputBorder(borderRadius: AppRadius.mdAll),
          prefixIcon: Icon(Icons.phone_outlined),
        ),
      ),
      const SizedBox(height: AppSpacing.md),
      if (_error != null) _ErrorText(_error!),
      const SizedBox(height: AppSpacing.xl),
      AppButton.primary(
        label: 'Gửi mã OTP',
        isLoading: _isLoading,
        onPressed: _requestOtp,
      ),
      const SizedBox(height: AppSpacing.lg),
      const _OrDivider(),
      const SizedBox(height: AppSpacing.lg),
      AppButton.outline(
        label: AppConfig.googleClientId.isEmpty
            ? 'Đăng nhập bằng Google (chưa cấu hình)'
            : 'Đăng nhập bằng Google',
        icon: Icons.g_mobiledata,
        isLoading: _isGoogleLoading,
        onPressed: AppConfig.googleClientId.isEmpty ? null : _loginWithGoogle,
      ),
    ];
  }

  List<Widget> _buildOtpStep() {
    return [
      Text(
        'Nhập mã OTP đã gửi tới $_normalizedPhone',
        textAlign: TextAlign.center,
        style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary),
      ),
      if (_debugOtpCode != null) ...[
        const SizedBox(height: AppSpacing.sm),
        Text(
          'Mã dev: $_debugOtpCode',
          textAlign: TextAlign.center,
          style: const TextStyle(color: AppColors.warning, fontWeight: FontWeight.w600),
        ),
      ],
      const SizedBox(height: AppSpacing.lg),
      TextField(
        controller: _codeCtrl,
        keyboardType: TextInputType.number,
        textInputAction: TextInputAction.done,
        enabled: !_isLoading,
        maxLength: 6,
        onSubmitted: (_) => _verifyOtp(),
        decoration: const InputDecoration(
          labelText: 'Mã OTP (6 số)',
          counterText: '',
          border: OutlineInputBorder(borderRadius: AppRadius.mdAll),
          prefixIcon: Icon(Icons.password_outlined),
        ),
      ),
      if (_error != null) _ErrorText(_error!),
      const SizedBox(height: AppSpacing.lg),
      AppButton.primary(
        label: 'Xác nhận',
        isLoading: _isLoading,
        onPressed: _verifyOtp,
      ),
      const SizedBox(height: AppSpacing.md),
      AppButton.text(
        label: _cooldownSeconds > 0 ? 'Gửi lại mã (${_cooldownSeconds}s)' : 'Gửi lại mã',
        onPressed: _cooldownSeconds > 0 ? null : _requestOtp,
      ),
      AppButton.text(
        label: 'Đổi số điện thoại',
        onPressed: () => setState(() {
          _step = _Step.phone;
          _error = null;
          _debugOtpCode = null;
        }),
      ),
    ];
  }

  Future<void> _requestOtp() async {
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
      final result = await _repo.requestOtp(phone);
      if (!mounted) return;
      setState(() {
        _normalizedPhone = phone;
        _step = _Step.otp;
        _debugOtpCode = result.debugOtpCode;
        _isLoading = false;
      });
      _startCooldown(60);
    } on ApiException catch (e) {
      if (mounted) {
        setState(() {
          _error = e.statusCode == 0 ? e.message : _friendlyOtpError(e);
          _isLoading = false;
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() {
          _error = 'Không gửi được mã. Vui lòng thử lại.';
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _verifyOtp() async {
    if (_isLoading) return;
    final code = _codeCtrl.text.trim();
    if (code.length != 6) {
      setState(() => _error = 'Mã OTP gồm 6 số');
      return;
    }

    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final result = await _repo.verifyOtp(phone: _normalizedPhone, code: code);
      widget.kycGate.reset();
      await widget.authState.login(
        accessToken: result.accessToken,
        refreshToken: result.refreshToken,
        riderId: result.riderId,
        storage: widget.tokenStorage,
      );
      // GoRouter's refreshListenable redirects (via Splash → KYC gate) automatically.
    } on ApiException catch (e) {
      if (mounted) {
        setState(() {
          _error = e.statusCode == 0 ? e.message : 'Mã OTP không đúng hoặc đã hết hạn.';
          _isLoading = false;
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() {
          _error = 'Xác thực thất bại. Vui lòng thử lại.';
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _loginWithGoogle() async {
    if (_isGoogleLoading) return;
    setState(() {
      _isGoogleLoading = true;
      _error = null;
    });
    try {
      final googleSignIn = GoogleSignIn(
        clientId: AppConfig.googleClientId,
        scopes: const ['email'],
      );
      final account = await googleSignIn.signIn();
      if (account == null) {
        if (mounted) setState(() => _isGoogleLoading = false);
        return;
      }
      final googleAuth = await account.authentication;
      final idToken = googleAuth.idToken;
      if (idToken == null) {
        throw const ApiException(statusCode: 0, message: 'Không lấy được id_token từ Google.');
      }
      final result = await _repo.loginWithGoogle(idToken);
      widget.kycGate.reset();
      await widget.authState.login(
        accessToken: result.accessToken,
        refreshToken: result.refreshToken,
        riderId: result.riderId,
        storage: widget.tokenStorage,
      );
    } on ApiException catch (e) {
      if (mounted) {
        setState(() {
          _error = e.statusCode == 0 ? e.message : 'Đăng nhập Google thất bại. Vui lòng thử lại.';
          _isGoogleLoading = false;
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() {
          _error = 'Đăng nhập Google thất bại. Vui lòng thử lại.';
          _isGoogleLoading = false;
        });
      }
    }
  }

  void _startCooldown(int seconds) {
    _cooldownTimer?.cancel();
    setState(() => _cooldownSeconds = seconds);
    _cooldownTimer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (!mounted) {
        timer.cancel();
        return;
      }
      setState(() {
        _cooldownSeconds -= 1;
        if (_cooldownSeconds <= 0) timer.cancel();
      });
    });
  }

  String _friendlyOtpError(ApiException e) {
    if (e.statusCode == 429) return 'Vui lòng đợi trước khi gửi lại mã.';
    return e.message;
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

class _ErrorText extends StatelessWidget {
  const _ErrorText(this.message);

  final String message;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(top: AppSpacing.sm),
      child: Text(
        message,
        style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppColors.error),
        textAlign: TextAlign.center,
      ),
    );
  }
}

class _OrDivider extends StatelessWidget {
  const _OrDivider();

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        const Expanded(child: Divider(color: AppColors.divider)),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: AppSpacing.sm),
          child: Text('hoặc', style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppColors.textTertiary)),
        ),
        const Expanded(child: Divider(color: AppColors.divider)),
      ],
    );
  }
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
