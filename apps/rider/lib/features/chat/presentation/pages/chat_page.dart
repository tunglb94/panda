import 'dart:async';

import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/shared/widgets/app_empty_state.dart';
import 'package:rider/shared/widgets/app_skeleton.dart';
import 'package:rider/shared/widgets/app_snackbar.dart';

import '../../data/chat_repository.dart';
import '../../data/offline_message_queue.dart';
import '../../domain/models/chat_message.dart';
import '../../domain/models/conversation.dart';
import '../../domain/models/quick_reply.dart';
import '../widgets/message_bubble.dart';
import '../widgets/quick_reply_row.dart';

/// This app's own side of every conversation — a Rider never renders a
/// bubble as "mine" unless the message's sender_role is exactly this.
const String _mySenderRole = 'rider';

/// In-app chat for one trip (Ride or Delivery — both share this page, see
/// Part 8). Long-polls `GET .../messages?poll=true`, immediately re-issuing
/// the next poll as soon as the previous one resolves — the server itself
/// holds each request open (~25s) until a new message arrives or it times
/// out, so this is a real long-poll loop, not a short fixed-interval one.
class ChatPage extends StatefulWidget {
  const ChatPage({super.key, required this.tripId, required this.apiClient});

  final String tripId;
  final ApiClient apiClient;

  @override
  State<ChatPage> createState() => _ChatPageState();
}

class _ChatPageState extends State<ChatPage> with WidgetsBindingObserver {
  late final ChatRepository _repo = ChatRepository(widget.apiClient);
  static const _queue = OfflineMessageQueue();

  Conversation? _conversation;
  final List<ChatMessage> _messages = [];
  final Set<String> _seenIds = {};
  int _lastSeq = 0;

