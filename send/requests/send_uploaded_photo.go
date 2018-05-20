package requests

type SendUploadedPhoto struct {
	Base
	Photo   string `json:"photo"`
	Caption string `json:"caption,omitempty"`
}
