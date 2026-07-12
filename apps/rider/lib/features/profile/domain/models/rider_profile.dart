import 'package:flutter/material.dart';

/// Rider loyalty tier shown on the Profile screen.
///
/// Mock only — mirrors the naming convention of the Driver commission tiers
/// in DOC-0002 §6.7 (Standard → Platinum) for flavour, but is not the same
/// concept and is not backed by any service.
enum MemberLevel { standard, silver, gold, platinum }

extension MemberLevelX on MemberLevel {
  String get label => switch (this) {
        MemberLevel.standard => 'Tiêu chuẩn',
        MemberLevel.silver => 'Bạc',
        MemberLevel.gold => 'Vàng',
        MemberLevel.platinum => 'Bạch kim',
      };

  Color get color => switch (this) {
        MemberLevel.standard => const Color(0xFF6B7280),
        MemberLevel.silver => const Color(0xFF64748B),
        MemberLevel.gold => const Color(0xFFB8860B),
        MemberLevel.platinum => const Color(0xFF7C3AED),
      };
}

/// Rider identity + stats shown on the Profile screen.
///
/// Mock data only — the real `user` service already has a `UserProfile`
/// entity (`backend/services/user`), but there is no API client or
/// authentication yet (see `docs/project/MVP_DEVELOPMENT_PLAN.md` Rider App
/// Roadmap stage R8: "Wire Profile tab to User service").
class RiderProfile {
  const RiderProfile({
    required this.fullName,
    required this.phoneNumber,
    required this.memberLevel,
    required this.rating,
    required this.totalCompletedTrips,
  });

  final String fullName;
  final String phoneNumber;
  final MemberLevel memberLevel;
  final double rating;
  final int totalCompletedTrips;

  /// Single-letter placeholder shown in a `CircleAvatar` in place of a real
  /// profile photo — no image asset or network fetch is involved.
  String get avatarInitial => fullName.isNotEmpty ? fullName[0].toUpperCase() : '?';
}
