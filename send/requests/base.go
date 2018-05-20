package requests

type Base struct {
	ChatID               interface{} `json:"chat_id"`
	DisableNotifications bool        `json:"disable_notification,omitempty"`
	ReplyToMessageID     int64       `json:"reply_to_message_id"`
	ReplyMarkup          interface{} `json:"reply_markup,omitempty"`
}
