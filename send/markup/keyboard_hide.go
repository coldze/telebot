package markup

type ReplyKeyboardHideType struct {
	Keyboard     [][]KeyboardButtonType `json:"keyboard"`
	Resize       bool                   `json:"resize_keyboard"`
	HideKeyboard bool                   `json:"hide_keyboard"`
	Selective    bool                   `json:"selective,omitempty"`
}
