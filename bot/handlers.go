package bot

import (
	"github.com/coldze/primitives/custom_error"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
)

type CommandCallType struct {
	Argument string
	MetaInfo *receive.UpdateType
}

type CommandHandler func(command *CommandCallType) ([]*send.SendType, custom_error.CustomError)
type MessageHandler func(message *receive.UpdateType) ([]*send.SendType, custom_error.CustomError)

type BotHandlers struct {
	handlers  map[string]CommandHandler
	onMessage MessageHandler
}

func (r *BotHandlers) RegisterCommand(commandName string, handler CommandHandler) custom_error.CustomError {
	_, ok := r.handlers[commandName]
	if ok {
		return custom_error.MakeErrorf("Command handler already registered: %s.", commandName)
	}
	r.handlers[commandName] = handler
	return nil
}

func (r *BotHandlers) OnCommand(commandName string, args *CommandCallType) ([]*send.SendType, custom_error.CustomError) {
	handler, ok := r.handlers[commandName]
	if !ok {
		return nil, custom_error.MakeErrorf("Handler not set for command: %s.", commandName)
	}
	res, err := handler(args)
	if err == nil {
		return res, nil
	}
	return nil, custom_error.NewErrorf(err, "Failed to handle command '%v'", commandName)
}

func (r *BotHandlers) OnMessage(message *receive.UpdateType) ([]*send.SendType, custom_error.CustomError) {
	if r.onMessage == nil {
		return nil, custom_error.MakeErrorf("No default handler.")
	}
	res, err := r.onMessage(message)
	if err == nil {
		return res, nil
	}
	return nil, custom_error.NewErrorf(err, "Failed to handle message.")
}

func NewBotHandlers(onMessage MessageHandler) *BotHandlers {
	return &BotHandlers{onMessage: onMessage, handlers: make(map[string]CommandHandler)}
}
