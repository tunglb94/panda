import 'package:flutter/material.dart';

/// Colored vehicle/service badge (Phần 7 — "Badge màu khác nhau cho
/// Bike/Bike Plus/Car/Car XL"). [serviceType] is the backend's
/// `ServiceType` wire value (motorcycle/bike_plus/car/car_xl).
class VehicleBadge extends StatelessWidget {
  const VehicleBadge({super.key, required this.serviceType});

  final String? serviceType;

  static const _labels = {
    'motorcycle': 'Bike',
    'bike_plus': 'Bike Plus',
    'car': 'Car',
    'car_xl': 'Car XL',
  };

  static const _colors = {
    'motorcycle': Color(0xFF0D9488),
    'bike_plus': Color(0xFF4F46E5),
    'car': Color(0xFF2563EB),
    'car_xl': Color(0xFF7C3AED),
  };

  @override
  Widget build(BuildContext context) {
    final type = serviceType;
    if (type == null || type.isEmpty) {
      return const Text('—', style: TextStyle(color: Color(0xFF9CA3AF)));
    }
    final color = _colors[type] ?? const Color(0xFF6B7280);
    final label = _labels[type] ?? type;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(color: color.withValues(alpha: 0.12), borderRadius: BorderRadius.circular(6)),
      child: Text(label, style: TextStyle(color: color, fontSize: 12, fontWeight: FontWeight.w600)),
    );
  }
}
