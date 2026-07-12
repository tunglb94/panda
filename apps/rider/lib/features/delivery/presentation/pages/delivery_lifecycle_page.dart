import 'dart:async';

import 'package:flutter/material.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/features/delivery/domain/models/delivery_status.dart';
import 'package:rider/features/delivery/presentation/widgets/delivery_map_view.dart';
import 'package:rider/features/delivery/presentation/widgets/delivery_receipt_sheet.dart';
import 'package:rider/features/trip/data/trip_repository.dart';
import 'package:rider/shared/widgets/app_button.dart';
import 'package:rider/shared/widgets/app_card.dart';
import 'package:rider/shared/widgets/mascot_image.dart';

/// Delivery lifecycle screen — Phase 2/3 of the Delivery production pass.
/// Polls `GET /api/v1/rides/{tripId}` every 5s exactly like
/// `TripLifecyclePage`, but drives its state machine off `delivery_status`
/// (the gateway's best-effort Trip-service enrichment), not `trip_status` —
/// see `DeliveryStatus`'s doc comment for why `trip_status` alone can't
/// represent a delivery's actual progress. Entirely separate from
/// `TripLifecyclePage`; Ride's lifecycle screen is untouched.
class DeliveryLifecyclePage extends StatefulWidget {
  const DeliveryLifecyclePage({
    super.key,
    required this.tripId,
    required this.apiClient,
    required this.pickupAddress,
    required this.pickupLocation,
    required this.receiverAddress,
    required this.receiverLocation,
    required this.receiverName,
    this.receiverPhone,
    this.estimatedFareCents,
    this.currency,
  });

  final String tripId;
  final ApiClient apiClient;
  final String pickupAddress;
  final LatLng pickupLocation;
  final String receiverAddress;
  final LatLng receiverLocation;
  final String receiverName;

  /// Captured at booking time — see `DeliveryReceiptContent`'s doc comment
  /// for why the receipt shows this instead of a settled backend fare.
  final String? receiverPhone;
  final int? estimatedFareCents;
  final String? currency;

  @override
  State<DeliveryLifecyclePage> createState() => _DeliveryLifecyclePageState();
}

class _DeliveryLifecyclePageState extends State<DeliveryLifecyclePage>
    with WidgetsBindingObserver {
  Timer? _pollTimer;
  bool _isPolling = false;
  DeliveryStatus _status = DeliveryStatus.created;
  String? _pollError;

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
    // Offline recovery: an immediate poll on resume, same as
    // TripLifecyclePage — don't make the rider wait up to 5s for the next
    // scheduled tick after backgrounding/foregrounding the app.
    if (state == AppLifecycleState.resumed && !_status.isTerminal) {
      _poll();
    }
  }

  Future<void> _poll() async {
    if (_isPolling) return;
    _isPolling = true;
    try {
      final detail = await TripRepository(widget.apiClient).getTrip(widget.tripId);
      if (!mounted) return;
      setState(() {
        // An empty delivery_status (enrichment unavailable this poll, or
        // not caught up yet right after booking) keeps the last known
        // status rather than flashing to "unknown".
        if (detail.deliveryStatus.isNotEmpty) {
          _status = DeliveryStatus.fromWire(detail.deliveryStatus);
        }
        _pollError = null;
      });
      if (_status.isTerminal) {
        _pollTimer?.cancel();
        _pollTimer = null;
      }
    } on ApiException catch (e) {
      if (mounted) {
        setState(() => _pollError = e.statusCode == 0 ? e.message : 'Không thể tải trạng thái đơn hàng. Đang thử lại…');
      }
    } catch (_) {
      if (mounted) setState(() => _pollError = 'Lỗi mạng. Đang thử lại…');
    } finally {
      _isPolling = false;
    }
  }

  void _finish() => Navigator.of(context).pop();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Đơn giao hàng'), automaticallyImplyLeading: false),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(AppSpacing.lg),
              child: Column(
                children: [
                  if (_pollError != null)
                    Padding(
                      padding: const EdgeInsets.only(bottom: AppSpacing.md),
                      child: Text(
                        _pollError!,
                        style: TextStyle(color: Theme.of(context).colorScheme.error, fontSize: 13),
                        textAlign: TextAlign.center,
                      ),
                    ),
                  AnimatedSwitcher(
                    duration: const Duration(milliseconds: 400),
                    transitionBuilder: (child, animation) => FadeTransition(
                      opacity: animation,
                      child: SlideTransition(
                        position: Tween<Offset>(begin: const Offset(0, 0.05), end: Offset.zero).animate(animation),
                        child: child,
                      ),
                    ),
                    child: _buildBody(key: ValueKey(_status)),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildBody({required Key key}) {
    switch (_status) {
      case DeliveryStatus.created:
        return _StatusView(
          key: key,
          mascot: 'mascot_waiting.png',
          animation: MascotAnimation.none,
          title: 'Đang tìm tài xế',
          subtitle: 'Chúng tôi đang tìm tài xế gần điểm lấy hàng của bạn.',
          showRoute: false,
          data: this,
        );
      case DeliveryStatus.accepted:
        return _StatusView(
          key: key,
          mascot: 'mascot_waiting.png',
          animation: MascotAnimation.fade,
          title: 'Tài xế đang đến điểm lấy hàng',
          subtitle: 'Tài xế đã nhận đơn và đang trên đường đến chỗ bạn.',
          showRoute: true,
          data: this,
        );
      case DeliveryStatus.parcelPickedUp:
        return _StatusView(
          key: key,
          mascot: 'mascot_waiting.png',
          animation: MascotAnimation.fade,
          title: 'Đã lấy hàng',
          subtitle: 'Tài xế đã lấy hàng và chuẩn bị giao đến người nhận.',
          showRoute: true,
          data: this,
        );
      case DeliveryStatus.inDelivery:
        return _StatusView(
          key: key,
          mascot: 'mascot_waiting.png',
          animation: MascotAnimation.fade,
          title: 'Đang giao hàng',
          subtitle: 'Tài xế đang trên đường giao hàng đến ${widget.receiverName.isEmpty ? "người nhận" : widget.receiverName}.',
          showRoute: true,
          data: this,
        );
      case DeliveryStatus.delivered:
      case DeliveryStatus.completed:
        return _DeliveredView(
          key: key,
          tripId: widget.tripId,
          apiClient: widget.apiClient,
          receiverName: widget.receiverName,
          receiverPhone: widget.receiverPhone,
          estimatedFareCents: widget.estimatedFareCents,
          currency: widget.currency,
          onDone: _finish,
        );
      case DeliveryStatus.cancelled:
        return _CancelledView(key: key, onDone: _finish);
      case DeliveryStatus.unknown:
        return _StatusView(
          key: key,
          mascot: 'mascot_waiting.png',
          animation: MascotAnimation.none,
          title: 'Đang cập nhật trạng thái',
          subtitle: 'Vui lòng chờ trong giây lát.',
          showRoute: false,
          data: this,
        );
    }
  }
}

class _StatusView extends StatelessWidget {
  const _StatusView({
    super.key,
    required this.mascot,
    required this.animation,
    required this.title,
    required this.subtitle,
    required this.showRoute,
    required this.data,
  });

  final String mascot;
  final MascotAnimation animation;
  final String title;
  final String subtitle;
  final bool showRoute;
  final _DeliveryLifecyclePageState data;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        MascotImage(asset: mascot, size: MascotSize.medium, animation: animation, semanticLabel: title),
        const SizedBox(height: AppSpacing.md),
        Text(title, style: theme.textTheme.titleLarge, textAlign: TextAlign.center),
        const SizedBox(height: AppSpacing.xs),
        Text(subtitle, style: theme.textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary), textAlign: TextAlign.center),
        const SizedBox(height: AppSpacing.lg),
        if (showRoute) ...[
          DeliveryMapView(pickup: data.widget.pickupLocation, receiver: data.widget.receiverLocation),
          const SizedBox(height: AppSpacing.md),
        ],
        AppCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _AddressLine(icon: Icons.circle, color: AppColors.primary, address: data.widget.pickupAddress),
              const SizedBox(height: AppSpacing.sm),
              _AddressLine(icon: Icons.location_on, color: AppColors.error, address: data.widget.receiverAddress),
            ],
          ),
        ),
      ],
    );
  }
}

