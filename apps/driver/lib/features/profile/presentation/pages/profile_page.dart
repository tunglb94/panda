import 'package:flutter/material.dart';

import 'developer_page.dart';

/// Profile tab placeholder. The only functional entry point in this phase
/// is "Developer" (Development Utilities requirement) — everything else
/// about the driver's own profile arrives in a future phase, alongside
/// `apps/rider`'s equivalent Profile module.
class ProfilePage extends StatelessWidget {
  const ProfilePage({super.key});

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    return Scaffold(
      appBar: AppBar(title: const Text('Profile')),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                children: [
                  CircleAvatar(
                    radius: 44,
                    backgroundColor: primary.withValues(alpha: 0.12),
                    child: Icon(Icons.person, size: 44, color: primary),
                  ),
                  const SizedBox(height: 12),
                  const Text(
                    'Mock Driver',
                    style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    'Driver profile details arrive in a future phase.',
                    style: TextStyle(color: Colors.grey.shade500),
                  ),
                  const SizedBox(height: 24),
                  ListTile(
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                      side: BorderSide(color: Colors.grey.shade200),
                    ),
                    leading: Icon(Icons.developer_mode_outlined, color: primary),
                    title: const Text('Developer'),
                    trailing: const Icon(Icons.chevron_right, color: Colors.grey),
                    onTap: () => Navigator.of(context).push(
                      MaterialPageRoute(builder: (_) => const DeveloperPage()),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
