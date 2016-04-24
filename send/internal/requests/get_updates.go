package send_requests

type GetUpdatesType struct {
	Offset  int64 `json:"offset,omitempty"`
	Limit   int64 `json:"limit,omitempty"`
	Timeout int64 `json:"timeout,omitempty"`
}
