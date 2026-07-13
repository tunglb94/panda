/// Contact Card data for the trip's other participant (the rider, from a
/// driver's point of view). [maskedPhone] is for display only — never the
/// real number (see `ContactRepository.call`, which is the only way to
/// obtain a real number, and only for immediately opening the dialer).
class ContactInfo {
  const ContactInfo({
    required this.name,
    required this.maskedPhone,
    required this.rating,
    required this.ratingCount,
    this.vehicleType = '',
    this.vehicleBrand = '',
    this.vehicleModel = '',
    this.plateNumber = '',
  });

  final String name;
  final String maskedPhone;
  final double rating;
  final int ratingCount;

  /// Empty for a rider (only the driver side of a trip has vehicle info).
  final String vehicleType;
  final String vehicleBrand;
  final String vehicleModel;
  final String plateNumber;

  bool get hasRating => ratingCount > 0;

  factory ContactInfo.fromJson(Map<String, dynamic> json) => ContactInfo(
        name: json['name'] as String? ?? '',
        maskedPhone: json['masked_phone'] as String? ?? '',
        rating: (json['rating'] as num?)?.toDouble() ?? 0,
        ratingCount: (json['rating_count'] as num?)?.toInt() ?? 0,
        vehicleType: json['vehicle_type'] as String? ?? '',
        vehicleBrand: json['vehicle_brand'] as String? ?? '',
        vehicleModel: json['vehicle_model'] as String? ?? '',
        plateNumber: json['plate_number'] as String? ?? '',
      );
}
