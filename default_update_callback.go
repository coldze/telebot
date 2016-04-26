package telebot

import (
	"errors"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
	"strings"
)

func parseCommand(update *receive.UpdateType) (*CommandCallType, string) {
	for index := range update.Message.Entities {
		if update.Message.Entities[index].Type != receive.ENTITY_TYPE_BOT_COMMAND {
			continue
		}
		var cmd CommandCallType
		cmdName := update.Message.Text[update.Message.Entities[index].Offset:update.Message.Entities[index].Length]
		argsStart := update.Message.Entities[index].Offset + update.Message.Entities[index].Length
		if argsStart < int64(len(update.Message.Text)) {
			cmd.Argument = strings.TrimSpace(update.Message.Text[argsStart:])
		}
		cmd.MetaInfo = update
		return &cmd, cmdName
	}
	return nil, ""
}

func NewDefaultUpdateCallback(logger Logger, handlers *BotHandlers) (UpdateCallback, error) {
	if logger == nil {
		return nil, errors.New("Invalid logger specified.")
	}
	if handlers == nil {
		return nil, errors.New("Invalid handlers specified.")
	}
	return func(update *receive.UpdateType) (*send.SendType, error) {
		if update == nil {
			return nil, nil
		}
		if update.Message == nil {
			return nil, nil
		}
		if update.Message.Entities == nil {
			return handlers.OnMessage(update)
		}
		cmd, cmdName := parseCommand(update)
		if cmd == nil {
			return handlers.OnMessage(update)
		}
		return handlers.OnCommand(cmdName, cmd)
	}, nil
}
