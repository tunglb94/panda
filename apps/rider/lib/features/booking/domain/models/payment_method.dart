import 'package:flutter/material.dart';

enum PaymentMethodType { cash, wallet, card }

/// A selectable payment method shown in the Payment Method Card.
///
/// Mock data only — the Wallet and Payment backend services do not exist
/// yet (see `docs/project/MVP_DEVELOPMENT_PLAN.md` §2.1). No charge is ever
/// made from this UI.
class PaymentMethod {
  const PaymentMethod({
    required this.type,
    required this.label,
    required this.subtitle,
    required this.icon,
  });

  final PaymentMethodType type;
  final String label;
  final String subtitle;
  final IconData icon;
}
