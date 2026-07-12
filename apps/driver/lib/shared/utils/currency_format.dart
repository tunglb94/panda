/// Formats a monetary [amount] (in the smallest unit of [currencyCode]) for
/// display.
///
/// VND has no decimal subunit (Business Rule Bible v1.0 §2.2.9) — amounts in
/// that currency are already whole VND, so they're shown with thousand
/// separators and a trailing "đ", not divided or given decimal places.
/// Other currencies (e.g. legacy USD test trips recorded before the pricing
/// service was calibrated to VND) keep the old cents/2-decimal formatting so
/// historical records aren't reinterpreted under the wrong scale.
String formatMoney(int amount, String currencyCode) {
  final code = currencyCode.toUpperCase();
  if (code.isEmpty) return '—';
  if (code == 'VND') return '${_withThousandSeparators(amount)} đ';
  final symbol = code == 'USD' ? r'$' : '$code ';
  return '$symbol${(amount / 100).toStringAsFixed(2)}';
}

String _withThousandSeparators(int value) {
  final digits = value.abs().toString();
  final buffer = StringBuffer();
  for (var i = 0; i < digits.length; i++) {
    final fromEnd = digits.length - i;
    if (i > 0 && fromEnd % 3 == 0) buffer.write('.');
    buffer.write(digits[i]);
  }
  return value < 0 ? '-${buffer.toString()}' : buffer.toString();
}
