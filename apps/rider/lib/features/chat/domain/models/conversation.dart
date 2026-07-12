class Conversation {
  const Conversation({
    required this.id,
    required this.tripId,
    required this.riderId,
    required this.driverId,
    required this.tripType,
    required this.status,
    required this.unreadCount,
  });

  final String id;
  final String tripId;
  final String riderId;
  final String driverId;

  /// "ride" or "delivery" — carried for display only.
  final String tripType;

  /// "open" or "closed". A closed conversation's input is disabled
  /// client-side (Part 7 — the trip already ended).
  final String status;
  final int unreadCount;

  bool get isOpen => status == 'open';

  factory Conversation.fromJson(Map<String, dynamic> json) => Conversation(
        id: json['id'] as String? ?? '',
        tripId: json['trip_id'] as String? ?? '',
        riderId: json['rider_id'] as String? ?? '',
        driverId: json['driver_id'] as String? ?? '',
        tripType: json['trip_type'] as String? ?? 'ride',
        status: json['status'] as String? ?? 'open',
        unreadCount: (json['unread_count'] as num?)?.toInt() ?? 0,
      );
}
