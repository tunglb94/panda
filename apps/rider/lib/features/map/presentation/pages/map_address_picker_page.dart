import 'package:flutter/material.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';

import 'package:rider/core/places/nominatim_places_service.dart';
import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/shared/widgets/app_button.dart';

/// Standalone "pick a location by dragging the map" page — the map/pin half
/// of [MapPage]'s pickup/destination selection, extracted so any flow that
/// needs an address (currently: `DeliveryFormPage`) can reuse it via
/// `Navigator.push`/`pop` instead of re-implementing map+pin logic.
///
/// Returns the confirmed `(address, LatLng)` via [Navigator.pop], or `null`
/// if the rider backs out without confirming.
class MapAddressPickerPage extends StatefulWidget {
  const MapAddressPickerPage({super.key, this.initialPosition, this.title = 'Chọn vị trí trên bản đồ'});

  /// Where the map centers on open. Falls back to a Ho Chi Minh City center
  /// point (the same sample coordinate `MockBookingCatalog.sampleTripSelection`
  /// already uses elsewhere) when the caller has no known position yet —
  /// e.g. opened from the Home Hub rather than from `MapPage`, which always
  /// has a resolved GPS fix by the time it's on screen.
  final LatLng? initialPosition;
  final String title;

  @override
  State<MapAddressPickerPage> createState() => _MapAddressPickerPageState();
}

class _MapAddressPickerPageState extends State<MapAddressPickerPage> {
  static const _placesService = NominatimPlacesService();
  static const _fallbackPosition = LatLng(10.7769, 106.7009);
  static const _defaultZoom = 16.0;

  late LatLng _center = widget.initialPosition ?? _fallbackPosition;
  String? _resolvedAddress;
  bool _resolving = false;

  @override
  void initState() {
    super.initState();
    _reverseGeocode(_center);
  }

  Future<void> _reverseGeocode(LatLng position) async {
    setState(() => _resolving = true);
    try {
      final placeId = '${position.latitude},${position.longitude}';
      final details = await _placesService.details(placeId);
      if (!mounted) return;
      setState(() {
        _resolvedAddress = details.formattedAddress.isNotEmpty ? details.formattedAddress : null;
        _resolving = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _resolvedAddress = null;
        _resolving = false;
      });
    }
  }

  void _onCameraMove(CameraPosition position) {
    _center = position.target;
  }

  void _onCameraIdle() {
    _reverseGeocode(_center);
  }

  void _confirm() {
    final address = _resolvedAddress ??
        '${_center.latitude.toStringAsFixed(5)}, ${_center.longitude.toStringAsFixed(5)}';
    Navigator.of(context).pop((address, _center));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(widget.title)),
      body: Stack(
        alignment: Alignment.center,
        children: [
          GoogleMap(
            initialCameraPosition: CameraPosition(target: _center, zoom: _defaultZoom),
            onCameraMove: _onCameraMove,
            onCameraIdle: _onCameraIdle,
            myLocationButtonEnabled: false,
            zoomControlsEnabled: false,
          ),
          const Padding(
            padding: EdgeInsets.only(bottom: 36),
            child: Icon(Icons.location_pin, size: 44, color: AppColors.primary),
          ),
          Positioned(
            left: AppSpacing.lg,
            right: AppSpacing.lg,
            bottom: AppSpacing.lg,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(AppSpacing.md),
                  decoration: BoxDecoration(
                    color: AppColors.surface,
                    borderRadius: BorderRadius.circular(12),
                    boxShadow: const [BoxShadow(color: Colors.black12, blurRadius: 8)],
                  ),
                  child: Text(
                    _resolving ? 'Đang xác định địa chỉ…' : (_resolvedAddress ?? 'Không xác định được địa chỉ — vẫn có thể dùng toạ độ'),
                    style: Theme.of(context).textTheme.bodyMedium,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                const SizedBox(height: AppSpacing.sm),
                SizedBox(
                  width: double.infinity,
                  child: AppButton.primary(label: 'Xác nhận vị trí này', onPressed: _confirm),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
