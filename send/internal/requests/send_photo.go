package send_requests

type SendPhoto struct {
	ChatID              string      `json:"chat_id"`
	Photo               string      `json:"photo"`
	Caption             string      `json:"caption,omitempty"`
	DisableNotification bool        `json:"disable_notification,omitempty"`
	ReplyToMessageID    int64       `json:"reply_to_message_id"`
	ReplyMarkup         interface{} `json:"reply_markup,omitempty"`
}
