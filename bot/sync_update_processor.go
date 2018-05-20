package bot

import (
	"github.com/coldze/primitives/custom_error"
	"github.com/coldze/primitives/logs"
	"github.com/coldze/telebot/receive"
)

type SyncUpdateProcessor struct {
	logger   logs.Logger
	onUpdate UpdateCallback
}

func (u *SyncUpdateProcessor) Process(update *receive.UpdateType) custom_error.CustomError {
	response, err := u.onUpdate(update)
	if err != nil {
		return custom_error.NewErrorf(err, "Failed to handle update")
	}
	if response == nil {
		return nil //errors.New("Reponse is empty")
	}
	responseSentResult, err := sendResponse(response)
	if err != nil {
		return custom_error.NewErrorf(err, "Failed to send response")
	}
	for i := range responseSentResult {
		res := responseSentResult[i]
		if res.Callback != nil {
			if res.Error != nil {
				res.Error = custom_error.NewErrorf(res.Error, "Failed to send response")
			}
			res.Callback(res.Result, res.Error)
		}
	}
	return nil
}

func NewUpdateProcessor(onUpdate UpdateCallback, logger logs.Logger) UpdateProcessor {
	return &SyncUpdateProcessor{
		logger:   logger,
		onUpdate: onUpdate,
	}
}
