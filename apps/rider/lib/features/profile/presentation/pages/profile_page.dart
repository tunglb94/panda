import 'package:flutter/material.dart';

class ProfilePage extends StatelessWidget {
  const ProfilePage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Profile')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            const _ProfileHeader(
              name: 'Alex Rider',
              phone: '+1 555 000 0000',
              rating: 4.9,
            ),
            const SizedBox(height: 24),
            const _SectionLabel('Activity'),
            const _InfoRow(label: 'Total Trips', value: '0'),
            const _InfoRow(label: 'Member Since', value: 'Jul 2026'),
            const SizedBox(height: 16),
            const _SectionLabel('Settings'),
            const _SettingsTile(
                icon: Icons.person_outline, label: 'Personal Information'),
            const _SettingsTile(
                icon: Icons.payment_outlined, label: 'Payment Methods'),
            const _SettingsTile(
                icon: Icons.history, label: 'Trip History'),
            const _SettingsTile(
                icon: Icons.help_outline, label: 'Help & Support'),
            const _SettingsTile(
                icon: Icons.shield_outlined, label: 'Privacy & Safety'),
            const SizedBox(height: 24),
            OutlinedButton(
              onPressed: () {},
              child: const Text('Sign Out'),
            ),
            const SizedBox(height: 16),
          ],
        ),
      ),
    );
  }
}

class _ProfileHeader extends StatelessWidget {
  const _ProfileHeader({
    required this.name,
    required this.phone,
    required this.rating,
  });

  final String name;
  final String phone;
  final double rating;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        const CircleAvatar(
          radius: 44,
          backgroundColor: Color(0xFFE8F5ED),
          child: Icon(Icons.person, size: 44, color: Color(0xFF1A8C4E)),
        ),
        const SizedBox(height: 12),
        Text(name,
            style: const TextStyle(
                fontSize: 20, fontWeight: FontWeight.bold)),
        const SizedBox(height: 4),
        Text(phone,
            style: TextStyle(color: Colors.grey.shade500)),
        const SizedBox(height: 12),
        Container(
          padding:
              const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
          decoration: BoxDecoration(
            color: const Color(0xFFE8F5ED),
            borderRadius: BorderRadius.circular(20),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.star, size: 16, color: Color(0xFF1A8C4E)),
              const SizedBox(width: 4),
              Text(
                rating.toStringAsFixed(1),
                style: const TextStyle(
                  color: Color(0xFF1A8C4E),
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _SectionLabel extends StatelessWidget {
  const _SectionLabel(this.title);

  final String title;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Align(
        alignment: Alignment.centerLeft,
        child: Text(
          title,
          style: const TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.w600,
            color: Color(0xFF6B7280),
          ),
        ),
      ),
    );
  }
}

class _InfoRow extends StatelessWidget {
  const _InfoRow({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label),
          Text(value,
              style: const TextStyle(fontWeight: FontWeight.w600)),
        ],
      ),
    );
  }
}

class _SettingsTile extends StatelessWidget {
  const _SettingsTile({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      contentPadding: EdgeInsets.zero,
      leading: Icon(icon, color: const Color(0xFF1A8C4E)),
      title: Text(label),
      trailing:
          const Icon(Icons.chevron_right, color: Colors.grey),
      onTap: () {},
    );
  }
}
