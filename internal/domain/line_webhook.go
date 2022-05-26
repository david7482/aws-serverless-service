package domain

type LineWebhook struct {
	ExternalChannelID string
	Signature         string
	Payload           []byte
}

type LineEventType string

const (
	LineEventTypeMessage  = LineEventType("message")
	LineEventTypeFollow   = LineEventType("follow")
	LineEventTypeUnfollow = LineEventType("unfollow")
)

type LineEvent struct {
	ExternalMemberID string
	EventType        LineEventType
	ReplyToken       string
	EventContent     []byte
}
