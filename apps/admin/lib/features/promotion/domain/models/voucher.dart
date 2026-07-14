/// Admin-facing Voucher shape — mirrors gateway's `adminVoucherJSON`
/// (backend/services/gateway/http/handlers/admin_promotion_handler.go).
class Voucher {
  const Voucher({
    required this.id,
    required this.code,
    required this.title,
    required this.description,
    required this.type,
    required this.value,
    required this.maxDiscount,
    required this.minOrder,
    required this.start,
    required this.end,
    required this.usageLimit,
    required this.perUserLimit,
    required this.budget,
    required this.remainingBudget,
    required this.serviceType,
    required this.tripType,
    required this.campaign,
    required this.enabled,
    required this.state,
    required this.usageCount,
    this.statsIssued,
    this.statsRedeemed,
    this.statsRemaining,
    this.statsExpired,
  });

  factory Voucher.fromJson(Map<String, dynamic> json) {
    final stats = json['stats'] as Map<String, dynamic>?;
    return Voucher(
      id: json['id'] as String? ?? '',
      code: json['code'] as String? ?? '',
      title: json['title'] as String? ?? '',
      description: json['description'] as String? ?? '',
      type: json['type'] as String? ?? 'percentage',
      value: (json['value'] as num?)?.toInt() ?? 0,
      maxDiscount: (json['max_discount'] as num?)?.toInt() ?? 0,
      minOrder: (json['min_order'] as num?)?.toInt() ?? 0,
      start: json['start'] as String? ?? '',
      end: json['end'] as String? ?? '',
      usageLimit: (json['usage_limit'] as num?)?.toInt() ?? 0,
      perUserLimit: (json['per_user_limit'] as num?)?.toInt() ?? 0,
      budget: (json['budget'] as num?)?.toInt() ?? 0,
      remainingBudget: (json['remaining_budget'] as num?)?.toInt() ?? 0,
      serviceType: (json['service_type'] as List<dynamic>? ?? []).cast<String>(),
      tripType: (json['trip_type'] as List<dynamic>? ?? []).cast<String>(),
      campaign: json['campaign'] as String? ?? '',
      enabled: json['enabled'] as bool? ?? false,
      state: json['state'] as String? ?? 'disabled',
      usageCount: (json['usage_count'] as num?)?.toInt() ?? 0,
      statsIssued: (stats?['issued'] as num?)?.toInt(),
      statsRedeemed: (stats?['redeemed'] as num?)?.toInt(),
      statsRemaining: (stats?['remaining'] as num?)?.toInt(),
      statsExpired: (stats?['expired'] as num?)?.toInt(),
    );
  }

  final String id;
  final String code;
  final String title;
  final String description;
  final String type; // "fixed" | "percentage"
  final int value;
  final int maxDiscount;
  final int minOrder;
  final String start; // RFC3339
  final String end; // RFC3339
  final int usageLimit;
  final int perUserLimit;
  final int budget;
  final int remainingBudget;
  final List<String> serviceType;
  final List<String> tripType;
  final String campaign;
  final bool enabled;
  final String state; // active | expired | disabled | exhausted
  final int usageCount;

  // Phase 4 admin stats — null when the backend omits "stats" (e.g. the
  // stats lookup failed server-side; see adminVoucherJSON's doc comment).
  final int? statsIssued;
  final int? statsRedeemed;
  final int? statsRemaining;
  final int? statsExpired;
}
