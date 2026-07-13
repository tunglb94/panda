// Widget tests for the Driver KYC + Vehicle Verification status page
// (Phần 6 — Status UI). Uses ApiClient's injectable `http.Client` (mirrors
// the Communication Module's tests) so these run against a `MockClient` —
// no real backend needed.
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

import 'package:driver/core/auth/auth_state.dart';
import 'package:driver/core/network/api_client.dart';
import 'package:driver/features/kyc/presentation/pages/kyc_status_page.dart';

http.Response _json(Map<String, dynamic> body, {int status = 200}) => http.Response(
      jsonEncode(body),
      status,
      headers: {'content-type': 'application/json'},
    );

ApiClient _mockApiClient(Future<http.Response> Function(http.Request) handler) => ApiClient(
      baseUrl: 'http://test.local',
      authState: AuthState(),
      httpClient: MockClient((req) => handler(req)),
    );

Map<String, dynamic> _driverVerification({String status = 'pending', String rejectReason = ''}) => {
      'full_name': 'Nguyễn Văn A',
      'date_of_birth': '1995-01-01',
      'address': '123 Lê Lợi, Q1, TP.HCM',
      'national_id_number': '079095001234',
      'license_number': '',
      'status': status,
      'reject_reason': rejectReason,
    };

Map<String, dynamic> _vehicleVerification({String status = 'pending', String rejectReason = ''}) => {
      'vehicle_type': 'motorcycle',
      'service_type': 'bike',
      'brand': 'Honda',
      'model': 'Wave',
      'year': 2020,
      'color': 'Đỏ',
      'plate_number': '59-X1 123.45',
      'license_class': 'A1',
      'ride_enabled': true,
      'delivery_enabled': false,
      'status': status,
      'reject_reason': rejectReason,
    };

void main() {
  testWidgets('KYCStatusPage shows the not-submitted-yet empty state', (tester) async {
    final client = _mockApiClient((req) async => _json({'error': 'not found'}, status: 404));

    await tester.pumpWidget(MaterialApp(home: KYCStatusPage(apiClient: client)));
    await tester.pump();
    await tester.pump();

    expect(find.text('Bạn chưa gửi hồ sơ xác minh'), findsOneWidget);
    expect(find.text('Bắt đầu xác minh'), findsOneWidget);
  });

  testWidgets('KYCStatusPage shows Pending status for both sections', (tester) async {
    final client = _mockApiClient((req) async {
      if (req.url.path.contains('/driver/verification')) {
        return _json(_driverVerification(status: 'pending'));
      }
      return _json(_vehicleVerification(status: 'pending'));
    });

    await tester.pumpWidget(MaterialApp(home: KYCStatusPage(apiClient: client)));
    await tester.pump();
    await tester.pump();

    expect(find.text('Hồ sơ cá nhân (KYC)'), findsOneWidget);
    expect(find.text('Đăng ký xe'), findsOneWidget);
    expect(find.text('Đang chờ duyệt'), findsNWidgets(2));
    expect(find.text('Tiếp tục hồ sơ'), findsNothing);
  });

  testWidgets('KYCStatusPage shows Approved status with no edit CTA', (tester) async {
    final client = _mockApiClient((req) async {
      if (req.url.path.contains('/driver/verification')) {
        return _json(_driverVerification(status: 'approved'));
      }
      return _json(_vehicleVerification(status: 'approved'));
    });

    await tester.pumpWidget(MaterialApp(home: KYCStatusPage(apiClient: client)));
    await tester.pump();
    await tester.pump();

    expect(find.text('Đã xác minh'), findsNWidgets(2));
    expect(find.text('Chỉnh sửa & gửi lại'), findsNothing);
  });

  testWidgets('KYCStatusPage shows Rejected status with the reject reason', (tester) async {
    final client = _mockApiClient((req) async {
      if (req.url.path.contains('/driver/verification')) {
        return _json(_driverVerification(status: 'rejected', rejectReason: 'Ảnh CCCD bị mờ'));
      }
      return _json(_vehicleVerification(status: 'approved'));
    });

    await tester.pumpWidget(MaterialApp(home: KYCStatusPage(apiClient: client)));
    await tester.pump();
    await tester.pump();

    expect(find.text('Bị từ chối'), findsOneWidget);
    expect(find.textContaining('Ảnh CCCD bị mờ'), findsOneWidget);
    expect(find.text('Chỉnh sửa & gửi lại'), findsOneWidget);
  });

  // ─── Phần 11 — Documents section: version + expiry banners ────────────────

  testWidgets('KYCStatusPage shows a red expired banner for an expired document', (tester) async {
    final client = _mockApiClient((req) async {
      final path = req.url.path;
      if (path.contains('/verification/documents')) {
        return _json({
          'documents': [
            {
              'document_type': 'vehicle_registration',
              'uploaded': true,
              'uploaded_at': '2025-01-01T00:00:00Z',
              'version': 2,
              'expires_at': '2025-01-01',
              'expired': true,
              'expiring_soon': false,
            },
          ],
        });
      }
      if (path == '/api/v1/driver/verification') {
        return _json(_driverVerification(status: 'expired', rejectReason: 'Đăng ký xe đã hết hạn'));
      }
      return _json(_vehicleVerification(status: 'expired', rejectReason: 'Đăng ký xe đã hết hạn'));
    });

    await tester.pumpWidget(MaterialApp(home: KYCStatusPage(apiClient: client)));
    await tester.pump();
    await tester.pump();

    expect(find.text('Đăng ký xe'), findsWidgets);
    expect(find.textContaining('Đã hết hạn: '), findsOneWidget);
    expect(find.text('v2'), findsOneWidget);
  });

  testWidgets('KYCStatusPage shows a yellow expiring-soon banner within 30 days', (tester) async {
    final client = _mockApiClient((req) async {
      final path = req.url.path;
      if (path.contains('/verification/documents')) {
        return _json({
          'documents': [
            {
              'document_type': 'vehicle_insurance',
              'uploaded': true,
              'uploaded_at': '2026-01-01T00:00:00Z',
              'version': 1,
              'expires_at': '2026-01-20',
              'expired': false,
              'expiring_soon': true,
            },
          ],
        });
      }
      if (path == '/api/v1/driver/verification') {
        return _json(_driverVerification(status: 'approved'));
      }
      return _json(_vehicleVerification(status: 'approved'));
    });

    await tester.pumpWidget(MaterialApp(home: KYCStatusPage(apiClient: client)));
    await tester.pump();
    await tester.pump();

    expect(find.textContaining('Sắp hết hạn'), findsOneWidget);
  });

  testWidgets('KYCStatusPage shows the progress summary with approved count', (tester) async {
    final client = _mockApiClient((req) async {
      final path = req.url.path;
      if (path.contains('/verification/documents')) {
        return _json({'documents': []});
      }
      if (path == '/api/v1/driver/verification') {
        return _json(_driverVerification(status: 'approved'));
      }
      return _json(_vehicleVerification(status: 'pending'));
    });

    await tester.pumpWidget(MaterialApp(home: KYCStatusPage(apiClient: client)));
    await tester.pump();
    await tester.pump();

    expect(find.text('1/2 đã duyệt'), findsOneWidget);
  });
}
