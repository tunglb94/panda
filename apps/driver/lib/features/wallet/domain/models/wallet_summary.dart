/// Mirrors the backend's wallet Wallet Projection (Settlement Engine spec,
/// Phần 3) — every field is derived from the ledger server-side; this model
/// only carries the numbers, never computes them client-side.
class WalletSummary {
  const WalletSummary({
    required this.currency,
    required this.availableCents,
    required this.pendingCents,
    required this.outstandingCents,
    required this.netCents,
    required this.lifetimeEarnedCents,
    required this.lifetimeWithdrawnCents,
  });

  final String currency;

  /// Withdrawable now — electronically-collected income minus withdrawals/penalties.
  final int availableCents;

  /// The driver's own in-flight (Pending/Approved) payout request amount.
  final int pendingCents;

  /// Commission owed from cash-collected trips (Phần 4).
  final int outstandingCents;

  /// Available - Outstanding — what actually gates a new payout (Phần 5).
  final int netCents;

  final int lifetimeEarnedCents;
  final int lifetimeWithdrawnCents;

  static const empty = WalletSummary(
    currency: 'VND',
    availableCents: 0,
    pendingCents: 0,
    outstandingCents: 0,
    netCents: 0,
    lifetimeEarnedCents: 0,
    lifetimeWithdrawnCents: 0,
  );

  factory WalletSummary.fromJson(Map<String, dynamic> json) => WalletSummary(
        currency: json['currency'] as String? ?? 'VND',
        availableCents: (json['available_cents'] as num?)?.toInt() ?? 0,
        pendingCents: (json['pending_cents'] as num?)?.toInt() ?? 0,
        outstandingCents: (json['outstanding_cents'] as num?)?.toInt() ?? 0,
        netCents: (json['net_cents'] as num?)?.toInt() ?? 0,
        lifetimeEarnedCents: (json['lifetime_earned_cents'] as num?)?.toInt() ?? 0,
        lifetimeWithdrawnCents: (json['lifetime_withdrawn_cents'] as num?)?.toInt() ?? 0,
      );

  Map<String, dynamic> toJson() => {
        'currency': currency,
        'available_cents': availableCents,
        'pending_cents': pendingCents,
        'outstanding_cents': outstandingCents,
        'net_cents': netCents,
        'lifetime_earned_cents': lifetimeEarnedCents,
        'lifetime_withdrawn_cents': lifetimeWithdrawnCents,
      };
}
