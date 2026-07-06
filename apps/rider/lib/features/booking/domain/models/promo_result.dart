/// Result of validating a promo code against the mock promo catalog.
///
/// There is no Promotion backend yet (see
/// `docs/project/MVP_DEVELOPMENT_PLAN.md` §2.1) — this is a client-side
/// stand-in so the Promo Code Entry widget has something to demo against.
class PromoResult {
  const PromoResult({
    required this.isValid,
    required this.message,
    this.discountPercent = 0,
  });

  final bool isValid;
  final String message;
  final int discountPercent;

  static const none = PromoResult(isValid: false, message: '');
}

/// Hardcoded acceptance list — purely for UI demo purposes. Real promo
/// validation belongs to the (not yet started) Promotion service.
class MockPromoValidator {
  const MockPromoValidator._();

  static const Map<String, int> _codes = {
    'FAIRRIDE10': 10,
    'WELCOME20': 20,
  };

  static PromoResult validate(String rawCode) {
    final code = rawCode.trim().toUpperCase();
    if (code.isEmpty) {
      return const PromoResult(isValid: false, message: 'Enter a promo code');
    }
    final discount = _codes[code];
    if (discount == null) {
      return const PromoResult(isValid: false, message: 'Invalid or expired code');
    }
    return PromoResult(
      isValid: true,
      message: '$discount% off applied',
      discountPercent: discount,
    );
  }
}