  bool _loading = true;
  String? _initError;
  bool _isPolling = false;
  bool _sending = false;
  Timer? _pollRetryTimer;
  final _textController = TextEditingController();
  final _scrollController = ScrollController();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _init();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _pollRetryTimer?.cancel();
    _textController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed && _conversation != null) {
      _pollRetryTimer?.cancel();
      _pollNext();
      _retryPendingQueue();
    }
  }

  Future<void> _init() async {
    try {
      final conv = await _repo.getOrCreateConversation(widget.tripId);
      if (!mounted) return;
      setState(() {
        _conversation = conv;
        _loading = false;
      });
      await _loadInitialMessages();
      unawaited(_repo.markRead(conv.id).catchError((_) {}));
      unawaited(_retryPendingQueue());
      _pollNext();
    } on ApiException catch (e) {
      if (mounted) {
        setState(() {
          _loading = false;
          _initError = e.statusCode == 0 ? e.message : _friendlyInitError(e);
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() {
          _loading = false;
          _initError = 'Không thể mở cuộc trò chuyện. Vui lòng thử lại.';
        });
      }
    }
  }

  String _friendlyInitError(ApiException e) {
    final msg = e.message.toLowerCase();
    if (msg.contains('driver')) return 'Chưa có tài xế được gán cho chuyến này.';
    if (msg.contains('participant')) return 'Bạn không có quyền truy cập cuộc trò chuyện này.';
    return 'Không thể mở cuộc trò chuyện. Vui lòng thử lại.';
  }

  Future<void> _loadInitialMessages() async {
    final conv = _conversation;
    if (conv == null) return;
    try {
      final msgs = await _repo.listMessages(conv.id);
      _mergeMessages(msgs);
    } catch (_) {
      // Non-fatal — the poll loop below will pick messages up regardless.
    }
  }

  void _mergeMessages(List<ChatMessage> incoming) {
    var changed = false;
    for (final m in incoming) {
      if (_seenIds.add(m.id)) {
        _messages.add(m);
        changed = true;
      }
      if (m.seq > _lastSeq) _lastSeq = m.seq;
    }
    if (!changed) return;
    _messages.sort((a, b) => a.seq.compareTo(b.seq));
    if (mounted) setState(() {});
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!_scrollController.hasClients) return;
      _scrollController.animateTo(
        _scrollController.position.maxScrollExtent,
        duration: const Duration(milliseconds: 220),
        curve: Curves.easeOut,
      );
    });
  }

  Future<void> _pollNext() async {
    if (!mounted || _isPolling) return;
    final conv = _conversation;
    if (conv == null) return;
    _isPolling = true;
    try {
      final msgs = await _repo.pollMessages(conv.id, sinceSeq: _lastSeq);
      if (!mounted) return;
      if (msgs.isNotEmpty) {
        _mergeMessages(msgs);
        unawaited(_repo.markRead(conv.id).catchError((_) {}));
      }
      unawaited(_retryPendingQueue());
    } catch (_) {
      // A single failed poll cycle is not shown to the rider — the loop
      // below just tries again after a short backoff.
    } finally {
      _isPolling = false;
    }
    if (!mounted) return;
    _pollRetryTimer?.cancel();
    _pollRetryTimer = Timer(const Duration(milliseconds: 400), _pollNext);
  }

  Future<void> _retryPendingQueue() async {
    final conv = _conversation;
    if (conv == null) return;
    final pending = await _queue.forConversation(conv.id);
    for (final p in pending) {
      try {
        final msg = await _repo.sendMessage(conv.id, text: p.text, quickReplyKey: p.quickReplyKey);
        await _queue.remove(p.localId);
        _messages.removeWhere((m) => m.id == p.localId);
        _mergeMessages([msg]);
      } on ApiException catch (e) {
        if (e.statusCode != 0) {
          // A genuine rejection (e.g. conversation closed meanwhile) — drop
          // it rather than retry forever.
          await _queue.remove(p.localId);
          if (mounted) setState(() => _messages.removeWhere((m) => m.id == p.localId));
        }
      } catch (_) {
        // Still offline — leave it queued for the next retry.
      }
    }
  }

  Future<void> _send({String text = '', String quickReplyKey = ''}) async {
    final conv = _conversation;
    final trimmed = text.trim();
    if (conv == null || _sending) return;
    if (trimmed.isEmpty && quickReplyKey.isEmpty) return;
    setState(() => _sending = true);
    try {
      final msg = await _repo.sendMessage(conv.id, text: trimmed, quickReplyKey: quickReplyKey);
      _mergeMessages([msg]);
      _textController.clear();
    } on ApiException catch (e) {
      if (e.statusCode == 0) {
        await _queueOffline(conv.id, trimmed, quickReplyKey);
      } else if (mounted) {
        AppSnackbar.error(context, 'Không thể gửi tin nhắn. Vui lòng thử lại.');
      }
    } catch (_) {
      if (mounted) AppSnackbar.error(context, 'Không thể gửi tin nhắn.');
    } finally {
      if (mounted) setState(() => _sending = false);
    }
  }

  Future<void> _queueOffline(String conversationId, String text, String quickReplyKey) async {
    final localId = 'local-${DateTime.now().microsecondsSinceEpoch}';
    await _queue.enqueue(PendingMessage(
      localId: localId,
      conversationId: conversationId,
      text: text,
      quickReplyKey: quickReplyKey,
      createdAt: DateTime.now(),
    ));
    _messages.add(ChatMessage(
      id: localId,
      seq: _lastSeq,
      conversationId: conversationId,
      senderId: 'me',
      senderRole: _mySenderRole,
      body: quickReplyKey.isNotEmpty ? (quickReplyLabels[quickReplyKey] ?? text) : text,
      quickReplyKey: quickReplyKey,
      createdAt: DateTime.now(),
      pending: true,
    ));
    _textController.clear();
    if (mounted) {
      setState(() {});
      AppSnackbar.warning(context, 'Mất kết nối — tin nhắn sẽ tự gửi lại khi có mạng.');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Trò chuyện')),
      body: SafeArea(child: _buildBody()),
    );
  }

  Widget _buildBody() {
    if (_loading) {
      return ListView.separated(
        padding: const EdgeInsets.all(AppSpacing.lg),
        itemCount: 5,
        separatorBuilder: (_, _) => const SizedBox(height: AppSpacing.sm),
        itemBuilder: (context, i) => const AppSkeletonListTile(),
      );
    }
    if (_conversation == null) {
      return AppEmptyState.error(
        subtitle: _initError ?? 'Không thể mở cuộc trò chuyện.',
        onAction: () {
          setState(() => _loading = true);
          _init();
        },
      );
    }

    final conv = _conversation!;
    final lastMineIndex = _lastReadableMineIndex();

    return Column(
      children: [
        Expanded(
          child: _messages.isEmpty
              ? const AppEmptyState(
                  icon: Icons.chat_bubble_outline,
                  title: 'Chưa có tin nhắn',
                  subtitle: 'Gửi lời chào hoặc chọn một trả lời nhanh bên dưới.',
                )
              : ListView.builder(
                  controller: _scrollController,
                  padding: const EdgeInsets.all(AppSpacing.md),
                  itemCount: _messages.length,
                  itemBuilder: (context, i) {
                    final m = _messages[i];
                    final isMine = m.senderRole == _mySenderRole;
                    return Padding(
                      padding: const EdgeInsets.symmetric(vertical: 2),
                      child: MessageBubble(
                        message: m,
                        isMine: isMine,
                        showRead: isMine && i == lastMineIndex && m.isRead,
                      ),
                    );
                  },
                ),
        ),
        if (conv.isOpen) ...[
          QuickReplyRow(tripType: conv.tripType, onSelect: (key) => _send(quickReplyKey: key)),
          _buildInputRow(),
        ] else
          const Padding(
            padding: EdgeInsets.all(AppSpacing.lg),
            child: Text(
              'Chuyến đi đã kết thúc — cuộc trò chuyện đã đóng.',
              textAlign: TextAlign.center,
              style: TextStyle(color: AppColors.textTertiary, fontSize: 13),
            ),
          ),
      ],
    );
  }

  int _lastReadableMineIndex() {
    for (var i = _messages.length - 1; i >= 0; i--) {
      if (_messages[i].senderRole == _mySenderRole) return i;
    }
    return -1;
  }

  Widget _buildInputRow() {
    return SafeArea(
      top: false,
      child: Padding(
        padding: const EdgeInsets.fromLTRB(AppSpacing.md, 0, AppSpacing.md, AppSpacing.md),
        child: Row(
          children: [
            Expanded(
              child: TextField(
                controller: _textController,
                minLines: 1,
                maxLines: 4,
                textInputAction: TextInputAction.send,
                onSubmitted: (v) => _send(text: v),
                decoration: const InputDecoration(
                  hintText: 'Nhập tin nhắn…',
                  border: OutlineInputBorder(borderRadius: BorderRadius.all(Radius.circular(24))),
                  contentPadding: EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                ),
              ),
            ),
            const SizedBox(width: AppSpacing.sm),
            Semantics(
              button: true,
              label: 'Gửi tin nhắn',
              child: Tooltip(
                message: 'Gửi',
                child: Material(
                  color: AppColors.primary,
                  shape: const CircleBorder(),
                  child: InkWell(
                    customBorder: const CircleBorder(),
                    onTap: _sending ? null : () => _send(text: _textController.text),
                    child: Padding(
                      padding: const EdgeInsets.all(12),
                      child: _sending
                          ? const SizedBox(
                              width: 18,
                              height: 18,
                              child: CircularProgressIndicator(strokeWidth: 2, color: AppColors.textOnPrimary),
                            )
                          : const Icon(Icons.send, size: 18, color: AppColors.textOnPrimary),
                    ),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
