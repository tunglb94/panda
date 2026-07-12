/// The reason a promotion applies to this trip — mirrors the promotion
/// types `backend/services/promotion`'s `PromotionType` already models
/// (First Ride, Birthday, Rain, Airport, Membership, ...), kept as a
/// separate, small UI-only enum since the Promotion Engine has no API
/// surface for the app to consume yet (see `voucher_catalog.dart`).
enum PromotionKind { firstRide, birthday, rain, airport, membership, referral, weekend, event }

extension PromotionKindPresentation on PromotionKind {
  String get emoji => switch (this) {
        PromotionKind.firstRide => '🎉',
        PromotionKind.birthday => '🎂',
        PromotionKind.rain => '🌧️',
        PromotionKind.airport => '🛫',
        PromotionKind.membership => '⭐',
        PromotionKind.referral => '🤝',
        PromotionKind.weekend => '📅',
        PromotionKind.event => '🎊',
      };
}

/// A promotion applicable to the current trip, with a rider-facing reason
/// (BRB §1.2 Transparency Before Conversion: "Promotions always show their
/// exact discount" — the [reason] string is what satisfies that here).
///
/// No backend source exists today — see [PromotionBanner]'s doc comment.
/// This model exists so the banner is fully built and ready to render the
/// moment a real promotion is available.
class PromotionInfo {
  const PromotionInfo({required this.kind, required this.title, required this.reason});

  final PromotionKind kind;
  final String title;

  /// Short, specific explanation of why this promotion applies — never a
  /// generic "Bạn có ưu đãi!" placeholder.
  final String reason;
}
