import 'package:flutter/material.dart';
import '../../../../core/auth/auth_state.dart';
import '../../../../core/network/api_client.dart';
import '../../../../core/storage/token_storage.dart';
import '../../data/availability_repository.dart';

class HomePage extends StatefulWidget {
  const HomePage({
    super.key,
    required this.authState,
    required this.tokenStorage,
    required this.apiClient,
  });

  final AuthState authState;
  final TokenStorage tokenStorage;
  final ApiClient apiClient;

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  late final AvailabilityRepository _availRepo;
  bool _isOnline = false;
  bool _isLoadingStatus = true;
  bool _isToggling = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _availRepo = AvailabilityRepository(widget.apiClient);
    _fetchAvailability();
  }

  Future<void> _fetchAvailability() async {
    try {
      final result = await _availRepo.getAvailability();
      if (mounted) setState(() => _isOnline = result.isOnline);
    } catch (_) {
      // Non-fatal — show last-known state (default offline)
    } finally {
      if (mounted) setState(() => _isLoadingStatus = false);
    }
  }

  Future<void> _toggle() async {
    if (_isToggling) return;
    setState(() {
      _isToggling = true;
      _error = null;
    });
    try {
      final result = _isOnline
          ? await _availRepo.goOffline()
          : await _availRepo.goOnline();
      if (mounted) setState(() => _isOnline = result.isOnline);
    } on ApiException catch (e) {
      if (mounted) setState(() => _error = e.message);
    } catch (_) {
      if (mounted) setState(() => _error = 'Could not update status. Please try again.');
    } finally {
      if (mounted) setState(() => _isToggling = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('FAIRRIDE Driver'),
        centerTitle: false,
      ),
      body: SafeArea(
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            _StatusCard(
              isOnline: _isOnline,
              isLoading: _isLoadingStatus || _isToggling,
              error: _error,
              onToggle: _toggle,
            ),
            const SizedBox(height: 16),
            const _SummaryCard(),
          ],
        ),
      ),
    );
  }
}

class _StatusCard extends StatelessWidget {
  const _StatusCard({
    required this.isOnline,
    required this.isLoading,
    this.error,
    required this.onToggle,
  });

  final bool isOnline;
  final bool isLoading;
  final String? error;
  final VoidCallback onToggle;

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                if (isLoading)
                  const SizedBox(
                    width: 12,
                    height: 12,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                else
                  Container(
                    width: 12,
                    height: 12,
                    decoration: BoxDecoration(
                      color: isOnline ? const Color(0xFF1A8C4E) : cs.outline,
                      shape: BoxShape.circle,
                    ),
                  ),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    isOnline ? 'You are online' : 'You are offline',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                ),
                Switch(
                  value: isOnline,
                  onChanged: isLoading ? null : (_) => onToggle(),
                ),
              ],
            ),
            if (error != null) ...[
              const SizedBox(height: 8),
              Text(
                error!,
                style: TextStyle(color: cs.error, fontSize: 12),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _SummaryCard extends StatelessWidget {
  const _SummaryCard();

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              "Today's Summary",
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 16),
            const Row(
              children: [
                _Stat(label: 'Trips', value: '0'),
                SizedBox(width: 32),
                _Stat(label: 'Earnings', value: '\$0.00'),
                SizedBox(width: 32),
                _Stat(label: 'Hours', value: '0.0'),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _Stat extends StatelessWidget {
  const _Stat({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          value,
          style: Theme.of(context)
              .textTheme
              .headlineSmall
              ?.copyWith(fontWeight: FontWeight.bold),
        ),
        Text(label, style: Theme.of(context).textTheme.bodySmall),
      ],
    );
  }
}
