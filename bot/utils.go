package bot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
	"io/ioutil"
	"net/http"
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

func sendRequest(message *send.SendType) ([]byte, error) {
	if message == nil {
		return nil, errors.New("Message is nil. Nothing to send.")
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
		return nil, err
	}
	replyBody, err := ioutil.ReadAll(reply.Body)
	if reply.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Responded bad status: %s. Body: %s", reply.Status, string(replyBody))
	}
	return replyBody, err
}

func poll(message *send.SendType) (*receive.UpdateResultType, error) {
	response, err := sendRequest(message)
	if err != nil {
		return nil, err
	}
	var updates receive.UpdateResultType
	err = json.Unmarshal(response, &updates)
	if err != nil {
		return nil, err
	}
	return &updates, nil
}

type hackSendResult struct {
  Ok          bool        `json:"ok"`
  ErrorCode   int64       `json:"error_code,omitempty"`
  Description *string      `json:"description,omitempty"`
  Result      interface{} `json:"result,omitempty"`
}

func convertHackSendToSendResult(res *hackSendResult) (*receive.SendResult, error) {
  var ok bool
  _, ok = res.Result.(bool)
  if ok {
    return &receive.SendResult{
      Ok: true,
    }, nil
  }
  data, err := json.Marshal(res.Result)
  if err != nil {
    return nil, err
  }
  msg := receive.MessageType{}
  err = json.Unmarshal(data, &msg)
  if err != nil {
    return nil, err
  }
  return &receive.SendResult{
    Ok: res.Ok,
    ErrorCode: res.ErrorCode,
    Description: res.Description,
    Result: &msg,
  }, nil
}

func sendSingleResponse(message *send.SendType) (*receive.SendResult, error) {
	response, err := sendRequest(message)
	if err != nil {
		return nil, err
	}
  var hackSend hackSendResult
	err = json.Unmarshal(response, &hackSend)
	if err != nil {
		/*if message.Type == send.SEND_TYPE_GET {
		  err = nil
		}*/
		return nil, err
	}
	return convertHackSendToSendResult(&hackSend)
}

func sendResponse(messages []*send.SendType) ([]*send.SendResultWithCallback, error) {
	results := make([]*send.SendResultWithCallback, 0, len(messages))
	for i := range messages {
		res, err := sendSingleResponse(messages[i])
		results = append(results, &send.SendResultWithCallback{
			Result:   res,
			Error:    err,
			Callback: messages[i].Callback,
		})
	}
	return results, nil
}
