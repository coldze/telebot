package bot

import (
	"github.com/coldze/primitives/custom_error"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
)

type UpdateCallback func(update *receive.UpdateType) ([]*send.SendType, custom_error.CustomError)

type Bot interface {
	Send([]*send.SendType) custom_error.CustomError
	Stop()
}

type UpdateProcessor interface {
	Process(update *receive.UpdateType) custom_error.CustomError
}
