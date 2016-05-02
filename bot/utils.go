package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
	"io/ioutil"
	"net/http"
)

func sendRequest(message *send.SendType) ([]byte, error) {
	if message == nil {
		return nil, fmt.Errorf("Message is nil. Nothing to send.")
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

func sendResponse(message *send.SendType) (*receive.SendResult, error) {
	response, err := sendRequest(message)
	if err != nil {
		return nil, err
	}
	var sendResult receive.SendResult
	err = json.Unmarshal(response, &sendResult)
	if err != nil {
		/*if message.Type == send.SEND_TYPE_GET {
			err = nil
		}*/
		return nil, err
	}
	return &sendResult, nil
}
