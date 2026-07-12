import 'dart:convert';

import 'package:shared_preferences/shared_preferences.dart';

/// A chat message queued locally because sending it failed while offline
/// (Part 6 — Offline Strategy). Persisted to [SharedPreferences] as JSON so
/// it survives an app restart before the retry succeeds.
class PendingMessage {
  const PendingMessage({
    required this.localId,
    required this.conversationId,
    required this.text,
    required this.quickReplyKey,
    required this.createdAt,
  });

  final String localId;
  final String conversationId;
  final String text;
  final String quickReplyKey;
  final DateTime createdAt;

  Map<String, dynamic> toJson() => {
        'local_id': localId,
        'conversation_id': conversationId,
        'text': text,
        'quick_reply_key': quickReplyKey,
        'created_at': createdAt.toIso8601String(),
      };

  factory PendingMessage.fromJson(Map<String, dynamic> json) => PendingMessage(
        localId: json['local_id'] as String? ?? '',
        conversationId: json['conversation_id'] as String? ?? '',
        text: json['text'] as String? ?? '',
        quickReplyKey: json['quick_reply_key'] as String? ?? '',
        createdAt: DateTime.tryParse(json['created_at'] as String? ?? '') ?? DateTime.now(),
      );
}

/// Local send queue for chat messages sent while offline. A message is
/// enqueued only when [ChatRepository.sendMessage] fails with
/// `ApiException.statusCode == 0` (the app's universal client-side
/// offline/timeout sentinel — see `ApiClient`), never for a genuine backend
/// rejection (e.g. closed conversation), which must surface as an error
/// instead of silently retrying forever.
class OfflineMessageQueue {
  const OfflineMessageQueue();

  static const _storageKey = 'chat_pending_messages_v1';

  Future<List<PendingMessage>> loadAll() async {
    final prefs = await SharedPreferences.getInstance();
    final raw = prefs.getString(_storageKey);
    if (raw == null || raw.isEmpty) return [];
    try {
      final list = jsonDecode(raw) as List<dynamic>;
      return list.map((e) => PendingMessage.fromJson(e as Map<String, dynamic>)).toList();
    } catch (_) {
      return [];
    }
  }

  Future<List<PendingMessage>> forConversation(String conversationId) async {
    final all = await loadAll();
    return all.where((m) => m.conversationId == conversationId).toList();
  }

  Future<void> enqueue(PendingMessage message) async {
    final all = await loadAll();
    all.add(message);
    await _saveAll(all);
  }

  Future<void> remove(String localId) async {
    final all = await loadAll();
    all.removeWhere((m) => m.localId == localId);
    await _saveAll(all);
  }

  Future<void> _saveAll(List<PendingMessage> items) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_storageKey, jsonEncode(items.map((e) => e.toJson()).toList()));
  }
}
