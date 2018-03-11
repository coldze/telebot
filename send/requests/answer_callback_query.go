package requests

type AnswerCallbackQuery struct {
	CallbackQueryID interface{} `json:"callback_query_id"`
	Text            string      `json:"text,omitempty"`
	ShowAlert       bool        `json:"show_alert,omitempty"`
	URL             string      `json:"url,omitempty"`
	CacheTime       int64       `json:"cache_time,omitempty"`
}
