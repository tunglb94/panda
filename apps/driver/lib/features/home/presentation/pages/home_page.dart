import 'package:flutter/material.dart';

import 'package:driver/shared/widgets/async_state_view.dart';

import '../../data/driver_home_repository.dart';
import '../../domain/models/driver_activity_status.dart';
import '../../domain/models/driver_availability_status.dart';
import '../../domain/models/driver_home_summary.dart';
import '../widgets/availability_toggle.dart';
import '../widgets/driver_summary_header.dart';
import '../widgets/home_stats_row.dart';
import '../widgets/home_status_card.dart';
import '../widgets/quick_actions_section.dart';

enum _PreviewAction { normalData, emptyData, errorData, simulateBusy, clearBusy }

/// Home dashboard: driver summary, today's stats, the Online/Offline
/// toggle, a status card, and Quick Actions. Fetches mock data through
/// `DriverHomeRepository` via the shared `AsyncStateView` (Loading/
/// Success/Empty/Error).
class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  static const _repository = DriverHomeRepository();

  DriverHomeDemoMode _mode = DriverHomeDemoMode.normal;
  late Future<DriverHomeSummary> _future;
  DriverActivityStatus _activity = DriverActivityStatus.offline;
  bool _busyOverride = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  void _load() {
    setState(() {
      _future = _repository.fetchSummary(mode: _mode);
    });
  }

  void _handleAvailabilityChanged(DriverAvailabilityStatus status) {
    if (_busyOverride) return;
    switch (status) {
      case DriverAvailabilityStatus.offline:
      case DriverAvailabilityStatus.goingOffline:
        setState(() => _activity = DriverActivityStatus.offline);
      case DriverAvailabilityStatus.goingOnline:
        break;
      case DriverAvailabilityStatus.online:
        setState(() => _activity = DriverActivityStatus.waitingForTrips);
        _scheduleSearchingNearby();
    }
  }

  Future<void> _scheduleSearchingNearby() async {
    await Future.delayed(const Duration(seconds: 3));
    if (!mounted || _busyOverride) return;
    if (_activity == DriverActivityStatus.waitingForTrips) {
      setState(() => _activity = DriverActivityStatus.searchingNearby);
    }
  }

  void _handlePreviewAction(_PreviewAction action) {
    switch (action) {
      case _PreviewAction.normalData:
        _mode = DriverHomeDemoMode.normal;
        _load();
      case _PreviewAction.emptyData:
        _mode = DriverHomeDemoMode.empty;
        _load();
      case _PreviewAction.errorData:
        _mode = DriverHomeDemoMode.error;
        _load();
      case _PreviewAction.simulateBusy:
        setState(() {
          _busyOverride = true;
          _activity = DriverActivityStatus.busy;
        });
      case _PreviewAction.clearBusy:
        setState(() {
          _busyOverride = false;
          _activity = DriverActivityStatus.offline;
        });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Home'),
        actions: [
          PopupMenuButton<_PreviewAction>(
            tooltip: 'Preview state (dev)',
            icon: const Icon(Icons.tune),
            onSelected: _handlePreviewAction,
            itemBuilder: (context) => const [
              PopupMenuItem(value: _PreviewAction.normalData, child: Text('Normal')),
              PopupMenuItem(value: _PreviewAction.emptyData, child: Text('Empty (dev)')),
              PopupMenuItem(value: _PreviewAction.errorData, child: Text('Error (dev)')),
              PopupMenuDivider(),
              PopupMenuItem(
                value: _PreviewAction.simulateBusy,
                child: Text('Simulate busy (dev)'),
              ),
              PopupMenuItem(value: _PreviewAction.clearBusy, child: Text('Clear busy (dev)')),
            ],
          ),
        ],
      ),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: AsyncStateView<DriverHomeSummary>(
                future: _future,
                isEmpty: (summary) => !summary.hasVehicle,
                emptyBuilder: (context) => Padding(
                  padding: const EdgeInsets.symmetric(vertical: 48, horizontal: 24),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.directions_car_outlined, size: 48, color: Colors.grey.shade400),
                      const SizedBox(height: 12),
                      const Text('No vehicle assigned yet',
                          style: TextStyle(fontWeight: FontWeight.w600)),
                      const SizedBox(height: 4),
                      Text(
                        'Add a vehicle to start going online and receiving trips.',
                        textAlign: TextAlign.center,
                        style: TextStyle(color: Colors.grey.shade500),
                      ),
                    ],
                  ),
                ),
                errorBuilder: (context, error) => Padding(
                  padding: const EdgeInsets.symmetric(vertical: 48, horizontal: 24),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.error_outline, size: 48, color: Colors.red.shade400),
                      const SizedBox(height: 12),
                      const Text("Couldn't load your dashboard",
                          style: TextStyle(fontWeight: FontWeight.w600)),
                      const SizedBox(height: 12),
                      OutlinedButton(onPressed: _load, child: const Text('Retry')),
                    ],
                  ),
                ),
                successBuilder: (context, summary) => Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    DriverSummaryHeader(summary: summary),
                    const SizedBox(height: 20),
                    HomeStatsRow(summary: summary),
                    const SizedBox(height: 20),
                    AvailabilityToggle(onStatusChanged: _handleAvailabilityChanged),
                    const SizedBox(height: 16),
                    HomeStatusCard(status: _activity),
                    const SizedBox(height: 24),
                    Text(
                      'Quick actions',
                      style: Theme.of(context)
                          .textTheme
                          .titleSmall
                          ?.copyWith(fontWeight: FontWeight.w600),
                    ),
                    const SizedBox(height: 12),
                    const QuickActionsSection(),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
