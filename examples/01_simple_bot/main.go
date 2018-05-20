package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/coldze/primitives/logs"
	"github.com/coldze/telebot/bot"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
)

const (
	BOT_TOKEN_KEY = "BOT_TOKEN"
)

func main() {
	logger := logs.NewStdoutLogger()
	botToken, ok := os.LookupEnv(BOT_TOKEN_KEY)
	if !ok {
		logger.Errorf("Failed to get bot-token. Expected to have environment variable '%s'.", BOT_TOKEN_KEY)
		return
	}
	factory := send.NewRequestFactory(botToken, logger)
	logger.Infof("Available bot functionality:\n%v", factory)
	logger.Infof("Request factory intialized.")
	onUpdate := func(update *receive.UpdateType) (*send.SendType, error) {
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
		var request *send.SendType
		if update.Message.Sticker != nil {
			request, err = factory.NewSendSticker(fmt.Sprintf("%v", update.Message.Chat.ID), "BQADAgADQAADyIsGAAGMQCvHaYLU_AI", false, 0, nil)
		} else {
			request, err = factory.NewSendMessage(fmt.Sprintf("%v", update.Message.Chat.ID), "*ECHO:*\n"+update.Message.Text, send.PARSE_MODE_MARKDOWN, false, false, 0, nil)
		}
		if err != nil {
			return nil, err
		}
		logger.Debugf("Response: %v.", string(request.Parameters))
		return request, nil
	}
	bot := bot.NewPollingBot(factory, onUpdate, 1000, logger)
	go func() {
		time.Sleep(10 * time.Second)
		msg, err := factory.NewSendMessage("-1001121852273", "TEST MESSAGE", send.PARSE_MODE_MARKDOWN, false, false, 0, nil)
		if err != nil {
			return
		}
		err = bot.Send(msg)
		if err != nil {
			return
		}
	}()
	defer bot.Stop()
	logger.Infof("Bot started. Press Enter to stop.")
	_, _ = fmt.Scanf("\n")
}
