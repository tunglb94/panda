/// Mirrors the backend's `entity.LicenseClass` matrix
/// (driver/domain/entity/license_class.go) exactly — for client-side
/// display/pre-validation only; the backend is the source of truth and
/// re-validates on submit.
abstract final class LicenseClass {
  static const a1 = 'A1';
  static const a2 = 'A2';
  static const b1 = 'B1';
  static const b2 = 'B2';

  static const all = [a1, a2, b1, b2];
}

String licenseClassLabel(String c) => switch (c) {
      LicenseClass.a1 => 'A1 — Xe máy dưới 175cm³',
      LicenseClass.a2 => 'A2 — Xe máy trên 175cm³',
      LicenseClass.b1 => 'B1 — Ô tô số tự động',
      LicenseClass.b2 => 'B2 — Ô tô số tự động & số sàn',
      _ => c,
    };

const Map<String, List<String>> _licenseClassServiceTypes = {
  LicenseClass.a1: ['motorcycle', 'bike_plus'],
  LicenseClass.a2: ['motorcycle', 'bike_plus'],
  LicenseClass.b1: ['car', 'car_xl'],
  LicenseClass.b2: ['motorcycle', 'bike_plus', 'car', 'car_xl'],
};

/// Phần 2 — license validation matrix. Returns false for an unknown class
/// or service type rather than throwing, so the UI can just disable/hide.
bool licenseAllowsServiceType(String licenseClass, String serviceType) {
  return _licenseClassServiceTypes[licenseClass]?.contains(serviceType) ?? false;
}
