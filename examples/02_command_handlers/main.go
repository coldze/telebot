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
			res, err := requestFactory.NewSendMessage(fmt.Sprintf("%d", command.MetaInfo.Message.Chat.ID), "I can't find arguments for command.", 0, false, false, 0, nil)
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
		res, err := requestFactory.NewSendMessage(fmt.Sprintf("%d", command.MetaInfo.Message.Chat.ID), "Will remember that :)", 0, false, false, 0, nil)
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
		res, err := requestFactory.NewSendMessage(fmt.Sprintf("%d", command.MetaInfo.Message.Chat.ID), message, 0, false, false, 0, nil)
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
			result, err = requestFactory.NewSendSticker(fmt.Sprintf("%v", update.Message.Chat.ID), STICKER_ID, false, 0, nil)
		} else {
			result, err = requestFactory.NewSendMessage(fmt.Sprintf("%v", update.Message.Chat.ID), "*ECHO:*\n"+update.Message.Text, send.PARSE_MODE_MARKDOWN, false, false, 0, nil)
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
