package bot

import (
	"runtime/debug"
	"time"

	"github.com/coldze/primitives/logs"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
)

type pollingBot struct {
	stopBot         chan struct{}
	logger          logs.Logger
	factory         *send.RequestFactory
	period          time.Duration
	updateProcessor UpdateProcessor
}

func (b *pollingBot) Stop() {
	b.stopBot <- struct{}{}
}

func (b *pollingBot) Send(msg []*send.SendType) error {
	_, err := sendResponse(msg)
	if err != nil {
		return err
	}
	return nil
}

func (b *pollingBot) sendRequests(messages []*send.SendType) error {
	var err error
	for i := range messages {
		_, err = sendRequest(messages[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *pollingBot) run() {
	go func() {
		unsubscribe, err := b.factory.NewUnsubscribe()
		if err != nil {
			b.logger.Errorf("Failed to create unsubscribe request. Error: %v.", err)
		} else {
			err = b.sendRequests(unsubscribe)
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

func (b *pollingBot) processUpdates(updates *receive.UpdateResultType, lastUpdateID int64) int64 {
	if !updates.Ok {
		b.logger.Errorf("Bad updates object...")
		return lastUpdateID
	}
	if len(updates.Updates) <= 0 {
		return lastUpdateID
	}
	lastUpdateIDValue := lastUpdateID
	for updateIndex := range updates.Updates {
		lastUpdate := updates.Updates[updateIndex]
		index := lastUpdate.ID
		if index > lastUpdateIDValue {
			lastUpdateIDValue = index
		}
		err := b.updateProcessor.Process(&lastUpdate)
		if err != nil {
			b.logger.Errorf("Error has happened, while processing update id '%d'. Error: %v.", lastUpdate.ID, err)
		}
	}
	return lastUpdateIDValue
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
	for i := range getUpdatesRequest {
		updates, err := poll(getUpdatesRequest[i])
		if err != nil {
			b.logger.Errorf("Failed to pull updates. Error: %v.", err)
			return
		}
		lastUpdateID = b.processUpdates(updates, lastUpdateID)
	}

	return
}

func NewPollingBot(factory *send.RequestFactory, onUpdate UpdateCallback, pollPeriodMs int64, logger logs.Logger) Bot {
	stopUpdatesChan := make(chan struct{})
	updateProcessor := &SyncUpdateProcessor{logger: logger, onUpdate: onUpdate}
	bot := pollingBot{stopBot: stopUpdatesChan, logger: logger, factory: factory, updateProcessor: updateProcessor, period: time.Duration(pollPeriodMs) * time.Millisecond}
	bot.run()
	return &bot
}
