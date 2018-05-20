package main

import (
	"bytes"
	"os"
	"strconv"
	"strings"

	"github.com/coldze/primitives/custom_error"
	"github.com/coldze/primitives/logs"
	"github.com/coldze/telebot/bot"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
	"github.com/coldze/telebot/send/markup"
	"github.com/coldze/telebot/send/requests"
)

const (
	BOT_TOKEN_KEY               = "BOT_TOKEN"
	BOT_SSL_PUBLIC_KEY          = "BOT_SSL_PUBLIC"
	BOT_SSL_PRIVATE_KEY         = "BOT_SSL_PRIVATE"
	BOT_SSL_SELF_SIGNED         = "BOT_SSL_SELF_SIGNED"
	BOT_UPDATE_CALLBACK_URL_KEY = "BOT_UPDATE_CALLBACK_URL"
	BOT_HTTPS_LISTEN_PORT_KEY   = "BOT_HTTPS_LISTEN_PORT"

	STICKER_ID = "BQADAgADQAADyIsGAAGMQCvHaYLU_AI"
)

type UsersMemory struct {
	Memorized map[int64][]string
}

func NewUsersMemory() *UsersMemory {
	return &UsersMemory{Memorized: make(map[int64][]string)}
}

func NewOnRememberCommand(users *UsersMemory, requestFactory *send.RequestFactory, logger logs.Logger) (bot.CommandHandler, error) {
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
			return nil, custom_error.NewErrorf(err, "Failed to create new send message.")
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
		return nil, custom_error.NewErrorf(err, "Failed to create new send message.")
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
		return nil, custom_error.NewErrorf(err, "Failed to create new send message.")
	}, nil
}

func NewOnStartCommand(requestFactory *send.RequestFactory, logger logs.Logger) (bot.CommandHandler, custom_error.CustomError) {
	return func(command *bot.CommandCallType) ([]*send.SendType, custom_error.CustomError) {
		logger.Infof("Start command handler invoked.")
		if command.MetaInfo.Message.From == nil {
			return nil, custom_error.MakeErrorf("FROM missing")
		}
		if command.MetaInfo.Message.Chat == nil {
			return nil, custom_error.MakeErrorf("CHAT missing")
		}

		logger.Infof("From: %+v. Chat: %+v. Argument: %+v", command.MetaInfo.Message.From, command.MetaInfo.Message.Chat, command.Argument)

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

		sendMessage := &requests.SendMessage{
			Base: requests.Base{
				ChatID:      command.MetaInfo.Message.Chat.ID,
				ReplyMarkup: replyKeyboard,
			},
			Text: "Choose:",
		}

		res, err := requestFactory.NewSendMessage(sendMessage, nil)
		if err == nil {
			return res, nil
		}
		return nil, custom_error.NewErrorf(err, "Failed to create new send message.")
	}, nil
}

func main() {
	logger := logs.NewStdLogger()
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
	onUpdate, err := bot.NewDefaultUpdateCallback(requestFactory, logger, registry)
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
