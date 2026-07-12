import 'dart:async';

import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/features/chat/data/chat_repository.dart';
import 'package:rider/features/contact/data/contact_repository.dart';
import 'package:rider/features/contact/domain/models/contact_info.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';
import 'package:rider/features/trip/data/trip_repository.dart';
import 'package:rider/features/trip/domain/models/driver_profile.dart';
import 'package:rider/features/trip/domain/models/rider_trip_status.dart';
import 'package:rider/shared/utils/currency_format.dart';
import 'package:rider/shared/widgets/app_button.dart';
import 'package:rider/shared/widgets/app_snackbar.dart';
import 'package:rider/shared/widgets/mascot_image.dart';

import 'driver_arriving_view.dart';
import 'driver_assigned_view.dart';
import 'searching_driver_view.dart';
import 'trip_cancelled_view.dart';
import 'trip_completed_view.dart';
import 'trip_in_progress_view.dart';
import '../widgets/trip_receipt_sheet.dart';

/// Live trip lifecycle screen.
///
/// Polls `GET /api/v1/rides/{tripId}` every 5 seconds and animates through
/// [RiderTripStatus] states. Polling stops automatically on a terminal state
/// (completed or cancelled).
///
/// [onDriverAssigned] is called once when the status first transitions to
/// [RiderTripStatus.driverAssigned] with a non-empty driverId. Callers that
/// have a map view can use this to start live driver tracking.
class TripLifecyclePage extends StatefulWidget {
  const TripLifecyclePage({
    super.key,
    required this.tripId,
    required this.tripSelection,
    required this.apiClient,
    this.onDriverAssigned,
  });

  final String tripId;
  final TripSelection tripSelection;
  final ApiClient apiClient;
  final void Function(String driverId)? onDriverAssigned;

  @override
  State<TripLifecyclePage> createState() => _TripLifecyclePageState();
}

