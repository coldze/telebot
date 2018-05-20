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
	"github.com/coldze/telebot/send/requests"
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
			sendSticker := &requests.SendSticker{
				Base: requests.Base{
					ChatID: update.Message.Chat.ID,
				},
				Sticker: STICKER_ID,
			}
			request, customErr = factory.NewSendSticker(sendSticker, nil)
		} else {
			sendMessage := &requests.SendMessage{
				Base: requests.Base{
					ChatID: update.Message.Chat.ID,
				},
				Text:      "*ECHO:*\n" + update.Message.Text,
				ParseMode: send.PARSE_MODE_MARKDOWN,
			}
			request, customErr = factory.NewSendMessage(sendMessage, nil)
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
		sendMessage := &requests.SendMessage{
			Base: requests.Base{
				ChatID: CHAT_ID,
			},
			Text:      "TEST MESSAGE",
			ParseMode: send.PARSE_MODE_MARKDOWN,
		}
		msg, err := factory.NewSendMessage(sendMessage, nil)
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
