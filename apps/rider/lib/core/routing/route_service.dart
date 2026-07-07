import 'package:google_maps_flutter/google_maps_flutter.dart';
import 'route_model.dart';

abstract interface class RouteService {
  Future<RouteModel> getRoute(LatLng origin, LatLng destination);
}