class _TripLifecyclePageState extends State<TripLifecyclePage>
    with WidgetsBindingObserver {
  RiderTripStatus _status = RiderTripStatus.searchingDriver;
  Timer? _pollTimer;
  bool _isPolling = false;
  int _finalFareCents = 0;
  String _currency = '';
  bool _trackingStarted = false;
  String? _pollError;
  // Which payment method button is currently mid-request ('cash'/'wallet'),
  // or null when idle. Tracking the specific method (not just a bool) lets
  // the UI show the loading morph on the exact button the rider tapped
  // while disabling — not hiding — the other one, and doubles as the
  // double-submit guard in [_pay].
  String? _payingMethod;
  // The method that actually succeeded, remembered for the Payment Success
  // screen ("Phương thức thanh toán") — the backend does not echo this back
  // on any response (see `PayRide`'s gateway handler), so this is the only
  // source of truth for it, and only for the trip just paid in this session.
  String? _paidMethod;
  DriverProfile _driverProfile = DriverProfile.loading;
  String _driverId = '';
  ContactInfo? _contactInfo;
  int _chatUnreadCount = 0;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _poll();
    _pollTimer = Timer.periodic(const Duration(seconds: 5), (_) => _poll());
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _pollTimer?.cancel();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed && !_status.isTerminal) {
      _poll();
    }
  }

  Future<void> _poll() async {
    if (_isPolling) return;
    _isPolling = true;
    try {
      final detail =
          await TripRepository(widget.apiClient).getTrip(widget.tripId);
      if (!mounted) return;
      final newStatus = _mapStatus(detail.status);
      setState(() {
        _status = newStatus;
        _finalFareCents = detail.finalFareCents;
        _currency = detail.currency;
        _pollError = null;
      });
      if (detail.driverId.isNotEmpty) {
        _driverId = detail.driverId;
      }
      if (newStatus == RiderTripStatus.driverAssigned &&
          !_trackingStarted &&
          detail.driverId.isNotEmpty) {
        _trackingStarted = true;
        widget.onDriverAssigned?.call(detail.driverId);
        _fetchDriverProfile(detail.driverId);
        _fetchContactInfo();
        _fetchChatUnread();
      }
      if (newStatus.isTerminal) {
        _pollTimer?.cancel();
        _pollTimer = null;
      }
    } on ApiException catch (e) {
      // statusCode 0 is only ever thrown client-side by ApiClient itself
      // (timeout/connectivity) with copy that's already Vietnamese and
      // safe to show verbatim; any real HTTP status is a raw backend
      // message and must never reach the rider as-is.
      if (mounted) {
        setState(() => _pollError = e.statusCode == 0 ? e.message : 'Không thể tải trạng thái chuyến đi. Đang thử lại…');
      }
    } catch (_) {
      if (mounted) setState(() => _pollError = 'Lỗi mạng. Đang thử lại…');
    } finally {
      _isPolling = false;
    }
  }

  Future<void> _fetchDriverProfile(String driverId) async {
    try {
      final profile = await TripRepository(widget.apiClient).fetchDriverProfile(driverId);
      if (mounted) setState(() => _driverProfile = profile);
    } on ApiException catch (_) {
      // Non-fatal: driver card stays on loading placeholder.
    } catch (_) {
      // Ignore.
    }
  }

  /// Best-effort — the driver name/rating (Part 4) are a nice-to-have on
  /// top of the vehicle info already shown; a failure here never blocks or
  /// errors the trip lifecycle screen.
  Future<void> _fetchContactInfo() async {
    try {
      final contact = await ContactRepository(widget.apiClient).getContact(widget.tripId);
      if (mounted) setState(() => _contactInfo = contact);
    } catch (_) {
      // Ignore — card just shows vehicle info without name/rating.
    }
  }

  /// Best-effort snapshot of the Chat unread badge (Part 5) — fetched once
  /// when the driver is assigned, not continuously refreshed.
  Future<void> _fetchChatUnread() async {
    try {
      final conv = await ChatRepository(widget.apiClient).getOrCreateConversation(widget.tripId);
      if (mounted) setState(() => _chatUnreadCount = conv.unreadCount);
    } catch (_) {
      // Ignore — badge just stays hidden.
    }
  }

  RiderTripStatus _mapStatus(String s) => switch (s) {
        'pending' || 'searching' => RiderTripStatus.searchingDriver,
        'driver_assigned' => RiderTripStatus.driverAssigned,
        'driver_arrived' => RiderTripStatus.driverArriving,
        'in_progress' => RiderTripStatus.inProgress,
        'completed' => RiderTripStatus.completed,
        'cancelled' => RiderTripStatus.cancelled,
        'payment_pending' => RiderTripStatus.paymentPending,
        'payment_success' => RiderTripStatus.paymentSuccess,
        'settled' => RiderTripStatus.settled,
        _ => _status,
      };

  Future<void> _pay(String method) async {
    if (_payingMethod != null) return; // double-submit / double-click guard
    setState(() {
      _payingMethod = method;
      _pollError = null;
    });
    // Whether we should force an immediate status refresh before releasing
    // the "paying" lock. Skipped on a genuine failure so the error message
    // set below isn't clobbered by _poll()'s own error handling.
    var refreshStatus = true;
    var succeeded = false;
    try {
      await TripRepository(widget.apiClient)
          .payRide(widget.tripId, paymentMethod: method);
      succeeded = true;
    } on ApiException catch (e) {
      if (_isAlreadySettledError(e)) {
        // The trip was already marked paid — most likely this exact request
        // succeeded once already (e.g. a retried tap during the old refresh
        // gap) and the backend correctly rejected the duplicate. From the
        // rider's point of view the payment did succeed, so treat it as
        // such — show an explicit confirmation rather than the backend's
        // internal precondition message, and never surface that raw error.
        succeeded = true;
        if (mounted) {
          AppSnackbar.success(context, 'Chuyến đi đã được thanh toán.');
        }
      } else if (e.statusCode == 0) {
        // Client-side timeout/connectivity message from ApiClient — already
        // Vietnamese, safe to show verbatim.
        refreshStatus = false;
        if (mounted) setState(() => _pollError = e.message);
      } else {
        // Any other backend rejection (voucher expired, promotion rejected,
        // insufficient balance, etc.) — never show the raw backend string.
        refreshStatus = false;
        if (mounted) {
          setState(() => _pollError = _genericPaymentError);
        }
      }
    } catch (_) {
      refreshStatus = false;
      if (mounted) {
        setState(() => _pollError = _genericPaymentError);
      }
    }
    if (succeeded) _paidMethod = method;
    if (refreshStatus) {
      // Pull the authoritative trip status now instead of waiting for the
      // next 5s poll tick, so the payment buttons never get a chance to
      // reappear once the trip is actually settled.
      await _poll();
    }
    if (mounted) setState(() => _payingMethod = null);
  }

  static const _genericPaymentError = 'Thanh toán thất bại hoặc mất kết nối. Nhấn nút để thử lại.';

  /// True when [e] is the backend's "already paid" precondition failure —
  /// see `PayTrip`/`PayRide` in the trip/booking services, which reject a
  /// second payment attempt once the trip has moved past `payment_pending`.
  bool _isAlreadySettledError(ApiException e) {
    final msg = e.message.toLowerCase();
    return msg.contains('cannot be marked paid') ||
        (msg.contains('paid') && msg.contains('settled'));
  }

  String get _fareText {
    if (_finalFareCents <= 0) return '—';
    return formatMoney(_finalFareCents, _currency);
  }

  Future<void> _submitRating(int stars, String? comment) async {
    await TripRepository(widget.apiClient).submitRating(
      widget.tripId,
      stars,
      driverId: _driverId,
      comment: comment,
    );
  }

  void _cancelRide() {
    _pollTimer?.cancel();
    _pollTimer = null;
    // Best-effort: fire-and-forget so the UI isn't blocked.
    TripRepository(widget.apiClient).cancelRide(widget.tripId).ignore();
    Navigator.of(context).pop();
  }
  void _finish() => Navigator.of(context).pop();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Chuyến đi của bạn'),
        automaticallyImplyLeading: false,
      ),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                children: [
                  AnimatedSize(
                    duration: const Duration(milliseconds: 220),
                    curve: Curves.easeOut,
                    alignment: Alignment.topCenter,
                    child: _pollError == null
                        ? const SizedBox(width: double.infinity)
                        : Padding(
                            padding: const EdgeInsets.only(bottom: 12),
                            child: AnimatedSwitcher(
                              duration: const Duration(milliseconds: 220),
                              child: Text(
                                _pollError!,
                                key: ValueKey(_pollError),
                                style: TextStyle(
                                  color: Theme.of(context).colorScheme.error,
                                  fontSize: 13,
                                ),
                                textAlign: TextAlign.center,
                              ),
                            ),
                          ),
                  ),
                  AnimatedSwitcher(
                    duration: const Duration(milliseconds: 400),
                    transitionBuilder: (child, animation) => FadeTransition(
                      opacity: animation,
                      child: SlideTransition(
                        position: Tween<Offset>(
                          begin: const Offset(0, 0.05),
                          end: Offset.zero,
                        ).animate(animation),
                        child: child,
                      ),
                    ),
                    child: _buildCurrentView(),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildCurrentView() {
    return switch (_status) {
      RiderTripStatus.searchingDriver => SearchingDriverView(
          key: const ValueKey(RiderTripStatus.searchingDriver),
          tripSelection: widget.tripSelection,
          onCancel: _cancelRide,
        ),
      RiderTripStatus.driverAssigned => DriverAssignedView(
          key: const ValueKey(RiderTripStatus.driverAssigned),
          tripSelection: widget.tripSelection,
          driver: _driverProfile,
          contact: _contactInfo,
          tripId: widget.tripId,
          apiClient: widget.apiClient,
          chatUnreadCount: _chatUnreadCount,
          onCancel: _cancelRide,
        ),
      RiderTripStatus.driverArriving => DriverArrivingView(
          key: const ValueKey(RiderTripStatus.driverArriving),
          tripSelection: widget.tripSelection,
          driver: _driverProfile,
          contact: _contactInfo,
          tripId: widget.tripId,
          apiClient: widget.apiClient,
          chatUnreadCount: _chatUnreadCount,
          onCancel: _cancelRide,
        ),
      RiderTripStatus.inProgress => TripInProgressView(
          key: const ValueKey(RiderTripStatus.inProgress),
          tripSelection: widget.tripSelection,
          driver: _driverProfile,
          contact: _contactInfo,
          tripId: widget.tripId,
          apiClient: widget.apiClient,
          chatUnreadCount: _chatUnreadCount,
        ),
      RiderTripStatus.completed => TripCompletedView(
          key: const ValueKey(RiderTripStatus.completed),
          tripSelection: widget.tripSelection,
          driver: _driverProfile,
          fareText: _fareText,
          onDone: _finish,
        ),
      RiderTripStatus.cancelled => TripCancelledView(
          key: const ValueKey(RiderTripStatus.cancelled),
          onDone: _finish,
        ),
      RiderTripStatus.paymentPending => _PaymentPendingView(
          key: const ValueKey(RiderTripStatus.paymentPending),
          fareText: _fareText,
          payingMethod: _payingMethod,
          hasError: _pollError != null,
          onPayCash: () => _pay('cash'),
          onPayWallet: () => _pay('wallet'),
        ),
      RiderTripStatus.paymentSuccess ||
      RiderTripStatus.settled =>
        _PostTripView(
          key: const ValueKey('payment_done'),
          fareText: _fareText,
          paidMethod: _paidMethod,
          tripId: widget.tripId,
          apiClient: widget.apiClient,
          onDone: _finish,
          onSubmitRating: _submitRating,
        ),
    };
  }
}

// ─── Payment views ────────────────────────────────────────────────────────────

class _PaymentPendingView extends StatelessWidget {
  const _PaymentPendingView({
    super.key,
    required this.fareText,
    required this.payingMethod,
    required this.hasError,
    required this.onPayCash,
    required this.onPayWallet,
  });

  final String fareText;

  /// 'cash'/'wallet' while that specific button's request is in flight,
  /// null when idle. Section 7 (Payment UX): both buttons stay visible and
  /// tappable-when-idle at all times — the one just tapped shows
  /// [AppButton]'s built-in loading morph (which also disables it), the
  /// other is explicitly disabled so a rider can't fire two payment
  /// methods at once, but neither button is ever hidden/replaced by a bare
  /// spinner the way this view used to work.
  final String? payingMethod;

  /// Whether the trip lifecycle page currently has a payment error message
  /// displayed above this view — used only to relabel the CTA hint text as
  /// an explicit "thử lại" (retry) prompt.
  final bool hasError;

  final VoidCallback onPayCash;
  final VoidCallback onPayWallet;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final cs = theme.colorScheme;
    final busy = payingMethod != null;
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          AnimatedSwitcher(
            duration: const Duration(milliseconds: 300),
            transitionBuilder: (child, animation) => ScaleTransition(scale: animation, child: FadeTransition(opacity: animation, child: child)),
            child: hasError
                ? const MascotImage(
                    key: ValueKey('payment_error_mascot'),
                    asset: 'mascot_no_connection.png',
                    size: MascotSize.medium,
                    animation: MascotAnimation.scale,
                    semanticLabel: 'Thanh toán chưa thành công',
                  )
                : Icon(Icons.payment, key: const ValueKey('payment_icon'), size: 64, color: cs.primary),
          ),
          const SizedBox(height: 12),
          Text(
            'Thanh toán',
            style: theme.textTheme.headlineSmall
                ?.copyWith(fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          Text(
            hasError
                ? 'Thanh toán chưa thành công. Vui lòng thử lại.'
                : 'Vui lòng thanh toán để hoàn tất chuyến đi',
            style: theme.textTheme.bodyMedium
                ?.copyWith(color: hasError ? cs.error : cs.onSurfaceVariant),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 24),
          Card(
            child: Padding(
              padding: const EdgeInsets.all(20),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('Tổng cước phí',
                      style: theme.textTheme.bodyLarge),
                  Flexible(
                    child: Text(
                      fareText,
                      textAlign: TextAlign.right,
                      style: theme.textTheme.headlineSmall?.copyWith(
                        fontWeight: FontWeight.bold,
                        color: cs.primary,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 24),
          AppButton.primary(
            label: hasError ? 'Thử lại · Tiền mặt' : 'Trả bằng tiền mặt',
            icon: Icons.money,
            isLoading: payingMethod == 'cash',
            onPressed: busy ? null : onPayCash,
          ),
          const SizedBox(height: 12),
          AppButton.outline(
            label: hasError ? 'Thử lại · Ví điện tử' : 'Trả bằng ví điện tử',
            icon: Icons.account_balance_wallet_outlined,
            isLoading: payingMethod == 'wallet',
            onPressed: busy ? null : onPayWallet,
          ),
        ],
      ),
    );
  }
}

class _PostTripView extends StatefulWidget {
  const _PostTripView({
    super.key,
    required this.fareText,
    required this.paidMethod,
    required this.tripId,
    required this.apiClient,
    required this.onDone,
    required this.onSubmitRating,
  });

  final String fareText;

  /// 'cash'/'wallet', or null if this trip was already settled before this
  /// app session started (e.g. resumed after being killed) — in that case
  /// there is no real source for the method (the backend does not echo it
  /// back on any response), so the row is omitted rather than guessed.
  final String? paidMethod;

  final String tripId;
  final ApiClient apiClient;
  final VoidCallback onDone;
  final Future<void> Function(int stars, String? comment) onSubmitRating;

  @override
  State<_PostTripView> createState() => _PostTripViewState();
}

class _PostTripViewState extends State<_PostTripView> {
  int _stars = 0;
  bool _submitted = false;
  bool _submitting = false;
  String? _error;
  bool _showSuccessSplash = true;
  Timer? _splashTimer;
  final _commentController = TextEditingController();

  String? get _paymentMethodLabel => switch (widget.paidMethod) {
        'cash' => 'tiền mặt',
        'wallet' => 'ví điện tử',
        _ => null,
      };

  @override
  void initState() {
    super.initState();
    // Brief success moment before handing off to the rating form. This
    // widget is keyed once per trip (see `_buildCurrentView`), so the timer
    // only ever fires the one time the view is first mounted, not on every
    // subsequent poll-driven rebuild.
    _splashTimer = Timer(const Duration(milliseconds: 1400), () {
      if (mounted) setState(() => _showSuccessSplash = false);
    });
  }

  @override
  void dispose() {
    _splashTimer?.cancel();
    _commentController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (_stars == 0 || _submitting) return;
    setState(() {
      _submitting = true;
      _error = null;
    });
    try {
      final comment = _commentController.text.trim();
      await widget.onSubmitRating(_stars, comment.isEmpty ? null : comment);
      if (mounted) setState(() { _submitted = true; _submitting = false; });
    } catch (_) {
      if (mounted) {
        setState(() {
          _submitting = false;
          _error = 'Không thể gửi đánh giá. Bạn có thể bỏ qua.';
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final cs = theme.colorScheme;

    if (_showSuccessSplash) {
      return Padding(
        key: const ValueKey('payment_success_splash'),
        padding: const EdgeInsets.all(16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const MascotImage(
              asset: 'mascot_success.png',
              size: MascotSize.large,
              animation: MascotAnimation.bounce,
              semanticLabel: 'Thanh toán thành công',
            ),
            const SizedBox(height: 16),
            Text(
              'Thanh toán hoàn tất',
              style: theme.textTheme.headlineSmall?.copyWith(
                fontWeight: FontWeight.bold,
                color: cs.primary,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              widget.fareText,
              style: theme.textTheme.headlineMedium?.copyWith(
                fontWeight: FontWeight.bold,
              ),
            ),
            if (_paymentMethodLabel != null) ...[
              const SizedBox(height: 4),
              Text(
                'Thanh toán bằng $_paymentMethodLabel',
                style: theme.textTheme.bodyMedium?.copyWith(color: cs.onSurfaceVariant),
              ),
            ],
            const SizedBox(height: 20),
            AppButton.outline(
              label: 'Xem hóa đơn',
              icon: Icons.receipt_long_outlined,
              onPressed: () => TripReceiptSheet.show(
                context,
                tripId: widget.tripId,
                apiClient: widget.apiClient,
                paymentMethodLabel: _paymentMethodLabel,
              ),
            ),
          ],
        ),
      );
    }

    return Padding(
      key: const ValueKey('payment_rating'),
      padding: const EdgeInsets.all(16),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (_submitted) ...[
            const MascotImage(
              asset: 'mascot_rating_5star.png',
              size: MascotSize.medium,
              animation: MascotAnimation.bounce,
              semanticLabel: 'Cảm ơn bạn đã đánh giá',
            ),
            const SizedBox(height: 12),
            Text(
              'Cảm ơn bạn đã sử dụng dịch vụ của Panda!',
              style: theme.textTheme.bodyMedium?.copyWith(color: cs.onSurfaceVariant),
            ),
            const SizedBox(height: 32),
            FilledButton(
              onPressed: widget.onDone,
              style: FilledButton.styleFrom(minimumSize: const Size.fromHeight(52)),
              child: const Text('Xong'),
            ),
          ] else ...[
            Text(
              'Đánh giá chuyến đi',
              style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w600),
            ),
            const SizedBox(height: 12),
            _StarRow(selected: _stars, onSelect: (s) => setState(() => _stars = s)),
            const SizedBox(height: 12),
            TextField(
              controller: _commentController,
              maxLines: 2,
              maxLength: 200,
              decoration: const InputDecoration(
                hintText: 'Nhận xét thêm (không bắt buộc)…',
                border: OutlineInputBorder(),
                contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              ),
            ),
            if (_error != null) ...[
              const SizedBox(height: 4),
              Text(_error!, style: TextStyle(color: cs.error, fontSize: 13)),
            ],
            const SizedBox(height: 8),
            if (_submitting)
              Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const CircularProgressIndicator(),
                  const SizedBox(height: 12),
                  Text(
                    'Đang gửi đánh giá…',
                    style: Theme.of(context)
                        .textTheme
                        .bodyMedium
                        ?.copyWith(color: cs.onSurfaceVariant),
                  ),
                ],
              )
            else ...[
              FilledButton(
                onPressed: _stars > 0 ? _submit : null,
                style: FilledButton.styleFrom(minimumSize: const Size.fromHeight(52)),
                child: const Text('Gửi đánh giá'),
              ),
              const SizedBox(height: 8),
              TextButton(
                onPressed: widget.onDone,
                child: const Text('Bỏ qua'),
              ),
            ],
          ],
        ],
      ),
    );
  }
}

class _StarRow extends StatelessWidget {
  const _StarRow({required this.selected, required this.onSelect});

  final int selected;
  final void Function(int) onSelect;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: List.generate(5, (i) {
        final star = i + 1;
        return IconButton(
          onPressed: () => onSelect(star),
          tooltip: '$star sao',
          icon: Icon(
            star <= selected ? Icons.star : Icons.star_border,
            size: 36,
            color: star <= selected ? Colors.amber : Colors.grey,
          ),
        );
      }),
    );
  }
}
