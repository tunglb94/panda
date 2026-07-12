import 'package:flutter/material.dart';

/// Item type the sender picks in the Delivery form. This is a UI-only
/// concept — `CreateTripRequest`/`BookRideRequest` has no `package_type`/
/// category field on the wire (only `package_note`, free text), so the
/// selected category is folded into the note the driver actually sees via
/// [DeliveryForm.notePrefix] rather than sent as a fake structured field
/// that doesn't exist server-side. Labels mirror
/// `backend/services/trip/domain/entity/delivery.go`'s `PackageType`
/// enum (DOCUMENT/SMALL/MEDIUM/LARGE) for consistency, even though that
/// field is currently hardcoded server-side and never read from the
/// request.
enum PackageCategory {
  document,
  small,
  medium,
  large;

  String get label => switch (this) {
        PackageCategory.document => 'Tài liệu',
        PackageCategory.small => 'Hàng nhỏ',
        PackageCategory.medium => 'Hàng vừa',
        PackageCategory.large => 'Hàng lớn',
      };

  IconData get icon => switch (this) {
        PackageCategory.document => Icons.description_outlined,
        PackageCategory.small => Icons.inventory_2_outlined,
        PackageCategory.medium => Icons.shopping_bag_outlined,
        PackageCategory.large => Icons.local_shipping_outlined,
      };
}
