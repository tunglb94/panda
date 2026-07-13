import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';

import '../../domain/models/vehicle_option.dart';
import 'booking_form_body.dart';

/// Modal presentation of the booking form, invoked from the Map page once
/// pickup + destination are confirmed ("Book this ride").
///
/// Shares its content with `BookingPage` via [BookingFormBody] so both entry
/// points stay in sync with a single implementation.
class BookingBottomSheet {
  const BookingBottomSheet._();

  static Future<void> show(
    BuildContext context, {
    required TripSelection tripSelection,
    required ApiClient apiClient,
    void Function(String driverId)? onDriverAssigned,
    VehicleCategory? initialCategory,
  }) {
    return showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (sheetContext) => DraggableScrollableSheet(
        initialChildSize: 0.75,
        minChildSize: 0.5,
        maxChildSize: 0.95,
        expand: false,
        builder: (context, scrollController) => Material(
          color: Theme.of(context).scaffoldBackgroundColor,
          borderRadius: const BorderRadius.vertical(top: Radius.circular(20)),
          child: SafeArea(
            top: false,
            child: SingleChildScrollView(
              controller: scrollController,
              padding: const EdgeInsets.fromLTRB(20, 12, 20, 24),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Center(
                    child: Container(
                      width: 40,
                      height: 4,
                      margin: const EdgeInsets.only(bottom: 16),
                      decoration: BoxDecoration(
                        color: Colors.grey.shade300,
                        borderRadius: BorderRadius.circular(2),
                      ),
                    ),
                  ),
                  BookingFormBody(
                    tripSelection: tripSelection,
                    apiClient: apiClient,
                    onDriverAssigned: onDriverAssigned,
                    initialCategory: initialCategory,
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
