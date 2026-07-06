import 'package:flutter/material.dart';

import 'package:rider/features/profile/domain/models/mock_profile_repository.dart';

import '../../domain/models/trip_history_entry.dart';
import '../widgets/receipt_row.dart';

/// Receipt-style summary of a past trip: Trip ID, rider, driver, vehicle,
/// date, payment method, fare breakdown, taxes, and total.
///
/// Reuses `MockFareBreakdown` (Booking module) for the fare numbers and
/// `MockRiderProfileCatalog` (Profile module) for the rider's name, rather
/// than reusing `FareSummaryCard`'s widget itself — a receipt is
/// deliberately styled differently (monospace rows, dashed dividers) from
/// the branded Booking fare card. Tax is a receipt-only concern computed
/// here; it is not part of the shared `MockFareBreakdown` model.
class ReceiptPage extends StatelessWidget {
  const ReceiptPage({super.key, required this.entry});

  final TripHistoryEntry entry;

  /// Flat mock tax rate — purely illustrative, no tax engine exists.
  static const double _taxRate = 0.08;

  @override
  Widget build(BuildContext context) {
    final fare = entry.fare;
    final taxCents = (fare.totalCents * _taxRate).round();
    final grandTotalCents = fare.totalCents + taxCents;
    final rider = MockRiderProfileCatalog.sample;

    return Scaffold(
      appBar: AppBar(title: const Text('Receipt')),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: Colors.grey.shade200),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    const Center(
                      child: Icon(Icons.directions_car_filled, size: 32, color: Color(0xFF1A8C4E)),
                    ),
                    const SizedBox(height: 6),
                    const Center(
                      child: Text(
                        'FAIRRIDE',
                        style: TextStyle(fontWeight: FontWeight.bold, letterSpacing: 2),
                      ),
                    ),
                    const Center(
                      child: Text(
                        'Trip Receipt',
                        style: TextStyle(fontSize: 12, color: Colors.grey),
                      ),
                    ),
                    const SizedBox(height: 16),
                    const ReceiptDivider(),
                    const SizedBox(height: 8),
                    ReceiptRow(label: 'Trip ID', value: entry.id),
                    ReceiptRow(label: 'Date', value: _formatDate(entry.dateTime)),
                    ReceiptRow(label: 'Rider', value: rider.fullName),
                    ReceiptRow(label: 'Driver', value: entry.driver.name),
                    ReceiptRow(
                      label: 'Vehicle',
                      value: '${entry.driver.vehicleModel} (${entry.driver.plateNumber})',
                    ),
                    ReceiptRow(label: 'Payment', value: entry.paymentMethod.label),
                    const SizedBox(height: 12),
                    const ReceiptDivider(),
                    const SizedBox(height: 8),
                    ReceiptRow(label: 'Base fare', value: fare.format(fare.baseFareCents)),
                    ReceiptRow(label: 'Distance', value: fare.format(fare.distanceFareCents)),
                    ReceiptRow(label: 'Time', value: fare.format(fare.timeFareCents)),
                    ReceiptRow(label: 'Booking fee', value: fare.format(fare.bookingFeeCents)),
                    if (fare.discountCents > 0)
                      ReceiptRow(
                        label: 'Promo discount',
                        value: '-${fare.format(fare.discountCents)}',
                      ),
                    ReceiptRow(label: 'Subtotal', value: fare.format(fare.totalCents)),
                    ReceiptRow(
                      label: 'Taxes (${(_taxRate * 100).toStringAsFixed(0)}%)',
                      value: fare.format(taxCents),
                    ),
                    const SizedBox(height: 8),
                    const ReceiptDivider(),
                    const SizedBox(height: 8),
                    ReceiptRow(
                      label: 'Total',
                      value: fare.format(grandTotalCents),
                      bold: true,
                    ),
                    const SizedBox(height: 16),
                    Center(
                      child: Text(
                        'Thank you for riding with FAIRRIDE',
                        style: TextStyle(fontSize: 11, color: Colors.grey.shade500),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  static const _months = [
    'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
    'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec',
  ];

  static String _formatDate(DateTime dt) {
    final hh = dt.hour.toString().padLeft(2, '0');
    final mm = dt.minute.toString().padLeft(2, '0');
    return '${_months[dt.month - 1]} ${dt.day}, ${dt.year} · $hh:$mm';
  }
}
