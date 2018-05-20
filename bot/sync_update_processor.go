package bot

import (
	"github.com/coldze/primitives/logs"
	"github.com/coldze/telebot/receive"
)

type SyncUpdateProcessor struct {
	logger   logs.Logger
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
	for i := range responseSentResult {
		res := responseSentResult[i]
		if res.Callback != nil {
			res.Callback(res.Result, res.Error)
		}
	}
	return nil
}
