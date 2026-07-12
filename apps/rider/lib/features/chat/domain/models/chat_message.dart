/// One chat bubble. [seq] is the backend's monotonic ordering cursor — used
/// as the `since_id` value for the next list/poll call, never shown in the UI.
class ChatMessage {
  const ChatMessage({
    required this.id,
    required this.seq,
    required this.conversationId,
    required this.senderId,
    required this.senderRole,
    required this.body,
    required this.quickReplyKey,
    required this.createdAt,
    this.deliveredAt,
    this.readAt,
    this.pending = false,
  });

  final String id;
  final int seq;
  final String conversationId;
  final String senderId;

  /// "rider" or "driver".
  final String senderRole;
  final String body;
  final String quickReplyKey;
  final DateTime createdAt;
  final DateTime? deliveredAt;
  final DateTime? readAt;

  /// True only for a message this device queued locally and hasn't
  /// confirmed as sent yet (Offline Strategy) — never set from a server
  /// response, always false for anything the backend actually returned.
  final bool pending;

  bool get isRead => readAt != null;

  factory ChatMessage.fromJson(Map<String, dynamic> json) => ChatMessage(
        id: json['id'] as String? ?? '',
        seq: (json['seq'] as num?)?.toInt() ?? 0,
        conversationId: json['conversation_id'] as String? ?? '',
        senderId: json['sender_id'] as String? ?? '',
        senderRole: json['sender_role'] as String? ?? '',
        body: json['body'] as String? ?? '',
        quickReplyKey: json['quick_reply_key'] as String? ?? '',
        createdAt: DateTime.tryParse(json['created_at'] as String? ?? '')?.toLocal() ?? DateTime.now(),
        deliveredAt: json['delivered_at'] == null
            ? null
            : DateTime.tryParse(json['delivered_at'] as String)?.toLocal(),
        readAt: json['read_at'] == null ? null : DateTime.tryParse(json['read_at'] as String)?.toLocal(),
      );

  ChatMessage copyWith({bool? pending}) => ChatMessage(
        id: id,
        seq: seq,
        conversationId: conversationId,
        senderId: senderId,
        senderRole: senderRole,
        body: body,
        quickReplyKey: quickReplyKey,
        createdAt: createdAt,
        deliveredAt: deliveredAt,
        readAt: readAt,
        pending: pending ?? this.pending,
      );
}
