import 'package:flutter/material.dart';

import '../auth/auth_state.dart';
import '../storage/token_storage.dart';

/// One sidebar destination — icon/label/route, matched against the current
/// location to highlight the active item.
class AdminNavItem {
  const AdminNavItem({required this.path, required this.label, required this.icon});

  final String path;
  final String label;
  final IconData icon;
}

/// Minimal shared shell: fixed sidebar (nav items + logout) + content area.
/// Replaces the previous per-page AppBar cross-link buttons (KYC page linked
/// to Vouchers and vice versa) with one place to add future sections.
/// Deliberately plain — no NavigationRail/Drawer framework, just a Row, to
/// match this app's existing "hand-rolled widgets over Flutter framework
/// abstractions" style (see driver_verifications_page.dart).
class AdminShell extends StatelessWidget {
  const AdminShell({
    super.key,
    required this.currentPath,
    required this.items,
    required this.onNavigate,
    required this.authState,
    required this.tokenStorage,
    required this.child,
  });

  final String currentPath;
  final List<AdminNavItem> items;
  final ValueChanged<String> onNavigate;
  final AuthState authState;
  final TokenStorage tokenStorage;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Row(
        children: [
          _Sidebar(
            currentPath: currentPath,
            items: items,
            onNavigate: onNavigate,
            onLogout: () => authState.logout(tokenStorage),
          ),
          const VerticalDivider(width: 1),
          Expanded(child: child),
        ],
      ),
    );
  }
}

class _Sidebar extends StatelessWidget {
  const _Sidebar({
    required this.currentPath,
    required this.items,
    required this.onNavigate,
    required this.onLogout,
  });

  final String currentPath;
  final List<AdminNavItem> items;
  final ValueChanged<String> onNavigate;
  final VoidCallback onLogout;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 220,
      color: const Color(0xFF111827),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const Padding(
            padding: EdgeInsets.fromLTRB(20, 24, 20, 24),
            child: Text(
              'Panda Admin',
              style: TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.bold),
            ),
          ),
          for (final item in items)
            _NavTile(
              item: item,
              selected: currentPath == item.path,
              onTap: () => onNavigate(item.path),
            ),
          const Spacer(),
          const Divider(color: Color(0xFF374151), height: 1),
          _NavTile(
            item: const AdminNavItem(path: '', label: 'Đăng xuất', icon: Icons.logout),
            selected: false,
            onTap: onLogout,
          ),
          const SizedBox(height: 12),
        ],
      ),
    );
  }
}

class _NavTile extends StatelessWidget {
  const _NavTile({required this.item, required this.selected, required this.onTap});

  final AdminNavItem item;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final color = selected ? Colors.white : const Color(0xFF9CA3AF);
    return Material(
      color: selected ? const Color(0xFF1F2937) : Colors.transparent,
      child: InkWell(
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 14),
          child: Row(
            children: [
              Icon(item.icon, color: color, size: 20),
              const SizedBox(width: 12),
              Text(item.label, style: TextStyle(color: color, fontWeight: selected ? FontWeight.w600 : FontWeight.normal)),
            ],
          ),
        ),
      ),
    );
  }
}
