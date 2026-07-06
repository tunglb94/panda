import 'package:flutter/material.dart';

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  bool _isOnline = false;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('FAIRRIDE Driver'),
        centerTitle: false,
      ),
      body: SafeArea(
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            _StatusCard(isOnline: _isOnline, onToggle: _toggle),
            const SizedBox(height: 16),
            const _SummaryCard(),
          ],
        ),
      ),
    );
  }

  void _toggle() => setState(() => _isOnline = !_isOnline);
}

class _StatusCard extends StatelessWidget {
  const _StatusCard({required this.isOnline, required this.onToggle});

  final bool isOnline;
  final VoidCallback onToggle;

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Row(
          children: [
            Container(
              width: 12,
              height: 12,
              decoration: BoxDecoration(
                color: isOnline ? const Color(0xFF1A8C4E) : cs.outline,
                shape: BoxShape.circle,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                isOnline ? 'You are online' : 'You are offline',
                style: Theme.of(context).textTheme.titleMedium,
              ),
            ),
            Switch(value: isOnline, onChanged: (_) => onToggle()),
          ],
        ),
      ),
    );
  }
}

class _SummaryCard extends StatelessWidget {
  const _SummaryCard();

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              "Today's Summary",
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 16),
            const Row(
              children: [
                _Stat(label: 'Trips', value: '0'),
                SizedBox(width: 32),
                _Stat(label: 'Earnings', value: '\$0.00'),
                SizedBox(width: 32),
                _Stat(label: 'Hours', value: '0.0'),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _Stat extends StatelessWidget {
  const _Stat({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          value,
          style: Theme.of(context)
              .textTheme
              .headlineSmall
              ?.copyWith(fontWeight: FontWeight.bold),
        ),
        Text(label, style: Theme.of(context).textTheme.bodySmall),
      ],
    );
  }
}
