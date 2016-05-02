package markup

type ReplyKeyboardHideType struct {
	HideKeyboard bool `json:"hide_keyboard"`
	Selective    bool `json:"selective,omitempty"`
}
