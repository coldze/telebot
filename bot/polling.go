package bot

import (
	"errors"
	"github.com/coldze/telebot"
	"github.com/coldze/telebot/send"
	"runtime/debug"
	"time"
)

type pollingBot struct {
	stopBot         chan struct{}
	logger          telebot.Logger
	factory         *send.RequestFactory
	period          time.Duration
	updateProcessor UpdateProcessor
}

func (b *pollingBot) Stop() {
	b.stopBot <- struct{}{}
}

func (b *pollingBot) Send(*send.SendType) error {
	return errors.New("Not implemented.")
}

func (b *pollingBot) run() {
	go func() {
		unsubscribe, err := b.factory.NewUnsubscribe()
		if err != nil {
			b.logger.Errorf("Failed to create unsubscribe request. Error: %v.", err)
		} else {
			_, err = sendRequest(unsubscribe)
			if err != nil {
				b.logger.Errorf("Failed to unsubscribe. Error: %v.", err)
			}
		}
		var lastUpdateID int64
		for {
			select {
			case _ = <-b.stopBot:
				b.logger.Infof("Update-polling goroutine exiting")
				return
			default:
				time.Sleep(b.period)
			}
			lastUpdateID = b.pollIteration(lastUpdateID)
		}
	}()
}

func (b *pollingBot) pollIteration(currentUpdateID int64) (lastUpdateID int64) {
	lastUpdateID = currentUpdateID
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		err, ok := r.(error)
		if ok {
			b.logger.Errorf("PANIC occured. Error: %v. Call-Stack:\n%s.", err, string(debug.Stack()))
		} else {
			b.logger.Errorf("PANIC occured. Recover-objet: %+v. Call-Stack:\n%s.", r, string(debug.Stack()))
		}
	}()
	getUpdatesRequest, err := b.factory.NewGetUpdates(currentUpdateID+1, 0, 0)
	if err != nil {
		b.logger.Errorf("Failed to prepare update request. Error: %v.", err)
		return
	}
	updates, err := poll(getUpdatesRequest)
	if err != nil {
		b.logger.Errorf("Failed to pull updates. Error: %v.", err)
		return
	}
	if !updates.Ok {
		b.logger.Errorf("Bad updates object...")
		return
	}
	if len(updates.Updates) <= 0 {
		return
	}
	for updateIndex := range updates.Updates {
		lastUpdate := updates.Updates[updateIndex]
		index := lastUpdate.ID
		if index > lastUpdateID {
			lastUpdateID = index
		}
		err = b.updateProcessor.Process(&lastUpdate)
		if err != nil {
			b.logger.Errorf("Error has happened, while processing update id '%d'. Error: %v.", lastUpdate.ID, err)
		}
	}
	return
}

func NewPollingBot(factory *send.RequestFactory, onUpdate UpdateCallback, pollPeriodMs int64, logger telebot.Logger) Bot {
	stopUpdatesChan := make(chan struct{})
	updateProcessor := &SyncUpdateProcessor{logger: logger, onUpdate: onUpdate}
	bot := pollingBot{stopBot: stopUpdatesChan, logger: logger, factory: factory, updateProcessor: updateProcessor, period: time.Duration(pollPeriodMs) * time.Millisecond}
	bot.run()
	return &bot
}
