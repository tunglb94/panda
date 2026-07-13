import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

import 'package:rider/core/auth/auth_state.dart';
import 'package:rider/core/network/api_client.dart';
import 'package:rider/features/booking/domain/models/mock_booking_catalog.dart';
import 'package:rider/features/booking/domain/models/vehicle_option.dart';
import 'package:rider/features/booking/presentation/widgets/vehicle_list_sheet.dart';

http.Response _fareResponse(String serviceType) => http.Response(
      jsonEncode({
        'service_type': serviceType,
        'vehicle_type': serviceType,
        'distance_km': 5.0,
        'duration_minutes': 15.0,
        'base_fare': 8000,
        'distance_fare': 14000,
        'time_fare': 3000,
        'booking_fee': 2000,
        'ride_fare': 25000,
        'total': 27000,
        'currency_code': 'VND',
        'is_final': false,
      }),
      200,
      headers: {'content-type': 'application/json'},
    );

ApiClient _testApiClient() => ApiClient(
      baseUrl: 'http://test.local',
      authState: AuthState(),
      httpClient: MockClient((req) async {
        final body = jsonDecode(req.body) as Map<String, dynamic>;
        return _fareResponse(body['service_type'] as String);
      }),
    );

void main() {
  testWidgets('VehicleListSheet renders every catalog entry with its real vehicle artwork and a real quote', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Builder(
          builder: (context) => ElevatedButton(
            onPressed: () => VehicleListSheet.show(
              context,
              options: MockBookingCatalog.vehicles,
              selected: VehicleCategory.motorcycle,
              pickup: const LatLng(10.7769, 106.7009),
              destination: const LatLng(10.8231, 106.6297),
              tripType: 'ride',
              apiClient: _testApiClient(),
            ),
            child: const Text('open'),
          ),
        ),
      ),
    );

    await tester.tap(find.text('open'));
    await tester.pumpAndSettle();

    expect(find.text('Bike'), findsOneWidget);
    expect(find.text('Bike Plus'), findsOneWidget);
    expect(find.text('Car'), findsOneWidget);
    expect(find.text('Car XL'), findsOneWidget);

    for (final vehicle in MockBookingCatalog.vehicles) {
      expect(vehicle.imageAsset, isNotNull, reason: '${vehicle.label} should have real vehicle artwork');
    }
    expect(find.byType(Image), findsNWidgets(MockBookingCatalog.vehicles.length));

    // Every row's price comes from the (mocked) backend response, not a
    // client-side computation — all 4 rows quote the same stubbed total.
    expect(find.text('27.000 đ'), findsNWidgets(MockBookingCatalog.vehicles.length));
  });
}
