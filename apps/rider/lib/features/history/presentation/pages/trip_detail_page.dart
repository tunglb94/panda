import 'package:flutter/material.dart';

import 'package:rider/features/booking/presentation/widgets/fare_summary_card.dart';
import 'package:rider/features/booking/presentation/widgets/trip_point_cards.dart';
import 'package:rider/features/trip/presentation/widgets/driver_info_card.dart';

import '../../domain/models/trip_history_entry.dart';
import '../widgets/distance_duration_card.dart';
import '../widgets/payment_method_row.dart';
import '../widgets/status_chip.dart';
import '../widgets/trip_timeline.dart';
import '../widgets/vehicle_info_card.dart';
import 'receipt_page.dart';

/// Trip Detail: route summary, driver/vehicle info, timeline, fare
/// breakdown (with promo discount), payment method, duration and distance.
///
/// Reuses `PickupCard`/`DestinationCard`/`FareSummaryCard` (Booking module)
/// and `DriverInfoCard` (Trip module) directly, per this phase's "reuse
/// existing components" requirement.
class TripDetailPage extends StatelessWidget {
  const TripDetailPage({super.key, required this.entry});

  final TripHistoryEntry entry;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Trip Details')),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        'Route summary',
                        style: Theme.of(context)
                            .textTheme
                            .titleMedium
                            ?.copyWith(fontWeight: FontWeight.w600),
                      ),
                      StatusChip(status: entry.status),
                    ],
                  ),
                  const SizedBox(height: 12),
                  PickupCard(
                    address: entry.route.pickupAddress,
                    coordinate: entry.route.pickup,
                  ),
                  const SizedBox(height: 8),
                  DestinationCard(
                    address: entry.route.destinationAddress,
                    coordinate: entry.route.destination,
                  ),
                  const SizedBox(height: 20),
                  Text(
                    'Driver',
                    style: Theme.of(context)
                        .textTheme
                        .titleSmall
                        ?.copyWith(fontWeight: FontWeight.w600),
                  ),
                  const SizedBox(height: 8),
                  DriverInfoCard(driver: entry.driver),
                  const SizedBox(height: 16),
                  Text(
                    'Vehicle',
                    style: Theme.of(context)
                        .textTheme
                        .titleSmall
                        ?.copyWith(fontWeight: FontWeight.w600),
                  ),
                  const SizedBox(height: 8),
                  VehicleInfoCard(category: entry.vehicleCategory, driver: entry.driver),
                  const SizedBox(height: 20),
                  Text(
                    'Timeline',
                    style: Theme.of(context)
                        .textTheme
                        .titleSmall
                        ?.copyWith(fontWeight: FontWeight.w600),
                  ),
                  const SizedBox(height: 12),
                  TripTimeline(events: entry.timeline),
                  const SizedBox(height: 20),
                  DistanceDurationCard(
                    distanceKm: entry.distanceKm,
                    durationMin: entry.durationMin,
                  ),
                  const SizedBox(height: 16),
                  FareSummaryCard(breakdown: entry.fare),
                  const SizedBox(height: 16),
                  Text(
                    'Payment method',
                    style: Theme.of(context)
                        .textTheme
                        .titleSmall
                        ?.copyWith(fontWeight: FontWeight.w600),
                  ),
                  const SizedBox(height: 8),
                  PaymentMethodRow(method: entry.paymentMethod),
                  const SizedBox(height: 24),
                  SizedBox(
                    width: double.infinity,
                    child: OutlinedButton.icon(
                      icon: const Icon(Icons.receipt_long_outlined),
                      label: const Text('View Receipt'),
                      onPressed: () => Navigator.of(context).push(
                        MaterialPageRoute(builder: (_) => ReceiptPage(entry: entry)),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
