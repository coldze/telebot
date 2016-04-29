package send

const (
	SEND_TYPE_POST = iota + 1
	SEND_TYPE_GET
)

type SendType struct {
	URL         string
	Type        int64
	Parameters  []byte
	ContentType string
}

type RequestSender interface {
	Send(request *SendType) error
}
