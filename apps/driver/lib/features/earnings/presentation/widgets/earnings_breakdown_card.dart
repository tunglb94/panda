import 'package:flutter/material.dart';

import '../../../../shared/widgets/fare_breakdown_waterfall.dart';
import '../../domain/models/earnings_models.dart';

/// "Khách trả → Voucher Platform chịu → Voucher Driver chịu → Platform giữ
/// → Thu nhập thực nhận" waterfall for the selected period — thin wrapper
/// around the shared [FareBreakdownWaterfall] (Section 6, Payment/Fare
/// production pass) so the Earnings tab's period-summary card and a single
/// trip's own completion card (`trip_page.dart`) render the exact same
/// waterfall shape and labels.
class EarningsBreakdownCard extends StatelessWidget {
  const EarningsBreakdownCard({super.key, required this.summary});

  final EarningsSummary summary;

  @override
  Widget build(BuildContext context) {
    return FareBreakdownWaterfall(
      grossAmountCents: summary.totalCents,
      currency: summary.currency,
    );
  }
}
