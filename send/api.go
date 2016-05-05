package send

import "github.com/coldze/telebot/receive"

const (
	SEND_TYPE_POST = iota + 1
	SEND_TYPE_GET
)

type OnSentCallback func(result *receive.SendResult)

type SendType struct {
	URL         string
	Type        int64
	Parameters  []byte
	ContentType string
	Callback    OnSentCallback
}

type RequestSender interface {
	Send(request *SendType) error
}
