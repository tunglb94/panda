import 'dart:async';

import 'package:flutter/material.dart';

import '../../../../core/network/api_client.dart';
import '../../data/trip_offer_repository.dart';

enum _PageState {
  polling,
  offerAvailable,
  acting,
  accepted,
  error,
}

class TripPage extends StatefulWidget {
  const TripPage({super.key, required this.apiClient});

  final ApiClient apiClient;

  @override
  State<TripPage> createState() => _TripPageState();
}

class _TripPageState extends State<TripPage> {
  late final TripOfferRepository _repo;

  _PageState _state = _PageState.polling;
  TripOffer? _offer;
  String? _errorMessage;
  int _countdownSeconds = 0;

  Timer? _pollTimer;
  Timer? _countdownTimer;
  bool _isPollingActive = false;

  @override
  void initState() {
    super.initState();
    _repo = TripOfferRepository(apiClient: widget.apiClient);
    _poll();
    _pollTimer = Timer.periodic(const Duration(seconds: 5), (_) => _poll());
  }

  @override
  void dispose() {
    _pollTimer?.cancel();
    _countdownTimer?.cancel();
    super.dispose();
  }

  Future<void> _poll() async {
    if (_isPollingActive) return;
    if (_state == _PageState.acting || _state == _PageState.accepted) return;
    _isPollingActive = true;
    try {
      final offer = await _repo.getCurrentOffer();
      if (!mounted) return;
      if (offer == null) {
        if (_state == _PageState.offerAvailable) {
          _countdownTimer?.cancel();
          setState(() {
            _state = _PageState.polling;
            _offer = null;
          });
        } else if (_state == _PageState.error) {
          setState(() => _state = _PageState.polling);
        }
      } else {
        if (_state != _PageState.offerAvailable ||
            _offer?.tripId != offer.tripId) {
          _startCountdown(offer);
          setState(() {
            _state = _PageState.offerAvailable;
            _offer = offer;
          });
        }
      }
    } on ApiException catch (e) {
      if (!mounted) return;
      if (_state == _PageState.polling) {
        setState(() {
          _state = _PageState.error;
          _errorMessage = e.message;
        });
      }
    } finally {
      _isPollingActive = false;
    }
  }

  void _startCountdown(TripOffer offer) {
    _countdownTimer?.cancel();
    _countdownSeconds =
        offer.offerExpiresAt.difference(DateTime.now().toUtc()).inSeconds;
    if (_countdownSeconds < 0) _countdownSeconds = 0;
    _countdownTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      if (!mounted) return;
      setState(() {
        _countdownSeconds =
            (_offer?.offerExpiresAt.difference(DateTime.now().toUtc()).inSeconds ?? 0)
                .clamp(0, 999);
      });
      if (_countdownSeconds == 0) {
        _countdownTimer?.cancel();
        if (mounted && _state == _PageState.offerAvailable) {
          setState(() {
            _state = _PageState.polling;
            _offer = null;
          });
        }
      }
    });
  }

  Future<void> _onAccept() async {
    final offer = _offer;
    if (offer == null) return;
    setState(() => _state = _PageState.acting);
    _countdownTimer?.cancel();
    try {
      await _repo.acceptOffer(offer.tripId);
      if (!mounted) return;
      setState(() {
        _state = _PageState.accepted;
        _offer = null;
      });
      Future.delayed(const Duration(seconds: 3), () {
        if (mounted && _state == _PageState.accepted) {
          setState(() => _state = _PageState.polling);
        }
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _state = _PageState.error;
        _errorMessage = e.message;
      });
    }
  }

  Future<void> _onReject() async {
    final offer = _offer;
    if (offer == null) return;
    setState(() => _state = _PageState.acting);
    _countdownTimer?.cancel();
    try {
      await _repo.rejectOffer(offer.tripId);
      if (!mounted) return;
      setState(() {
        _state = _PageState.polling;
        _offer = null;
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _state = _PageState.error;
        _errorMessage = e.message;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Trip Offers')),
      body: SafeArea(child: _buildBody()),
    );
  }

  Widget _buildBody() {
    return switch (_state) {
      _PageState.polling => _PollingView(onRetry: _poll),
      _PageState.offerAvailable => _OfferCard(
          offer: _offer!,
          countdown: _countdownSeconds,
          onAccept: _onAccept,
          onReject: _onReject,
        ),
      _PageState.acting => const Center(child: CircularProgressIndicator()),
      _PageState.accepted => const _AcceptedView(),
      _PageState.error => _ErrorView(
          message: _errorMessage ?? 'An error occurred',
          onRetry: _poll,
        ),
    };
  }
}

// ─── Sub-widgets ─────────────────────────────────────────────────────────────

class _PollingView extends StatelessWidget {
  const _PollingView({required this.onRetry});

  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.directions_car_outlined, size: 64, color: cs.outline),
          const SizedBox(height: 16),
          Text(
            'Waiting for trip offers…',
            style: Theme.of(context)
                .textTheme
                .titleMedium
                ?.copyWith(color: cs.onSurfaceVariant),
          ),
          const SizedBox(height: 8),
          const SizedBox(
            width: 24,
            height: 24,
            child: CircularProgressIndicator(strokeWidth: 2),
          ),
        ],
      ),
    );
  }
}

