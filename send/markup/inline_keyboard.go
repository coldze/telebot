package markup

type InlineKeyboardButtonType struct {
	Text              string `json:"text"`
	URL               string `json:"url,omitempty"`
	CallbackData      string `json:"callback_data,omitempty"`
	SwitchInlineQuery string `json:"switch_inline_query"`
}

type InlineKeyboardMarkupType struct {
	Buttons [][]InlineKeyboardButtonType `json:"inline_keyboard"`
}
