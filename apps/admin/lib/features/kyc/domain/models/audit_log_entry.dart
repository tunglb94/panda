/// One entry of the KYC audit timeline (Phần 9) — append-only history of
/// Submitted -> Rejected -> Submitted -> Approved, etc.
class AuditLogEntry {
  const AuditLogEntry({
    required this.entityType,
    required this.action,
    required this.actorId,
    required this.reason,
    required this.createdAt,
  });

  final String entityType;
  final String action;
  final String actorId;
  final String reason;
  final DateTime createdAt;

  factory AuditLogEntry.fromJson(Map<String, dynamic> json) => AuditLogEntry(
        entityType: json['entity_type'] as String? ?? '',
        action: json['action'] as String? ?? '',
        actorId: json['actor_id'] as String? ?? '',
        reason: json['reason'] as String? ?? '',
        createdAt: DateTime.tryParse(json['created_at'] as String? ?? '') ?? DateTime(1970),
      );
}