class _OfferCard extends StatelessWidget {
  const _OfferCard({
    required this.offer,
    required this.countdown,
    required this.onAccept,
    required this.onReject,
  });

  final TripOffer offer;
  final int countdown;
  final VoidCallback onAccept;
  final VoidCallback onReject;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final cs = theme.colorScheme;
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Card(
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    'New Trip Request',
                    style: theme.textTheme.titleLarge
                        ?.copyWith(fontWeight: FontWeight.bold),
                  ),
                  _CountdownBadge(seconds: countdown),
                ],
              ),
              const SizedBox(height: 20),
              _AddressRow(
                icon: Icons.location_on,
                color: cs.primary,
                label: 'Pickup',
                address: offer.pickupAddress,
              ),
              const SizedBox(height: 12),
              _AddressRow(
                icon: Icons.flag,
                color: cs.error,
                label: 'Destination',
                address: offer.dropoffAddress,
              ),
              const SizedBox(height: 12),
              Row(
                children: [
                  _InfoChip(
                    icon: Icons.straighten,
                    label: '—',
                    sublabel: 'Distance',
                  ),
                  const SizedBox(width: 12),
                  _InfoChip(
                    icon: Icons.attach_money,
                    label: '—',
                    sublabel: 'Est. fare',
                  ),
                ],
              ),
              const SizedBox(height: 24),
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton(
                      onPressed: onReject,
                      style: OutlinedButton.styleFrom(
                        foregroundColor: cs.error,
                        side: BorderSide(color: cs.error),
                        padding: const EdgeInsets.symmetric(vertical: 16),
                      ),
                      child: const Text('Reject'),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: FilledButton(
                      onPressed: onAccept,
                      style: FilledButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 16),
                      ),
                      child: const Text('Accept'),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _CountdownBadge extends StatelessWidget {
  const _CountdownBadge({required this.seconds});

  final int seconds;

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final isUrgent = seconds <= 10;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(
        color: isUrgent ? cs.errorContainer : cs.secondaryContainer,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Text(
        '${seconds}s',
        style: Theme.of(context).textTheme.labelLarge?.copyWith(
              color: isUrgent ? cs.onErrorContainer : cs.onSecondaryContainer,
              fontWeight: FontWeight.bold,
            ),
      ),
    );
  }
}

class _AddressRow extends StatelessWidget {
  const _AddressRow({
    required this.icon,
    required this.color,
    required this.label,
    required this.address,
  });

  final IconData icon;
  final Color color;
  final String label;
  final String address;

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, color: color, size: 20),
        const SizedBox(width: 8),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(label,
                  style: Theme.of(context)
                      .textTheme
                      .labelSmall
                      ?.copyWith(color: Theme.of(context).colorScheme.outline)),
              Text(address, style: Theme.of(context).textTheme.bodyMedium),
            ],
          ),
        ),
      ],
    );
  }
}

class _InfoChip extends StatelessWidget {
  const _InfoChip({
    required this.icon,
    required this.label,
    required this.sublabel,
  });

  final IconData icon;
  final String label;
  final String sublabel;

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
        decoration: BoxDecoration(
          color: cs.surfaceContainerHighest,
          borderRadius: BorderRadius.circular(8),
        ),
        child: Row(
          children: [
            Icon(icon, size: 16, color: cs.onSurfaceVariant),
            const SizedBox(width: 6),
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(label,
                    style: Theme.of(context)
                        .textTheme
                        .titleSmall
                        ?.copyWith(fontWeight: FontWeight.bold)),
                Text(sublabel,
                    style: Theme.of(context).textTheme.labelSmall?.copyWith(
                        color: cs.onSurfaceVariant)),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _AcceptedView extends StatelessWidget {
  const _AcceptedView();

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.check_circle_outline, size: 64, color: cs.primary),
          const SizedBox(height: 16),
          Text(
            'Trip Accepted!',
            style: Theme.of(context)
                .textTheme
                .titleLarge
                ?.copyWith(color: cs.primary),
          ),
          const SizedBox(height: 8),
          Text(
            'Head to the pickup location.',
            style: Theme.of(context)
                .textTheme
                .bodyMedium
                ?.copyWith(color: cs.onSurfaceVariant),
          ),
        ],
      ),
    );
  }
}

class _ErrorView extends StatelessWidget {
  const _ErrorView({required this.message, required this.onRetry});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.wifi_off_outlined, size: 48, color: cs.error),
            const SizedBox(height: 16),
            Text(
              message,
              textAlign: TextAlign.center,
              style: Theme.of(context)
                  .textTheme
                  .bodyMedium
                  ?.copyWith(color: cs.onSurfaceVariant),
            ),
            const SizedBox(height: 16),
            FilledButton.icon(
              onPressed: onRetry,
              icon: const Icon(Icons.refresh),
              label: const Text('Retry'),
            ),
          ],
        ),
      ),
    );
  }
}
