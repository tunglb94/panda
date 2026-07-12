/// Fixed quick-reply keys, mirrored exactly from the backend's
/// `notification/domain/entity/quick_reply.go` (`QuickReplyText`) — the
/// server resolves key -&gt; canonical Vietnamese text, this map is only used
/// to render the chip labels client-side. Ride conversations get the first
/// 4; Delivery conversations get all 6 (see [quickRepliesFor]).
const Map<String, String> quickReplyLabels = {
  'arrived': 'Tôi tới rồi',
  'wait_2min': 'Đợi tôi 2 phút',
  'on_the_way_down': 'Tôi đang xuống',
  'thanks': 'Cảm ơn',
  'picked_up': 'Tôi đã lấy hàng',
  'delivered': 'Tôi đã giao',
};

const List<String> _rideQuickReplyKeys = ['arrived', 'wait_2min', 'on_the_way_down', 'thanks'];
const List<String> _deliveryOnlyQuickReplyKeys = ['picked_up', 'delivered'];

/// Returns the quick-reply keys valid for a conversation's trip type, in
/// display order.
List<String> quickRepliesFor(String tripType) => tripType == 'delivery'
    ? [..._rideQuickReplyKeys, ..._deliveryOnlyQuickReplyKeys]
    : _rideQuickReplyKeys;
