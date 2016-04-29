package bot

import (
	"errors"
	"fmt"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
)

type CommandCallType struct {
	Argument string
	MetaInfo *receive.UpdateType
}

type CommandHandler func(command *CommandCallType) (*send.SendType, error)
type MessageHandler func(message *receive.UpdateType) (*send.SendType, error)

type BotHandlers struct {
	handlers  map[string]CommandHandler
	onMessage MessageHandler
}

func (r *BotHandlers) RegisterCommand(commandName string, handler CommandHandler) error {
	_, ok := r.handlers[commandName]
	if ok {
		return fmt.Errorf("Command handler already registered: %s.", commandName)
	}
	r.handlers[commandName] = handler
	return nil
}

func (r *BotHandlers) OnCommand(commandName string, args *CommandCallType) (*send.SendType, error) {
	handler, ok := r.handlers[commandName]
	if !ok {
		return nil, fmt.Errorf("Handler not set for command: %s.", commandName)
	}
	return handler(args)
}

func (r *BotHandlers) OnMessage(message *receive.UpdateType) (*send.SendType, error) {
	if r.onMessage == nil {
		return nil, errors.New("No default handler.")
	}
	return r.onMessage(message)
}

func NewBotHandlers(onMessage MessageHandler) *BotHandlers {
	return &BotHandlers{onMessage: onMessage, handlers: make(map[string]CommandHandler)}
}
