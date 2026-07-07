import 'package:google_maps_flutter/google_maps_flutter.dart';

List<LatLng> decodePolyline(String encoded) {
  final points = <LatLng>[];
  int index = 0;
  final len = encoded.length;
  int lat = 0;
  int lng = 0;

  while (index < len) {
    int shift = 0;
    int result = 0;
    int b;
    do {
      b = encoded.codeUnitAt(index++) - 63;
      result |= (b & 0x1f) << shift;
      shift += 5;
    } while (b >= 0x20);
    final deltaLat = (result & 1) != 0 ? ~(result >> 1) : (result >> 1);
    lat += deltaLat;

    shift = 0;
    result = 0;
    do {
      b = encoded.codeUnitAt(index++) - 63;
      result |= (b & 0x1f) << shift;
      shift += 5;
    } while (b >= 0x20);
    final deltaLng = (result & 1) != 0 ? ~(result >> 1) : (result >> 1);
    lng += deltaLng;

    points.add(LatLng(lat / 1e5, lng / 1e5));
  }
  return points;
}
