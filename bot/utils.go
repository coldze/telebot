package bot

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/coldze/primitives/custom_error"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
)

func GetMessage(update *receive.UpdateType) *receive.MessageType {
	if update.Message != nil {
		return update.Message
	}
	if update.EditedMessage != nil {
		return update.EditedMessage
	}
	if update.ChannelPost != nil {
		return update.ChannelPost
	}
	if update.EditedChannelPost != nil {
		return update.EditedChannelPost
	}
	if update.CallbackQuery != nil {
		if update.CallbackQuery.Message != nil {
			return update.CallbackQuery.Message
		}
	}
	return nil
}

func sendRequest(message *send.SendType) ([]byte, custom_error.CustomError) {
	if message == nil {
		return nil, custom_error.MakeErrorf("Message is nil. Nothing to send.")
	}

	var reply *http.Response
	var err error
	switch message.Type {
	case send.SEND_TYPE_POST:
		buffer := bytes.NewReader(message.Parameters)
		reply, err = http.Post(message.URL, message.ContentType, buffer)
	case send.SEND_TYPE_GET:
		reply, err = http.Get(message.URL)
	}
	if reply != nil {
		defer reply.Body.Close()
	}
	if err != nil {
		return nil, custom_error.MakeErrorf("Failed to send request. Error: %v", err)
	}
	replyBody, err := ioutil.ReadAll(reply.Body)
	if reply.StatusCode != http.StatusOK {
		return nil, custom_error.MakeErrorf("Responded bad status: %s. Body: %s", reply.Status, string(replyBody))
	}
	if err != nil {
		return replyBody, custom_error.MakeErrorf("Failed to read body. Error: %v", err)
	}
	return replyBody, nil
}

func poll(message *send.SendType) (*receive.UpdateResultType, custom_error.CustomError) {
	response, customErr := sendRequest(message)
	if customErr != nil {
		return nil, custom_error.NewErrorf(customErr, "Failed to execute poll.")
	}
	var updates receive.UpdateResultType
	err := json.Unmarshal(response, &updates)
	if err != nil {
		return nil, custom_error.MakeErrorf("Failed to unmarshal poll result. Error: %v", err)
	}
	return &updates, nil
}

type hackSendResult struct {
	Ok          bool        `json:"ok"`
	ErrorCode   int64       `json:"error_code,omitempty"`
	Description *string     `json:"description,omitempty"`
	Result      interface{} `json:"result,omitempty"`
}

func convertHackSendToSendResult(res *hackSendResult) (*receive.SendResult, custom_error.CustomError) {
	var ok bool
	_, ok = res.Result.(bool)
	if ok {
		return &receive.SendResult{
			Ok: true,
		}, nil
	}
	data, err := json.Marshal(res.Result)
	if err != nil {
		return nil, custom_error.MakeErrorf("Failed to marshal hack-send. Error: %v", err)
	}
	msg := receive.MessageType{}
	err = json.Unmarshal(data, &msg)
	if err != nil {
		return nil, custom_error.MakeErrorf("Failed to unmarshal daata to message-type. Error: %v", err)
	}
	return &receive.SendResult{
		Ok:          res.Ok,
		ErrorCode:   res.ErrorCode,
		Description: res.Description,
		Result:      &msg,
	}, nil
}

func sendSingleResponse(message *send.SendType) (*receive.SendResult, custom_error.CustomError) {
	response, customErr := sendRequest(message)
	if customErr != nil {
		return nil, custom_error.NewErrorf(customErr, "Failed to send request.")
	}
	var hackSend hackSendResult
	err := json.Unmarshal(response, &hackSend)
	if err != nil {
		/*if message.Type == send.SEND_TYPE_GET {
		  err = nil
		}*/
		return nil, custom_error.MakeErrorf("Failed to unmarshal response of send-request. Error: %v", err)
	}
	res, customErr := convertHackSendToSendResult(&hackSend)
	if customErr == nil {
		return res, nil
	}
	return nil, custom_error.NewErrorf(customErr, "Failed to convert to send-result")
}

func sendResponse(messages []*send.SendType) ([]*send.SendResultWithCallback, custom_error.CustomError) {
	results := make([]*send.SendResultWithCallback, 0, len(messages))
	for i := range messages {
		res, err := sendSingleResponse(messages[i])
		if err != nil {
			err = custom_error.NewErrorf(err, "Failed to send response")
		}
		results = append(results, &send.SendResultWithCallback{
			Result:   res,
			Error:    err,
			Callback: messages[i].Callback,
		})
	}
	return results, nil
}
