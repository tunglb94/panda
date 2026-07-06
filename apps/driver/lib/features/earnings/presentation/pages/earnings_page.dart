import 'package:flutter/material.dart';

class EarningsPage extends StatelessWidget {
  const EarningsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Earnings')),
      body: const SafeArea(
        child: Padding(
          padding: EdgeInsets.all(16),
          child: Column(
            children: [
              _EarningsTotalCard(),
              SizedBox(height: 16),
              Expanded(child: _WeeklySummaryCard()),
            ],
          ),
        ),
      ),
    );
  }
}

class _EarningsTotalCard extends StatelessWidget {
  const _EarningsTotalCard();

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Card(
      color: cs.primaryContainer,
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              "Today's Earnings",
              style: Theme.of(context)
                  .textTheme
                  .titleMedium
                  ?.copyWith(color: cs.onPrimaryContainer),
            ),
            const SizedBox(height: 8),
            Text(
              '\$0.00',
              style: Theme.of(context).textTheme.displaySmall?.copyWith(
                    fontWeight: FontWeight.bold,
                    color: cs.onPrimaryContainer,
                  ),
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                _EarningsLine(label: 'Trips', value: '\$0.00', cs: cs),
                const SizedBox(width: 24),
                _EarningsLine(label: 'Bonuses', value: '\$0.00', cs: cs),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _EarningsLine extends StatelessWidget {
  const _EarningsLine({
    required this.label,
    required this.value,
    required this.cs,
  });

  final String label;
  final String value;
  final ColorScheme cs;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          value,
          style: Theme.of(context).textTheme.titleLarge?.copyWith(
                color: cs.onPrimaryContainer,
                fontWeight: FontWeight.w600,
              ),
        ),
        Text(
          label,
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: cs.onPrimaryContainer.withValues(alpha: 0.7),
              ),
        ),
      ],
    );
  }
}

class _WeeklySummaryCard extends StatelessWidget {
  const _WeeklySummaryCard();

  static const _days = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('This Week', style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 16),
            for (final day in _days) ...[
              _DayRow(day: day, amount: '\$0.00'),
              const SizedBox(height: 8),
            ],
          ],
        ),
      ),
    );
  }
}

class _DayRow extends StatelessWidget {
  const _DayRow({required this.day, required this.amount});

  final String day;
  final String amount;

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Row(
      children: [
        SizedBox(
          width: 36,
          child: Text(day, style: Theme.of(context).textTheme.bodyMedium),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: LinearProgressIndicator(
            value: 0,
            backgroundColor: cs.surfaceContainerHighest,
          ),
        ),
        const SizedBox(width: 12),
        Text(amount, style: Theme.of(context).textTheme.bodyMedium),
      ],
    );
  }
}
