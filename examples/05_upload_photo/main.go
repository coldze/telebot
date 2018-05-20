package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/coldze/primitives/logs"
	"github.com/coldze/telebot/bot"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
)

const (
	BOT_TOKEN_KEY      = "BOT_TOKEN"
	BOT_IMAGE_FILE_KEY = "BOT_IMAGE_FILE"
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

type testImageHolder struct {
	lock    sync.RWMutex
	imageID string
}

func (t *testImageHolder) GetImageID() string {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.imageID
}

func (t *testImageHolder) SetImageID(in string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.imageID = in
}

func NewOnTestCommand(users *UsersMemory, requestFactory *send.RequestFactory, imageFile string, logger logs.Logger) (bot.CommandHandler, error) {
	if users == nil {
		return nil, nil
	}
	if requestFactory == nil {
		return nil, nil
	}
	var imageHolder testImageHolder
	return func(command *bot.CommandCallType) (*send.SendType, error) {
		logger.Infof("Test command handler invoked.")
		if command.MetaInfo.Message.From == nil {
			return nil, errors.New("FROM missing")
		}
		if command.MetaInfo.Message.Chat == nil {
			return nil, errors.New("CHAT missing")
		}

		imgID := imageHolder.GetImageID()
		if len(imgID) > 0 {
			logger.Infof("Re-sending.")
			return requestFactory.NewResendPhoto(fmt.Sprintf("%d", command.MetaInfo.Message.Chat.ID), imgID, "", false, 0, nil, nil)
		}
		logger.Infof("Uploading and sending.")
		return requestFactory.NewUploadPhoto(fmt.Sprintf("%d", command.MetaInfo.Message.Chat.ID), imageFile, "", false, 0, nil, func(result *receive.SendResult) {
			if !result.Ok {
				logger.Errorf("Failed to send response. Received error: code - '%d', description '%s'.", result.ErrorCode, result.Description)
				return
			}
			var size int64
			var imageID string
			for i := range result.Result.Photos {
				if result.Result.Photos[i].Size > size {
					size = result.Result.Photos[i].Size
					imageID = result.Result.Photos[i].ID
				}
			}
			if len(imageID) > 0 {
				imageHolder.SetImageID(imageID)
			}
		})
	}, nil
}

func main() {
	logger := logs.NewStdLogger()
	botToken, ok := os.LookupEnv(BOT_TOKEN_KEY)
	if !ok {
		logger.Errorf("Failed to get bot-token. Expected to have environment variable '%s'.", BOT_TOKEN_KEY)
		return
	}

	imageFile, ok := os.LookupEnv(BOT_IMAGE_FILE_KEY)
	if !ok {
		logger.Errorf("Failed to get ssl public key. Expected to have environment variable '%s'.", BOT_IMAGE_FILE_KEY)
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
	onTest, err := NewOnTestCommand(usersMemory, requestFactory, imageFile, logger)
	if err != nil {
		logger.Errorf("Failed to create on-inline handler. Error: %v.", err)
		return
	}
	err = registry.RegisterCommand("/test", onTest)
	if err != nil {
		logger.Errorf("Initialization failed. Error: %v.", err)
		return
	}
	onUpdate, err := bot.NewDefaultUpdateCallback(logger, registry)
	if err != nil {
		logger.Errorf("Failed to create default update-callback. Error: %v.", err)
		return
	}

	bot := bot.NewPollingBot(requestFactory, onUpdate, 1000, logger)
	defer bot.Stop()
	logger.Infof("Bot started. Press Enter to stop.")
	_, _ = fmt.Scanf("\n")
}
