package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/coldze/telebot"
	"github.com/coldze/telebot/bot"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
	"github.com/coldze/telebot/send/markup"
	"os"
	"strconv"
	"strings"
)

const (
	BOT_TOKEN_KEY               = "BOT_TOKEN"
	BOT_SSL_PUBLIC_KEY          = "BOT_SSL_PUBLIC"
	BOT_SSL_PRIVATE_KEY         = "BOT_SSL_PRIVATE"
	BOT_SSL_SELF_SIGNED         = "BOT_SSL_SELF_SIGNED"
	BOT_UPDATE_CALLBACK_URL_KEY = "BOT_UPDATE_CALLBACK_URL"
	BOT_HTTPS_LISTEN_PORT_KEY   = "BOT_HTTPS_LISTEN_PORT"
)

type UsersMemory struct {
	Memorized map[int64][]string
}

func NewUsersMemory() *UsersMemory {
	return &UsersMemory{Memorized: make(map[int64][]string)}
}

func NewOnRememberCommand(users *UsersMemory, requestFactory *send.RequestFactory, logger telebot.Logger) (bot.CommandHandler, error) {
	if users == nil {
		return nil, nil
	}
	if requestFactory == nil {
		return nil, nil
	}
	return func(command *bot.CommandCallType) (*send.SendType, error) {
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

func NewOnListCommand(users *UsersMemory, requestFactory *send.RequestFactory, logger telebot.Logger) (bot.CommandHandler, error) {
	if users == nil {
		return nil, nil
	}
	if requestFactory == nil {
		return nil, nil
	}
	return func(command *bot.CommandCallType) (*send.SendType, error) {
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

func NewOnStartCommand(requestFactory *send.RequestFactory, logger telebot.Logger) (bot.CommandHandler, error) {
	return func(command *bot.CommandCallType) (*send.SendType, error) {
		logger.Infof("Start command handler invoked.")
		if command.MetaInfo.Message.From == nil {
			return nil, errors.New("FROM missing")
		}
		if command.MetaInfo.Message.Chat == nil {
			return nil, errors.New("CHAT missing")
		}

		logger.Infof("From: %+v. Chat: %+v. Argument: %+v", command.MetaInfo.Message.From, command.MetaInfo.Message.Chat, command.Argument)

		/*var message string
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
		  return requestFactory.NewSendMessage(fmt.Sprintf("%d", command.MetaInfo.Message.Chat.ID), message, 0, false, false, 0, nil)*/

		var inlineKeyboardMarkup markup.InlineKeyboardMarkupType
		inlineKeyboardMarkup.Buttons = make([][]markup.InlineKeyboardButtonType, 1)
		inlineKeyboardMarkup.Buttons[0] = make([]markup.InlineKeyboardButtonType, 2)
		inlineKeyboardMarkup.Buttons[0][0].Text = "Start training"
		inlineKeyboardMarkup.Buttons[0][0].CallbackData = "/start_training"
		inlineKeyboardMarkup.Buttons[0][1].Text = "Don't start"
		inlineKeyboardMarkup.Buttons[0][1].CallbackData = "/go_now"
		replyKeyboard := markup.ReplyKeyboardMarkupType{}
		replyKeyboard.OneTimeKeyboard = true
		replyKeyboard.Keyboard = make([][]markup.KeyboardButtonType, 1)
		replyKeyboard.Keyboard[0] = make([]markup.KeyboardButtonType, 9)
		replyKeyboard.Keyboard[0][0].Text = "1"
		replyKeyboard.Keyboard[0][1].Text = "2"
		replyKeyboard.Keyboard[0][2].Text = "3"
		replyKeyboard.Keyboard[0][3].Text = "4"
		replyKeyboard.Keyboard[0][4].Text = "5"
		replyKeyboard.Keyboard[0][5].Text = "6"
		replyKeyboard.Keyboard[0][6].Text = "7"
		replyKeyboard.Keyboard[0][7].Text = "8"
		replyKeyboard.Keyboard[0][8].Text = "9"

		return requestFactory.NewSendMessage(fmt.Sprintf("%d", command.MetaInfo.Message.Chat.ID), "Choose:", 0, false, false, 0, replyKeyboard)

		//return requestFactory.NewSendMessage("@cold3e", "TEST directly", 0, false, false, 0, nil)
		//return nil, nil
	}, nil
}

func main() {
	logger := telebot.NewStdoutLogger()
	botToken, ok := os.LookupEnv(BOT_TOKEN_KEY)
	if !ok {
		logger.Errorf("Failed to get bot-token. Expected to have environment variable '%s'.", BOT_TOKEN_KEY)
		return
	}

	sslPublic, ok := os.LookupEnv(BOT_SSL_PUBLIC_KEY)
	if !ok {
		logger.Errorf("Failed to get ssl public key. Expected to have environment variable '%s'.", BOT_SSL_PUBLIC_KEY)
		return
	}

	sslPrivate, ok := os.LookupEnv(BOT_SSL_PRIVATE_KEY)
	if !ok {
		logger.Errorf("Failed to get ssl private key. Expected to have environment variable '%s'.", BOT_SSL_PRIVATE_KEY)
		return
	}

	updateCallbackURL, ok := os.LookupEnv(BOT_UPDATE_CALLBACK_URL_KEY)
	if !ok {
		logger.Errorf("Failed to get update callback URL. Expected to have environment variable '%s'.", BOT_UPDATE_CALLBACK_URL_KEY)
		return
	}

	listenPortStr, ok := os.LookupEnv(BOT_HTTPS_LISTEN_PORT_KEY)
	if !ok {
		logger.Errorf("Failed to get listening port. Expected to have environment variable '%s'.", BOT_HTTPS_LISTEN_PORT_KEY)
		return
	}

	isSelfSigned := false
	isSelfSignedStr, ok := os.LookupEnv(BOT_SSL_SELF_SIGNED)
	if ok {
		var err error
		isSelfSigned, err = strconv.ParseBool(isSelfSignedStr)
		if err != nil {
			logger.Errorf("Failed to parse self-signed env variable: '%s' = '%s'. Error: %v.", BOT_SSL_SELF_SIGNED, isSelfSignedStr, err)
			return
		}
	}

	listenPort, err := strconv.ParseInt(listenPortStr, 10, 64)
	if err != nil {
		logger.Errorf("Failed to get listening port. Error: %v.", err)
		return
	}

	requestFactory := send.NewRequestFactory(botToken, logger)
	logger.Infof("Available bot functionality:\n%v", requestFactory)
	logger.Infof("Request factory intialized.")
	onMessage := func(update *receive.UpdateType) (result *send.SendType, err error) {
		if update.Message.Sticker != nil {
			result, err = requestFactory.NewSendSticker(fmt.Sprintf("%v", update.Message.Chat.ID), "BQADAgADQAADyIsGAAGMQCvHaYLU_AI", false, 0, nil)
		} else {
			result, err = requestFactory.NewSendMessage(fmt.Sprintf("%v", update.Message.Chat.ID), "*ECHO:*\n"+update.Message.Text, send.PARSE_MODE_MARKDOWN, false, false, 0, nil)
		}
		logger.Debugf("Response: %v.", string(result.Parameters))
		return
	}
	registry := bot.NewBotHandlers(onMessage)

	onStart, err := NewOnStartCommand(requestFactory, logger)
	if err != nil {
		logger.Errorf("Failed to create on-start handler. Error: %v.", err)
		return
	}
	err = registry.RegisterCommand("/start", onStart)

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
	onUpdate, err := bot.NewDefaultUpdateCallback(logger, registry)
	if err != nil {
		logger.Errorf("Failed to create default update-callback. Error: %v.", err)
		return
	}

	_, err = bot.NewWebHookBot(requestFactory, onUpdate, updateCallbackURL, listenPort, sslPrivate, sslPublic, isSelfSigned, logger)
	if err != nil {
		logger.Errorf("Failed to start bot. Error: %v.", err)
		return
	}
}
