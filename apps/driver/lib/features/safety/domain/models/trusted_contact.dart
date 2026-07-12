/// A driver-added emergency contact. Session-only — there is no
/// trusted-contact backend anywhere in this project, so the list this model
/// backs starts empty and lives purely in memory for the lifetime of the
/// page (see `TrustedContactsPage`), not persisted to disk or a server.
class TrustedContact {
  const TrustedContact({
    required this.id,
    required this.name,
    required this.phone,
    required this.relationship,
  });

  final String id;
  final String name;
  final String phone;
  final String relationship;

  TrustedContact copyWith({String? name, String? phone, String? relationship}) {
    return TrustedContact(
      id: id,
      name: name ?? this.name,
      phone: phone ?? this.phone,
      relationship: relationship ?? this.relationship,
    );
  }
}
