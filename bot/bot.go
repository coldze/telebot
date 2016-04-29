package bot

import (
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
)

type UpdateCallback func(update *receive.UpdateType) (*send.SendType, error)

type Bot interface {
	Send(*send.SendType) error
	Stop()
}

type UpdateProcessor interface {
	Process(update *receive.UpdateType) error
}
