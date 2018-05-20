package requests

type SendSticker struct {
	Base
	Sticker string `json:"sticker"`
}