class _AddressLine extends StatelessWidget {
  const _AddressLine({required this.icon, required this.color, required this.address});

  final IconData icon;
  final Color color;
  final String address;

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, size: 14, color: color),
        const SizedBox(width: AppSpacing.sm),
        Expanded(child: Text(address, style: Theme.of(context).textTheme.bodyMedium)),
      ],
    );
  }
}

class _DeliveredView extends StatelessWidget {
  const _DeliveredView({
    super.key,
    required this.tripId,
    required this.apiClient,
    required this.onDone,
    this.receiverName,
    this.receiverPhone,
    this.estimatedFareCents,
    this.currency,
  });

  final String tripId;
  final ApiClient apiClient;
  final VoidCallback onDone;
  final String? receiverName;
  final String? receiverPhone;
  final int? estimatedFareCents;
  final String? currency;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        const MascotImage(
          asset: 'mascot_celebration.png',
          size: MascotSize.large,
          animation: MascotAnimation.bounce,
          semanticLabel: 'Giao hàng thành công',
        ),
        const SizedBox(height: AppSpacing.md),
        Text('Giao hàng thành công!',
            style: theme.textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.bold, color: AppColors.primary)),
        const SizedBox(height: AppSpacing.xl),
        AppButton.outline(
          label: 'Xem hóa đơn',
          icon: Icons.receipt_long_outlined,
          onPressed: () => DeliveryReceiptSheet.show(
            context,
            tripId: tripId,
            apiClient: apiClient,
            receiverName: receiverName,
            receiverPhone: receiverPhone,
            estimatedFareCents: estimatedFareCents,
            currency: currency,
          ),
        ),
        const SizedBox(height: AppSpacing.sm),
        AppButton.primary(label: 'Xong', onPressed: onDone),
      ],
    );
  }
}

class _CancelledView extends StatelessWidget {
  const _CancelledView({super.key, required this.onDone});

  final VoidCallback onDone;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        const MascotImage(
          asset: 'mascot_no_connection.png',
          size: MascotSize.medium,
          animation: MascotAnimation.fade,
          semanticLabel: 'Đơn giao hàng đã bị hủy',
        ),
        const SizedBox(height: AppSpacing.md),
        Text('Đơn giao hàng đã bị hủy', style: theme.textTheme.titleLarge, textAlign: TextAlign.center),
        const SizedBox(height: AppSpacing.xl),
        AppButton.primary(label: 'Xong', onPressed: onDone),
      ],
    );
  }
}
