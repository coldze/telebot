package send

import (
	"github.com/coldze/primitives/custom_error"
	"github.com/coldze/telebot/receive"
)

const (
	SEND_TYPE_POST = iota + 1
	SEND_TYPE_GET
)

type OnSentCallback func(result *receive.SendResult, error custom_error.CustomError)

type SendType struct {
	URL         string
	Type        int64
	Parameters  []byte
	ContentType string
	Callback    OnSentCallback
}

type SendResultWithCallback struct {
	Result   *receive.SendResult
	Callback OnSentCallback
	Error    custom_error.CustomError
}

type RequestSender interface {
	Send(request *SendType) error
}
