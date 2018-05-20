package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/coldze/primitives/custom_error"
	"github.com/coldze/primitives/logs"
	"github.com/coldze/telebot/bot"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
	"github.com/coldze/telebot/send/requests"
)

const (
	BOT_TOKEN_KEY = "BOT_TOKEN"
	STICKER_ID    = "BQADAgADQAADyIsGAAGMQCvHaYLU_AI"
)

type UsersMemory struct {
	Memorized map[int64][]string
}

func NewUsersMemory() *UsersMemory {
	return &UsersMemory{Memorized: make(map[int64][]string)}
}

func NewOnRememberCommand(users *UsersMemory, requestFactory *send.RequestFactory, logger logs.Logger) (bot.CommandHandler, custom_error.CustomError) {
	if users == nil {
		return nil, nil
	}
	if requestFactory == nil {
		return nil, nil
	}
	return func(command *bot.CommandCallType) ([]*send.SendType, custom_error.CustomError) {
		logger.Infof("Remember command handler invoked.")
		if command.MetaInfo.Message.From == nil {
			return nil, custom_error.MakeErrorf("FROM missing")
		}
		if command.MetaInfo.Message.Chat == nil {
			return nil, custom_error.MakeErrorf("CHAT missing")
		}
		args := strings.TrimSpace(command.Argument)
		if len(args) <= 0 {
			sendMessage := &requests.SendMessage{
				Base: requests.Base{
					ChatID: command.MetaInfo.Message.Chat.ID,
				},
				Text: "I can't find arguments for command.",
			}
			res, err := requestFactory.NewSendMessage(sendMessage, nil)
			if err == nil {
				return res, nil
			}
			return nil, custom_error.NewErrorf(err, "Failed to send message.")
		}
		_, ok := users.Memorized[command.MetaInfo.Message.From.ID]
		if !ok {
			users.Memorized[command.MetaInfo.Message.From.ID] = []string{args}
		} else {
			users.Memorized[command.MetaInfo.Message.From.ID] = append(users.Memorized[command.MetaInfo.Message.From.ID], args)
		}
		sendMessage := &requests.SendMessage{
			Base: requests.Base{
				ChatID: command.MetaInfo.Message.Chat.ID,
			},
			Text: "Will remember that :)",
		}
		res, err := requestFactory.NewSendMessage(sendMessage, nil)
		if err == nil {
			return res, nil
		}
		return nil, custom_error.NewErrorf(err, "Failed to send message.")
	}, nil
}

func NewOnListCommand(users *UsersMemory, requestFactory *send.RequestFactory, logger logs.Logger) (bot.CommandHandler, custom_error.CustomError) {
	if users == nil {
		return nil, nil
	}
	if requestFactory == nil {
		return nil, nil
	}
	return func(command *bot.CommandCallType) ([]*send.SendType, custom_error.CustomError) {
		logger.Infof("List command handler invoked.")
		if command.MetaInfo.Message.From == nil {
			return nil, custom_error.MakeErrorf("FROM missing")
		}
		if command.MetaInfo.Message.Chat == nil {
			return nil, custom_error.MakeErrorf("CHAT missing")
		}

		memory, ok := users.Memorized[command.MetaInfo.Message.From.ID]
		var message string
		if !ok {
			message = "I have no history for you, sorry :("
		} else {
			var buffer bytes.Buffer
			for i := range memory {
				buffer.WriteString(memory[i])
				buffer.WriteString("\n")
			}
			message = buffer.String()
		}
		sendMessage := &requests.SendMessage{
			Base: requests.Base{
				ChatID: command.MetaInfo.Message.Chat.ID,
			},
			Text: message,
		}
		res, err := requestFactory.NewSendMessage(sendMessage, nil)
		if err == nil {
			return res, nil
		}
		return nil, custom_error.NewErrorf(err, "Failed to send message.")
	}, nil
}

func main() {
	logger := logs.NewStdLogger()
	botToken, ok := os.LookupEnv(BOT_TOKEN_KEY)
	if !ok {
		logger.Errorf("Failed to get bot-token. Expected to have environment variable '%s'.", BOT_TOKEN_KEY)
		return
	}
	requestFactory := send.NewRequestFactory(botToken, logger)
	logger.Infof("Available bot functionality:\n%v", requestFactory)
	logger.Infof("Request factory intialized.")
	onMessage := func(update *receive.UpdateType) ([]*send.SendType, custom_error.CustomError) {
		var result []*send.SendType
		var err custom_error.CustomError
		if update.Message.Sticker != nil {
			sendSticker := &requests.SendSticker{
				Base: requests.Base{
					ChatID: update.Message.Chat.ID,
				},
				Sticker: STICKER_ID,
			}
			result, err = requestFactory.NewSendSticker(sendSticker, nil)
		} else {
			sendMessage := &requests.SendMessage{
				Base: requests.Base{
					ChatID: update.Message.Chat.ID,
				},
				Text:      "*ECHO:*\n" + update.Message.Text,
				ParseMode: send.PARSE_MODE_MARKDOWN,
			}
			result, err = requestFactory.NewSendMessage(sendMessage, nil)
		}
		if err != nil {
			logger.Errorf("Failed to process message. Error: %v", custom_error.NewErrorf(err, "Failed to process message."))
			return nil, custom_error.NewErrorf(err, "Failed to process message.")
		}
		if result == nil {
			logger.Warningf("Processing result is empty.")
			return nil, nil
		}
		for i := range result {
			logger.Debugf("Response[%v]: %v.", i, string(result[i].Parameters))
		}
		return result, nil
	}
	registry := bot.NewBotHandlers(onMessage)
	usersMemory := NewUsersMemory()
	onRem, err := NewOnRememberCommand(usersMemory, requestFactory, logger)
	if err != nil {
		logger.Errorf("Failed to create on-rem handler. Error: %v.", err)
		return
	}
	err = registry.RegisterCommand("/rem", onRem)
	onEcho, err := NewOnListCommand(usersMemory, requestFactory, logger)
	if err != nil {
		logger.Errorf("Failed to create on-echo handler. Error: %v.", err)
		return
	}
	err = registry.RegisterCommand("/list", onEcho)
	if err != nil {
		logger.Errorf("Initialization failed. Error: %v.", err)
		return
	}
	onUpdate, err := bot.NewDefaultUpdateCallback(requestFactory, logger, registry)
	if err != nil {
		logger.Errorf("Failed to create default update-callback. Error: %v.", err)
		return
	}

	botApp := bot.NewPollingBot(requestFactory, onUpdate, 1000, logger)
	defer botApp.Stop()
	logger.Infof("Bot started. Press Enter to stop.")
	_, _ = fmt.Scanf("\n")
}
