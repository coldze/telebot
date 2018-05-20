package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/coldze/primitives/custom_error"
	"github.com/coldze/primitives/logs"
	"github.com/coldze/telebot/bot"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
	"github.com/coldze/telebot/send/requests"
	"time"
)

const (
	BOT_TOKEN_KEY      = "BOT_TOKEN"
	BOT_IMAGE_FILE_KEY = "BOT_IMAGE_FILE"

	STICKER_ID = "BQADAgADQAADyIsGAAGMQCvHaYLU_AI"
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

func NewOnTestCommand(users *UsersMemory, requestFactory *send.RequestFactory, imageFile string, logger logs.Logger) (bot.CommandHandler, custom_error.CustomError) {
	if users == nil {
		return nil, nil
	}
	if requestFactory == nil {
		return nil, nil
	}
	var imageHolder testImageHolder
	return func(command *bot.CommandCallType) ([]*send.SendType, custom_error.CustomError) {
		logger.Infof("Test command handler invoked.")
		if command.MetaInfo.Message.From == nil {
			return nil, custom_error.MakeErrorf("FROM missing")
		}
		if command.MetaInfo.Message.Chat == nil {
			return nil, custom_error.MakeErrorf("CHAT missing")
		}

		imgID := imageHolder.GetImageID()
		if len(imgID) > 0 {
			logger.Infof("Re-sending.")
			sendUploaded := &requests.SendUploadedPhoto{
				Base: requests.Base{
					ChatID: command.MetaInfo.Message.Chat.ID,
				},
				Photo: imgID,
			}
			res, err := requestFactory.NewSendUploadedPhoto(sendUploaded, nil)
			if err == nil {
				return res, nil
			}
			return nil, custom_error.NewErrorf(err, "Failed to create new resend photo.")
		}
		logger.Infof("Uploading and sending.")
		onSentCallback := func(result *receive.SendResult, err custom_error.CustomError) {
			if err != nil {
				logger.Errorf("Failed to send response. Error: %v", err)
			}
			if result == nil {
				logger.Warningf("Result is nil.")
			}
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
		}
		sendFile := &requests.SendFileBase{
			Base: requests.Base{
				ChatID: command.MetaInfo.Message.Chat.ID,
			},
			FileName: imageFile,
		}
		res, err := requestFactory.NewUploadPhoto(sendFile, onSentCallback)
		if err == nil {
			return res, nil
		}
		return nil, custom_error.NewErrorf(err, "Failed to create new upload photo.")
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
	onUpdate, err := bot.NewDefaultUpdateCallback(requestFactory, logger, registry)
	if err != nil {
		logger.Errorf("Failed to create default update-callback. Error: %v.", err)
		return
	}

	updateProcessor := bot.NewUpdateProcessor(onUpdate, logger)
	botApp := bot.NewPollingBot(requestFactory, updateProcessor, time.Second, logger)
	defer botApp.Stop()
	logger.Infof("Bot started. Press Enter to stop.")
	_, _ = fmt.Scanf("\n")
}
