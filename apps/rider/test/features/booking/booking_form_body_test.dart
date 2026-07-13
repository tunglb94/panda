// Regression test: BookingFormBody must show the real backend fare quote
// (POST /api/v1/rides/estimate-fare) — never a client-computed number. On
// success it renders the quoted total; on failure it shows "Không thể tính
// giá" with no fallback estimate.
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

import 'package:rider/core/auth/auth_state.dart';
import 'package:rider/core/network/api_client.dart';
import 'package:rider/features/booking/domain/models/vehicle_option.dart';
import 'package:rider/features/booking/presentation/widgets/booking_form_body.dart';
import 'package:rider/features/booking/presentation/widgets/fare_summary_card.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';

const _pickup = LatLng(10.78778, 106.73614);
const _destination = LatLng(10.8181, 106.6656);

const _trip = TripSelection(pickup: _pickup, destination: _destination);

ApiClient _apiClientReturning(http.Response Function(http.Request) respond) =>
    ApiClient(baseUrl: 'http://test.local', authState: AuthState(), httpClient: MockClient((req) async => respond(req)));

void main() {
  testWidgets('Renders the real backend fare quote from /api/v1/rides/estimate-fare', (tester) async {
    var requestedServiceType = '';
    final apiClient = _apiClientReturning((req) {
      final body = jsonDecode(req.body) as Map<String, dynamic>;
      requestedServiceType = body['service_type'] as String;
      return http.Response(
        jsonEncode({
          'service_type': body['service_type'],
          'vehicle_type': body['service_type'],
          'distance_km': 16.0,
          'duration_minutes': 40.0,
          'base_fare': 15000,
          'distance_fare': 104000,
          'time_fare': 16000,
          'booking_fee': 2000,
          'ride_fare': 135000,
          'total': 137000,
          'currency_code': 'VND',
          'is_final': false,
        }),
        200,
        headers: {'content-type': 'application/json'},
      );
    });

    await tester.pumpWidget(MaterialApp(
      home: Scaffold(
        body: SingleChildScrollView(
          child: BookingFormBody(
            tripSelection: _trip,
            apiClient: apiClient,
            initialCategory: VehicleCategory.car,
          ),
        ),
      ),
    ));
    await tester.pumpAndSettle();

    expect(requestedServiceType, 'car');
    expect(find.textContaining('137.000'), findsWidgets);
  });

  testWidgets('Shows "Không thể tính giá" and no fallback number when the API call fails', (tester) async {
    final apiClient = _apiClientReturning(
      (req) => http.Response(jsonEncode({'error': 'pricing service unavailable'}), 503),
    );

    await tester.pumpWidget(MaterialApp(
      home: Scaffold(
        body: SingleChildScrollView(
          child: BookingFormBody(
            tripSelection: _trip,
            apiClient: apiClient,
            initialCategory: VehicleCategory.car,
          ),
        ),
      ),
    ));
    await tester.pumpAndSettle();

    expect(find.text('Không thể tính giá'), findsOneWidget);
    // No client-computed fare must ever be shown in its place.
    expect(find.byType(FareSummaryCard), findsNothing);
  });

  testWidgets('Shows a loading state while the estimate is in flight', (tester) async {
    final apiClient = _apiClientReturning(
      (req) => http.Response(
        jsonEncode({
          'service_type': 'car',
          'vehicle_type': 'car',
          'distance_km': 5.0,
          'duration_minutes': 15.0,
          'base_fare': 15000,
          'distance_fare': 32500,
          'time_fare': 6000,
          'booking_fee': 2000,
          'ride_fare': 53500,
          'total': 55500,
          'currency_code': 'VND',
          'is_final': false,
        }),
        200,
        headers: {'content-type': 'application/json'},
      ),
    );

    await tester.pumpWidget(MaterialApp(
      home: Scaffold(
        body: SingleChildScrollView(
          child: BookingFormBody(
            tripSelection: _trip,
            apiClient: apiClient,
            initialCategory: VehicleCategory.car,
          ),
        ),
      ),
    ));
    // Before the mocked HTTP response resolves, the loading skeleton (a
    // CircularProgressIndicator) must be visible, not a fare figure.
    await tester.pump();
    expect(find.byType(CircularProgressIndicator), findsWidgets);

    await tester.pumpAndSettle();
    expect(find.textContaining('55.500'), findsWidgets);
  });
}
