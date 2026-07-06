import 'package:flutter/material.dart';

import '../../domain/models/driver_availability_status.dart';

/// Large primary Online/Offline switch. Supports all four states
/// (Offline/Going Online/Online/Going Offline) with a mock transition delay
/// — there is no real availability API call here (see
/// `docs/project/MVP_DEVELOPMENT_PLAN.md` Driver App Roadmap stage D3).
///
/// This widget owns the mock transition timing; [onStatusChanged] is called
/// on every state change (including the transient ones) so the parent page
/// can drive the Home Status Card in sync.
class AvailabilityToggle extends StatefulWidget {
  const AvailabilityToggle({super.key, required this.onStatusChanged});

  final ValueChanged<DriverAvailabilityStatus> onStatusChanged;

  @override
  State<AvailabilityToggle> createState() => _AvailabilityToggleState();
}

class _AvailabilityToggleState extends State<AvailabilityToggle> {
  DriverAvailabilityStatus _status = DriverAvailabilityStatus.offline;

  Future<void> _toggle() async {
    if (_status.isTransitioning) return;

    if (_status == DriverAvailabilityStatus.offline) {
      _setStatus(DriverAvailabilityStatus.goingOnline);
      await Future.delayed(const Duration(milliseconds: 1200));
      if (!mounted) return;
      _setStatus(DriverAvailabilityStatus.online);
    } else {
      _setStatus(DriverAvailabilityStatus.goingOffline);
      await Future.delayed(const Duration(milliseconds: 900));
      if (!mounted) return;
      _setStatus(DriverAvailabilityStatus.offline);
    }
  }

  void _setStatus(DriverAvailabilityStatus status) {
    setState(() => _status = status);
    widget.onStatusChanged(status);
  }

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    final isOnline = _status.isOnlineOrBecomingOnline;
    final background = isOnline ? primary : Colors.grey.shade200;
    final foreground = isOnline ? Colors.white : Colors.black87;

    return GestureDetector(
      onTap: _toggle,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 350),
        curve: Curves.easeOut,
        width: double.infinity,
        padding: const EdgeInsets.symmetric(vertical: 22, horizontal: 20),
        decoration: BoxDecoration(
          color: background,
          borderRadius: BorderRadius.circular(16),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            _statusIcon(foreground),
            const SizedBox(width: 12),
            Flexible(
              child: AnimatedSwitcher(
                duration: const Duration(milliseconds: 200),
                child: Text(
                  _status.actionLabel,
                  key: ValueKey(_status),
                  textAlign: TextAlign.center,
                  style: TextStyle(
                    color: foreground,
                    fontWeight: FontWeight.w700,
                    fontSize: 16,
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _statusIcon(Color color) {
    if (_status.isTransitioning) {
      return SizedBox(
        width: 22,
        height: 22,
        child: CircularProgressIndicator(strokeWidth: 2.4, color: color),
      );
    }
    return Icon(_status.icon, color: color, size: 24);
  }
}
