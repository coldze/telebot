package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/coldze/primitives/custom_error"
	"github.com/coldze/primitives/logs"
	"github.com/coldze/telebot/bot"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
)

const (
	BOT_TOKEN_KEY = "BOT_TOKEN"
	CHAT_ID       = "-1001121852273"
	STICKER_ID    = "BQADAgADQAADyIsGAAGMQCvHaYLU_AI"
)

func main() {
	logger := logs.NewStdLogger()
	botToken, ok := os.LookupEnv(BOT_TOKEN_KEY)
	if !ok {
		logger.Errorf("Failed to get bot-token. Expected to have environment variable '%s'.", BOT_TOKEN_KEY)
		return
	}
	factory := send.NewRequestFactory(botToken, logger)
	logger.Infof("Available bot functionality:\n%v", factory)
	logger.Infof("Request factory intialized.")
	onUpdate := func(update *receive.UpdateType) ([]*send.SendType, custom_error.CustomError) {
		return nil, nil
		if update == nil {
			return nil, nil
		}
		if update.Message == nil {
			return nil, nil
		}
		dumpedUpdate, err := json.MarshalIndent(update, "", "    ")
		if err != nil {
			logger.Errorf("Failed to marshal update object. Error: %v.", err)
		} else {
			logger.Debugf("Received update:\n%v", string(dumpedUpdate))
		}
		var request []*send.SendType
		var customErr custom_error.CustomError
		if update.Message.Sticker != nil {
			request, customErr = factory.NewSendSticker(fmt.Sprintf("%v", update.Message.Chat.ID), STICKER_ID, false, 0, nil)
		} else {
			request, customErr = factory.NewSendMessage(fmt.Sprintf("%v", update.Message.Chat.ID), "*ECHO:*\n"+update.Message.Text, send.PARSE_MODE_MARKDOWN, false, false, 0, nil)
		}
		if customErr != nil {
			return nil, custom_error.NewErrorf(customErr, "Failed to process update.")
		}
		if request == nil {
			logger.Warningf("Processing result is empty.")
			return nil, nil
		}
		for i := range request {
			logger.Debugf("Response[%v]: %v.", i, string(request[i].Parameters))
		}
		return request, nil
	}
	botApp := bot.NewPollingBot(factory, onUpdate, 1000, logger)
	go func() {
		time.Sleep(10 * time.Second)
		msg, err := factory.NewSendMessage(CHAT_ID, "TEST MESSAGE", send.PARSE_MODE_MARKDOWN, false, false, 0, nil)
		if err != nil {
			return
		}
		err = botApp.Send(msg)
		if err != nil {
			return
		}
	}()
	defer botApp.Stop()
	logger.Infof("Bot started. Press Enter to stop.")
	_, _ = fmt.Scanf("\n")
}
