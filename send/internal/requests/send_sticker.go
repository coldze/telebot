package send_requests

type SendSticker struct {
	ChatID              string      `json:"chat_id"`
	Sticker             string      `json:"sticker"`
	DisableNotification bool        `json:"disable_notification,omitempty"`
	ReplyToMessageID    int64       `json:"reply_to_message_id"`
	ReplyMarkup         interface{} `json:"reply_markup,omitempty"`
}
