package entity

// QuickReplyText maps a fixed quick-reply key to its canonical Vietnamese
// message text. The server resolves key -> text (SendMessageUseCase), so a
// client only ever transmits the key, never freeform text for a canned
// reply — both Flutter apps' quick-reply chip rows must use these same keys
// (see apps/*/lib/features/chat).
var QuickReplyText = map[string]string{
	"arrived":         "Tôi tới rồi",
	"wait_2min":       "Đợi tôi 2 phút",
	"on_the_way_down": "Tôi đang xuống",
	"thanks":          "Cảm ơn",
	"picked_up":       "Tôi đã lấy hàng",
	"delivered":       "Tôi đã giao",
}

// deliveryOnlyQuickReplyKeys are valid only when the conversation's TripType
// is "delivery" — kept separate so a Ride conversation rejects them with a
// clear error instead of silently accepting a delivery-flavored reply.
var deliveryOnlyQuickReplyKeys = map[string]bool{
	"picked_up": true,
	"delivered": true,
}

// IsQuickReplyValidForTripType reports whether key is a known quick-reply
// key usable for the given trip type ("ride" or "delivery").
func IsQuickReplyValidForTripType(key, tripType string) bool {
	if _, ok := QuickReplyText[key]; !ok {
		return false
	}
	if deliveryOnlyQuickReplyKeys[key] && tripType != "delivery" {
		return false
	}
	return true
}
