import 'package:flutter/material.dart';

/// Mock waiting-fee display: free for the first [_freeMinutes] minutes,
/// then [_perMinuteFeeVnd] VND per additional minute. Pure UI — no backend
/// billing, no repository; [elapsedMinutes] is supplied by the parent
/// (fed from `WaitingTimer.onMinutePassed`).
class WaitingFeeCard extends StatelessWidget {
  const WaitingFeeCard({super.key, required this.elapsedMinutes});

  final int elapsedMinutes;

  static const _freeMinutes = 5;
  static const _perMinuteFeeVnd = 2000;

  int get _feeVnd =>
      elapsedMinutes <= _freeMinutes ? 0 : (elapsedMinutes - _freeMinutes) * _perMinuteFeeVnd;

  String get _formattedFee {
    if (_feeVnd == 0) return 'Free';
    final digits = _feeVnd.toString();
    final buffer = StringBuffer();
    for (var i = 0; i < digits.length; i++) {
      if (i > 0 && (digits.length - i) % 3 == 0) buffer.write('.');
      buffer.write(digits[i]);
    }
    return '$bufferđ';
  }

  @override
  Widget build(BuildContext context) {
    final isFree = _feeVnd == 0;
    final color = isFree ? Colors.green.shade700 : Colors.orange.shade800;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(10),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text('Waiting fee', style: TextStyle(fontSize: 12, color: Colors.grey.shade600)),
          Text(_formattedFee, style: TextStyle(fontWeight: FontWeight.bold, color: color)),
        ],
      ),
    );
  }
}
