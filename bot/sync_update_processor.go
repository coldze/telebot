package bot

import (
	"errors"
	"fmt"
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
		return errors.New("Reponse is empty")
	}
	responseSentResult, err := sendResponse(response)
	if err != nil {
		return err
	}
	if responseSentResult.Ok {
		u.logger.Infof("Response sent.")
		return nil
	}
	return fmt.Errorf("Failed to send response for update id '%d'. Received error: code - '%d', description '%s'.", update.ID, responseSentResult.ErrorCode, responseSentResult.Description)
}
