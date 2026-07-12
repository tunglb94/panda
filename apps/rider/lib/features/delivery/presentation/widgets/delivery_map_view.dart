import 'package:flutter/material.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';

import 'package:rider/core/theme/app_radius.dart';

/// Pickup + receiver markers and a straight polyline between them — a
/// self-contained widget, entirely separate from `MapPage`'s ride
/// pickup/destination map (which itself draws no markers/polyline of its
/// own today; see the Delivery audit). Building a new widget here rather
/// than extending `MapPage` keeps Ride's map untouched.
///
/// The polyline is a straight line between the two points, not a routed
/// path — this app has no routing/directions backend (see
/// `MockTripMetrics`'s doc comment on the same limitation for distance/
/// duration), so drawing a fake routed polyline would misrepresent the
/// actual road path. A straight line honestly shows only what's known:
/// the two endpoints.
class DeliveryMapView extends StatelessWidget {
  const DeliveryMapView({
    super.key,
    required this.pickup,
    required this.receiver,
    this.height = 200,
  });

  final LatLng pickup;
  final LatLng receiver;
  final double height;

  @override
  Widget build(BuildContext context) {
    final bounds = LatLngBounds(
      southwest: LatLng(
        pickup.latitude < receiver.latitude ? pickup.latitude : receiver.latitude,
        pickup.longitude < receiver.longitude ? pickup.longitude : receiver.longitude,
      ),
      northeast: LatLng(
        pickup.latitude > receiver.latitude ? pickup.latitude : receiver.latitude,
        pickup.longitude > receiver.longitude ? pickup.longitude : receiver.longitude,
      ),
    );
    final center = LatLng(
      (pickup.latitude + receiver.latitude) / 2,
      (pickup.longitude + receiver.longitude) / 2,
    );

    return ClipRRect(
      borderRadius: AppRadius.mdAll,
      child: SizedBox(
        height: height,
        child: GoogleMap(
          initialCameraPosition: CameraPosition(target: center, zoom: 13),
          onMapCreated: (controller) {
            Future.delayed(const Duration(milliseconds: 300), () {
              controller.animateCamera(CameraUpdate.newLatLngBounds(bounds, 48));
            });
          },
          markers: {
            Marker(
              markerId: const MarkerId('pickup'),
              position: pickup,
              icon: BitmapDescriptor.defaultMarkerWithHue(BitmapDescriptor.hueGreen),
              infoWindow: const InfoWindow(title: 'Điểm lấy hàng'),
            ),
            Marker(
              markerId: const MarkerId('receiver'),
              position: receiver,
              icon: BitmapDescriptor.defaultMarkerWithHue(BitmapDescriptor.hueRed),
              infoWindow: const InfoWindow(title: 'Điểm giao hàng'),
            ),
          },
          polylines: {
            Polyline(
              polylineId: const PolylineId('delivery_route'),
              points: [pickup, receiver],
              color: const Color(0xFF1A8C4E),
              width: 3,
              patterns: [PatternItem.dash(12), PatternItem.gap(8)],
            ),
          },
          zoomControlsEnabled: false,
          myLocationButtonEnabled: false,
        ),
      ),
    );
  }
}
