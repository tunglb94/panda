import 'package:flutter/material.dart';
import '../../../../core/auth/auth_state.dart';
import '../../../../core/network/api_client.dart';
import '../../../../core/storage/token_storage.dart';
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
    return Scaffold(
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 360),
          child: Card(
            child: Padding(
              padding: const EdgeInsets.all(32),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Icon(Icons.admin_panel_settings_outlined, size: 48, color: Theme.of(context).colorScheme.primary),
                  const SizedBox(height: 16),
                  Text('Panda Admin', style: Theme.of(context).textTheme.headlineSmall, textAlign: TextAlign.center),
                  const SizedBox(height: 4),
                  Text(
                    'KYC Review Dashboard',
                    style: Theme.of(context).textTheme.bodyMedium,
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 32),
                  TextField(
                    controller: _phoneCtrl,
                    keyboardType: TextInputType.phone,
                    textInputAction: TextInputAction.done,
                    enabled: !_isLoading,
                    onSubmitted: (_) => _login(),
                    decoration: const InputDecoration(
                      labelText: 'Số điện thoại Admin',
                      hintText: '090 000 0099',
                      prefixIcon: Icon(Icons.phone_outlined),
                    ),
                  ),
                  const SizedBox(height: 12),
                  if (_error != null)
                    Text(_error!, style: TextStyle(color: Theme.of(context).colorScheme.error), textAlign: TextAlign.center),
                  const SizedBox(height: 16),
                  FilledButton(
                    onPressed: _isLoading ? null : _login,
                    child: _isLoading
                        ? const SizedBox(
                            width: 20,
                            height: 20,
                            child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                          )
                        : const Text('Đăng nhập'),
                  ),
                ],
              ),
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
      final result = await repo.loginAdmin(phone);
      await widget.authState.login(
        accessToken: result.accessToken,
        refreshToken: result.refreshToken,
        adminId: result.adminId,
        storage: widget.tokenStorage,
      );
      // GoRouter's refreshListenable redirects to the dashboard automatically.
    } on ApiException catch (e) {
      if (mounted) {
        setState(() {
          _error = e.statusCode == 404
              ? 'Không tìm thấy tài khoản Admin với số này'
              : 'Đăng nhập thất bại. Vui lòng thử lại.';
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

/// Same normalization as apps/driver and apps/rider — accepts local-style
/// input (`090 123 4567`) and converts it to the `+84...` form the backend
/// stores phone numbers in.
String _normalizeVietnamesePhone(String raw) {
  final digitsOnly = raw.replaceAll(RegExp(r'[^0-9+]'), '');
  if (digitsOnly.startsWith('+')) return digitsOnly;
  if (digitsOnly.startsWith('84')) return '+$digitsOnly';
  if (digitsOnly.startsWith('0')) return '+84${digitsOnly.substring(1)}';
  return '+84$digitsOnly';
}
