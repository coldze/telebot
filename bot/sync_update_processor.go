package bot

import (
	"github.com/coldze/telebot"
	"github.com/coldze/telebot/receive"
)

type SyncUpdateProcessor struct {
	logger   telebot.Logger
	onUpdate UpdateCallback
}

func (u *SyncUpdateProcessor) Process(update *receive.UpdateType) error {
	response, err := u.onUpdate(update)
	if err != nil {
		return err
	}
	if response == nil {
		return nil //errors.New("Reponse is empty")
	}
	responseSentResult, err := sendResponse(response)
	if err != nil {
		return err
	}
	if response.Callback != nil {
		response.Callback(responseSentResult)
	}
	return nil
}
