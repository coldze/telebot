package bot

import (
	"errors"
	"fmt"
	"github.com/coldze/telebot"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
	"log"
	"strings"
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

func NewDefaultUpdateCallback(factory *send.RequestFactory, logger telebot.Logger, handlers *BotHandlers) (UpdateCallback, error) {
	if logger == nil {
		return nil, errors.New("Invalid logger specified.")
	}
	if handlers == nil {
		return nil, errors.New("Invalid handlers specified.")
	}
	wrapResult := func(chatID interface{}, resp []*send.SendType, errValue error) ([]*send.SendType, error) {
		if errValue == nil {
			return resp, errValue
		}
		return factory.NewSendMessage(chatID, fmt.Sprintf("Internal error: %v", errValue), 0, false, false, 0, nil)
	}
  wrapCallbackResult := func(chatID interface{}, callbackID interface{}, resp []*send.SendType, errValue error) ([]*send.SendType, error) {
    if errValue == nil {
      return resp, errValue
    }
    resp, err := factory.NewSendMessage(chatID, fmt.Sprintf("Internal error: %v", errValue), 0, false, false, 0, nil)
    if err != nil {
      return nil, err
    }
    respCallback, err := factory.NewAnswerCallbackQuery(&requests.AnswerCallbackQuery{
      CallbackQueryID: callbackID,
      ShowAlert: true,
      Text: "Failed to process",
    })
    if err != nil {
      return resp, err
    }
    return append(resp, respCallback...), nil
  }
	return func(update *receive.UpdateType) ([]*send.SendType, error) {
		log.Printf("%+v", update)
		log.Printf("%+v", update.CallbackQuery)
		if update.CallbackQuery != nil {
			log.Printf("%+v", update.CallbackQuery.Message)
			cmd, cmdName := parseCommandFromCallbackQuery(update)
			resp, err := handlers.OnCommand(cmdName, cmd)
			return wrapCallbackResult(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.ID, resp, err)
		}
		if update == nil {
			return nil, nil
		}
		msg := GetMessage(update)
		if msg == nil {
			return handlers.OnMessage(update)
		}
		chatID := int64(-1)
		if msg.Chat != nil {
			chatID = msg.Chat.ID
		}
		if msg.Entities == nil {
			resp, err := handlers.OnMessage(update)
			return wrapResult(chatID, resp, err)
		}
		cmd, cmdName := parseCommand(msg, update)
		if cmd == nil {
			resp, err := handlers.OnMessage(update)
			return wrapResult(chatID, resp, err)
		}
		resp, err := handlers.OnCommand(cmdName, cmd)
		return wrapResult(chatID, resp, err)
	}, nil
}
