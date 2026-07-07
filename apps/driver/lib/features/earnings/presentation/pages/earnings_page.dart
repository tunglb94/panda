import 'package:flutter/material.dart';

import 'package:driver/shared/widgets/placeholder_tab_content.dart';

/// Earnings tab placeholder. Will eventually show the earnings dashboard
/// and payout balance (Driver App Roadmap stage D7 — blocked on the Wallet
/// backend, which does not exist yet).
class EarningsPage extends StatelessWidget {
  const EarningsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Earnings')),
      body: const PlaceholderTabContent(
        icon: Icons.account_balance_wallet_outlined,
        title: 'Earnings',
        subtitle: 'Your earnings dashboard and payout balance will appear '
            'here in a future phase.',
      ),
    );
  }
}
