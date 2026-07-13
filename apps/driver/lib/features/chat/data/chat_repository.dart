import '../../../core/network/api_client.dart';

import '../domain/models/chat_message.dart';
import '../domain/models/conversation.dart';

/// Long-poll GET timeout — the client waits a little longer than the
/// server's own hold time (`notificationapp.DefaultPollTimeout`, ~25s) so a
/// slow-but-real server response is never mistaken for a client timeout.
const Duration chatPollTimeout = Duration(seconds: 30);

class ChatRepository {
  const ChatRepository(this._client);

  final ApiClient _client;

  Future<Conversation> getOrCreateConversation(String tripId) async {
    final body = await _client.get('/api/v1/rides/$tripId/conversation');
    return Conversation.fromJson(body);
  }

  Future<List<ChatMessage>> listMessages(String conversationId, {int sinceSeq = 0}) async {
    final body = await _client.get('/api/v1/conversations/$conversationId/messages?since_id=$sinceSeq');
    return _parseMessages(body);
  }

  /// Genuine long-poll: the server holds this request open until either a
  /// new message arrives or its own timeout elapses, so this resolves
  /// almost immediately after a new message is sent rather than waiting for
  /// the next scheduled poll tick.
  Future<List<ChatMessage>> pollMessages(String conversationId, {int sinceSeq = 0}) async {
    final body = await _client.get(
      '/api/v1/conversations/$conversationId/messages?since_id=$sinceSeq&poll=true',
      timeout: chatPollTimeout,
    );
    return _parseMessages(body);
  }

  List<ChatMessage> _parseMessages(Map<String, dynamic> body) {
    final raw = (body['messages'] as List<dynamic>?) ?? const [];
    return raw.map((e) => ChatMessage.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<ChatMessage> sendMessage(
    String conversationId, {
    String text = '',
    String quickReplyKey = '',
  }) async {
    final body = await _client.post(
      '/api/v1/conversations/$conversationId/messages',
      body: {'text': text, 'quick_reply_key': quickReplyKey},
    );
    return ChatMessage.fromJson(body);
  }

  Future<void> markRead(String conversationId) async {
    await _client.post('/api/v1/conversations/$conversationId/read');
  }
}
