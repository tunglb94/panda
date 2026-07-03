import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:rider/core/router/app_router.dart';

class HomePage extends StatelessWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    return Scaffold(
      appBar: AppBar(
        title: const Text(
          'FAIRRIDE',
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.notifications_outlined),
            onPressed: () {},
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Where to?',
              style: textTheme.headlineMedium
                  ?.copyWith(fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 16),
            _SearchBar(onTap: () => context.go(AppRoutes.booking)),
            const SizedBox(height: 24),
            Text(
              'Recent places',
              style: textTheme.titleMedium
                  ?.copyWith(fontWeight: FontWeight.w600),
            ),
            const SizedBox(height: 12),
            const _RecentPlaceTile(
              icon: Icons.home_outlined,
              label: 'Home',
              subtitle: '123 Main Street',
            ),
            const _RecentPlaceTile(
              icon: Icons.work_outlined,
              label: 'Work',
              subtitle: '456 Office Boulevard',
            ),
            const SizedBox(height: 24),
            Text(
              'Ride categories',
              style: textTheme.titleMedium
                  ?.copyWith(fontWeight: FontWeight.w600),
            ),
            const SizedBox(height: 12),
            const Row(
              children: [
                _RideCategory(icon: Icons.directions_car, label: 'Car'),
                SizedBox(width: 12),
                _RideCategory(icon: Icons.two_wheeler, label: 'Moto'),
                SizedBox(width: 12),
                _RideCategory(icon: Icons.airport_shuttle, label: 'Van'),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _SearchBar extends StatelessWidget {
  const _SearchBar({required this.onTap});

  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: Colors.grey.shade200),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.05),
              blurRadius: 8,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Row(
          children: [
            const Icon(Icons.search, color: Color(0xFF1A8C4E)),
            const SizedBox(width: 12),
            Text(
              'Enter destination',
              style: TextStyle(color: Colors.grey.shade500, fontSize: 16),
            ),
          ],
        ),
      ),
    );
  }
}

class _RecentPlaceTile extends StatelessWidget {
  const _RecentPlaceTile({
    required this.icon,
    required this.label,
    required this.subtitle,
  });

  final IconData icon;
  final String label;
  final String subtitle;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      contentPadding: EdgeInsets.zero,
      leading: CircleAvatar(
        backgroundColor: Colors.grey.shade100,
        child: Icon(icon, color: const Color(0xFF1A8C4E)),
      ),
      title: Text(label,
          style: const TextStyle(fontWeight: FontWeight.w500)),
      subtitle: Text(
        subtitle,
        style: TextStyle(color: Colors.grey.shade500, fontSize: 13),
      ),
      trailing: const Icon(Icons.north_west, size: 18),
      onTap: () {},
    );
  }
}

class _RideCategory extends StatelessWidget {
  const _RideCategory({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: Colors.grey.shade200),
        ),
        child: Column(
          children: [
            Icon(icon, color: const Color(0xFF1A8C4E), size: 28),
            const SizedBox(height: 8),
            Text(
              label,
              style: const TextStyle(
                  fontWeight: FontWeight.w500, fontSize: 13),
            ),
          ],
        ),
      ),
    );
  }
}
