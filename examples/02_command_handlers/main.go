package main

import (
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
	"fmt"
	"github.com/coldze/telebot"
	"strings"
	"errors"
	"bytes"
	"os"
)

const (
	BOT_TOKEN_KEY = "BOT_TOKEN"
)

type UsersMemory struct {
	Memorized map[int64][]string
}

func NewUsersMemory() *UsersMemory{
	return &UsersMemory{Memorized: make(map[int64][]string)}
}

func NewOnRememberCommand(users *UsersMemory, requestFactory *send.RequestFactory, logger telebot.Logger) (telebot.CommandHandler, error) {
	if users == nil {
		return nil, nil
	}
	if requestFactory == nil {
		return nil, nil
	}
	return func(command *telebot.CommandCallType)(*send.SendType, error) {
		logger.Infof("Remember command handler invoked.")
		if command.MetaInfo.Message.From == nil {
			return nil, errors.New("FROM missing")
		}
		if command.MetaInfo.Message.Chat == nil {
			return nil, errors.New("CHAT missing")
		}
		args := strings.TrimSpace(command.Argument)
		if len(args) <= 0 {
			return requestFactory.NewSendMessage(fmt.Sprintf("%d", command.MetaInfo.Message.Chat.ID), "I can't find arguments for command.", 0, false, false, 0, nil)
		}
		_, ok := users.Memorized[command.MetaInfo.Message.From.ID]
		if !ok {
			users.Memorized[command.MetaInfo.Message.From.ID] = []string{args}
		} else {
			users.Memorized[command.MetaInfo.Message.From.ID] = append(users.Memorized[command.MetaInfo.Message.From.ID], args)
		}
		return requestFactory.NewSendMessage(fmt.Sprintf("%d", command.MetaInfo.Message.Chat.ID), "Will remember that :)", 0, false, false, 0, nil)
	}, nil
}

func NewOnListCommand(users *UsersMemory, requestFactory *send.RequestFactory, logger telebot.Logger) (telebot.CommandHandler, error) {
	if users == nil {
		return nil, nil
	}
	if requestFactory == nil {
		return nil, nil
	}
	return func(command *telebot.CommandCallType)(*send.SendType, error) {
		logger.Infof("List command handler invoked.")
		if command.MetaInfo.Message.From == nil {
			return nil, errors.New("FROM missing")
		}
		if command.MetaInfo.Message.Chat == nil {
			return nil, errors.New("CHAT missing")
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
		return requestFactory.NewSendMessage(fmt.Sprintf("%d", command.MetaInfo.Message.Chat.ID), message, 0, false, false, 0, nil)
	}, nil
}

func main() {
	logger := telebot.NewStdoutLogger()
	botToken, ok := os.LookupEnv(BOT_TOKEN_KEY)
	if !ok {
		logger.Errorf("Failed to get bot-token. Expected to have environment variable '%s'.", botToken)
		return
	}
	requestFactory := send.NewRequestFactory(botToken)
	logger.Infof("Available bot functionality:\n%v", requestFactory)
	logger.Infof("Request factory intialized.")
	onMessage := func (update *receive.UpdateType) (result *send.SendType, err error) {
		if update.Message.Sticker != nil {
			result, err = requestFactory.NewSendSticker(fmt.Sprintf("%v", update.Message.Chat.ID), "BQADAgADQAADyIsGAAGMQCvHaYLU_AI", false, 0, nil)
		} else {
			result, err = requestFactory.NewSendMessage(fmt.Sprintf("%v", update.Message.Chat.ID), "*ECHO:*\n" + update.Message.Text, send.PARSE_MODE_MARKDOWN, false, false, 0, nil)
		}
		logger.Debugf("Response: %v.", string(result.Parameters))
		return
	}
	registry := telebot.NewBotHandlers(onMessage)
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
	onUpdate, err := telebot.NewDefaultUpdateCallback(logger, registry)
	if err != nil {
		logger.Errorf("Failed to create default update-callback. Error: %v.", err)
		return
	}

	bot := telebot.NewPollingBot(requestFactory, onUpdate, 1000, logger)
	defer bot.Stop()
	logger.Infof("Bot started. Press Enter to stop.")
	_, _ = fmt.Scanf("\n")
}
