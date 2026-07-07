/// Rider identity shown on an incoming trip offer. Mock data only — the
/// User service already has a real `UserProfile` entity
/// (`backend/services/user`), but nothing here calls it.
class RiderInfo {
  const RiderInfo({required this.name, required this.rating});

  final String name;
  final double rating;

  /// Single-letter placeholder shown in a `CircleAvatar` in place of a real
  /// rider photo — no image asset or network fetch is involved.
  String get avatarInitial => name.isNotEmpty ? name[0].toUpperCase() : '?';
}
