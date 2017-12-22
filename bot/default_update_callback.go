package bot

import (
	"errors"
	"github.com/coldze/telebot"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
	"log"
	"strings"
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

func getMessage(update *receive.UpdateType) *receive.MessageType {
	if update.Message != nil {
		return update.Message
	}
	if update.EditedMessage != nil {
		return update.EditedMessage
	}
	if update.ChannelPost != nil {
		return update.ChannelPost
	}
	if update.EditedChannelPost != nil {
		return update.EditedChannelPost
	}
	return nil
}

func NewDefaultUpdateCallback(logger telebot.Logger, handlers *BotHandlers) (UpdateCallback, error) {
	if logger == nil {
		return nil, errors.New("Invalid logger specified.")
	}
	if handlers == nil {
		return nil, errors.New("Invalid handlers specified.")
	}
	return func(update *receive.UpdateType) (*send.SendType, error) {
		log.Printf("%+v", update)
		log.Printf("%+v", update.CallbackQuery)
		if update.CallbackQuery != nil {
			log.Printf("%+v", update.CallbackQuery.Message)
		}
		if update == nil {
			return nil, nil
		}
		msg := getMessage(update)
		if msg == nil {
			return handlers.OnMessage(update)
		}
		if msg.Entities == nil {
			return handlers.OnMessage(update)
		}
		cmd, cmdName := parseCommand(msg, update)
		if cmd == nil {
			return handlers.OnMessage(update)
		}
		return handlers.OnCommand(cmdName, cmd)
	}, nil
}
