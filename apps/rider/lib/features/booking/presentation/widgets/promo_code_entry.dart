import 'dart:math' as math;

import 'package:flutter/material.dart';

import '../../domain/models/promo_result.dart';

/// Text entry + validation for a mock promo code.
///
/// There is no Promotion backend yet (see
/// `docs/project/MVP_DEVELOPMENT_PLAN.md` §2.1) — validation happens
/// entirely against [MockPromoValidator]. An invalid code triggers a small
/// shake; a valid one shows an inline confirmation.
class PromoCodeEntry extends StatefulWidget {
  const PromoCodeEntry({super.key, required this.onApplied});

  final ValueChanged<PromoResult> onApplied;

  @override
  State<PromoCodeEntry> createState() => _PromoCodeEntryState();
}

class _PromoCodeEntryState extends State<PromoCodeEntry>
    with SingleTickerProviderStateMixin {
  final _controller = TextEditingController();
  late final AnimationController _shakeController;
  PromoResult? _result;

  @override
  void initState() {
    super.initState();
    _shakeController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 400),
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    _shakeController.dispose();
    super.dispose();
  }

  void _apply() {
    final result = MockPromoValidator.validate(_controller.text);
    setState(() => _result = result);
    widget.onApplied(result);
    if (!result.isValid) {
      _shakeController.forward(from: 0);
    }
  }

  @override
  Widget build(BuildContext context) {
    final result = _result;
    return AnimatedBuilder(
      animation: _shakeController,
      builder: (context, child) {
        // Three quick, decaying wiggles that settle back to 0 by t == 1.
        final t = _shakeController.value;
        final shake = 6 * (1 - t) * math.sin(t * 3 * 2 * math.pi);
        return Transform.translate(offset: Offset(shake, 0), child: child);
      },
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                child: TextField(
                  controller: _controller,
                  textCapitalization: TextCapitalization.characters,
                  decoration: InputDecoration(
                    hintText: 'Promo code',
                    isDense: true,
                    prefixIcon: const Icon(Icons.local_offer_outlined),
                    border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12)),
                    errorText:
                        result != null && !result.isValid ? result.message : null,
                  ),
                  onSubmitted: (_) => _apply(),
                ),
              ),
              const SizedBox(width: 8),
              FilledButton.tonal(
                onPressed: _apply,
                child: const Text('Apply'),
              ),
            ],
          ),
          AnimatedSwitcher(
            duration: const Duration(milliseconds: 200),
            child: (result != null && result.isValid)
                ? Padding(
                    key: const ValueKey('applied'),
                    padding: const EdgeInsets.only(top: 6, left: 4),
                    child: Row(
                      children: [
                        Icon(Icons.check_circle,
                            size: 16, color: Colors.green.shade600),
                        const SizedBox(width: 6),
                        Text(
                          result.message,
                          style:
                              TextStyle(color: Colors.green.shade700, fontSize: 12),
                        ),
                      ],
                    ),
                  )
                : const SizedBox.shrink(key: ValueKey('empty')),
          ),
        ],
      ),
    );
  }
}
