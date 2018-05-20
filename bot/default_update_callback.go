package bot

import (
	"fmt"
	"strings"

	"github.com/coldze/primitives/custom_error"
	"github.com/coldze/primitives/logs"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
	"github.com/coldze/telebot/send/requests"
)

func parseCommand(message *receive.MessageType, update *receive.UpdateType) (*CommandCallType, string) {
	for index := range message.Entities {
		if message.Entities[index].Type != receive.ENTITY_TYPE_BOT_COMMAND {
			continue
		}
		var cmd CommandCallType
		cmdName := message.Text[message.Entities[index].Offset:message.Entities[index].Length]
		argsStart := message.Entities[index].Offset + message.Entities[index].Length
		if argsStart < int64(len(message.Text)) {
			cmd.Argument = strings.TrimSpace(message.Text[argsStart:])
		}
		cmd.MetaInfo = update
		return &cmd, cmdName
	}
	return nil, ""
}

func parseCommandFromCallbackQuery(update *receive.UpdateType) (*CommandCallType, string) {
	return &CommandCallType{
		MetaInfo: update,
	}, update.CallbackQuery.Data
}

func wrapResult(factory *send.RequestFactory, chatID interface{}, errValue custom_error.CustomError) ([]*send.SendType, custom_error.CustomError) {
	res, customErr := factory.NewSendMessage(chatID, fmt.Sprintf("Internal error: %v", errValue), 0, false, false, 0, nil)
	if customErr == nil {
		return res, nil
	}
	return nil, custom_error.NewErrorf(customErr, "Failed to send error notification. Original error: %v", errValue)
}

func wrapCallbackResult(factory *send.RequestFactory, chatID interface{}, callbackID interface{}, errValue custom_error.CustomError) ([]*send.SendType, custom_error.CustomError) {
	resp, customErr := factory.NewSendMessage(chatID, fmt.Sprintf("Internal error: %v", errValue), 0, false, false, 0, nil)
	if customErr != nil {
		return nil, custom_error.NewErrorf(customErr, "Failed to send error notification. Original error: %v", errValue)
	}
	respCallback, customErr := factory.NewAnswerCallbackQuery(&requests.AnswerCallbackQuery{
		CallbackQueryID: callbackID,
		ShowAlert:       true,
		Text:            "Failed to process",
	})
	if customErr != nil {
		return resp, custom_error.NewErrorf(customErr, "Failed to create answer callback query")
	}
	return append(resp, respCallback...), nil
}

func NewDefaultUpdateCallback(factory *send.RequestFactory, logger logs.Logger, handlers *BotHandlers) (UpdateCallback, custom_error.CustomError) {
	if logger == nil {
		return nil, custom_error.MakeErrorf("Invalid logger specified.")
	}
	if handlers == nil {
		return nil, custom_error.MakeErrorf("Invalid handlers specified.")
	}

	return func(update *receive.UpdateType) ([]*send.SendType, custom_error.CustomError) {
		if update.CallbackQuery != nil {
			cmd, cmdName := parseCommandFromCallbackQuery(update)
			resp, err := handlers.OnCommand(cmdName, cmd)
			if err == nil {
				return resp, nil
			}
			logger.Errorf("Failed to handle command '%v'. Error: %v", cmd, err)
			errResp, customErr := wrapCallbackResult(factory, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.ID, custom_error.NewErrorf(err, "Failed to handle command"))
			if customErr != nil {
				customErr = custom_error.NewErrorf(customErr, "Failed to send error through callback.")
				logger.Errorf("Failed to send error through callback after error handling command '%v'. Error: %v", cmd, customErr)
				return nil, customErr
			}
			return errResp, nil
		}
		if update == nil {
			return nil, nil
		}
		msg := GetMessage(update)
		if msg == nil {
			resp, customErr := handlers.OnMessage(update)
			if customErr != nil {
				customErr = custom_error.NewErrorf(customErr, "Failed to handle message")
				logger.Errorf("Failed to handle empty message. Error: %v", customErr)
				return nil, customErr
			}
			return resp, nil
		}
		chatID := int64(-1)
		if msg.Chat != nil {
			chatID = msg.Chat.ID
		}
		if msg.Entities == nil {
			resp, err := handlers.OnMessage(update)
			if err == nil {
				return resp, nil
			}
			logger.Errorf("Failed to handle message. Error: %v", err)
			errResp, customErr := wrapResult(factory, chatID, custom_error.NewErrorf(err, "Failed to handle message"))
			if customErr != nil {
				customErr = custom_error.NewErrorf(customErr, "Failed to send error")
				logger.Errorf("Failed to send error after error handling message. Error: %v", customErr)
				return nil, customErr
			}
			return errResp, nil
		}
		cmd, cmdName := parseCommand(msg, update)
		if cmd == nil {
			resp, err := handlers.OnMessage(update)
			logger.Errorf("Failed to handle message. Error: %v", err)
			if err == nil {
				return resp, nil
			}
			errResp, customErr := wrapResult(factory, chatID, custom_error.NewErrorf(err, "Failed to handle message"))
			if customErr != nil {
				customErr = custom_error.NewErrorf(customErr, "Failed to send error")
				logger.Errorf("Failed to send error after error handling message. Error: %v", customErr)
				return nil, customErr
			}
			return errResp, nil
		}
		resp, err := handlers.OnCommand(cmdName, cmd)
		if err == nil {
			return resp, nil
		}
		errResp, customErr := wrapResult(factory, chatID, custom_error.NewErrorf(err, "Failed to handle command '%v'", cmdName))
		if customErr != nil {
			customErr = custom_error.NewErrorf(customErr, "Failed to send error")
			logger.Errorf("Failed to send error after error handling command. Error: %v", customErr)
			return nil, customErr
		}
		return errResp, nil
	}, nil
}
