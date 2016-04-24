package telebot

import (
	"github.com/coldze/telebot/send"
	"fmt"
	"github.com/coldze/telebot/receive"
	"net/http"
	"bytes"
	"io/ioutil"
	"encoding/json"
	"time"
	"errors"
)

type Logger interface {
	Warningf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
}

type UpdateCallback func(update *receive.UpdateType) (*send.SendType, error)

type Bot interface {
	Send(*send.SendType) error
	Stop()
}

type botImpl struct {
	stopBot chan struct{}
	logger Logger
	factory *send.RequestFactory
	OnUpdate UpdateCallback
}

func (b *botImpl) Stop() {
	b.stopBot <- struct{}{}
}

func (b *botImpl) Send(*send.SendType) error {
	return errors.New("Not implemented.")
}

func post(message *send.SendType) (result []byte, err error) {
	if message == nil {
		return nil, fmt.Errorf("Message is nil. Nothing to send.")
	}

	var reply *http.Response
	switch message.Type {
	case send.SEND_TYPE_POST:
		buffer := bytes.NewReader(message.Parameters)
		reply, err = http.Post(message.URL, "application/json", buffer)
	case send.SEND_TYPE_GET:
		reply, err = http.Get(message.URL)
	}
	if reply != nil {
		defer reply.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	respBody, err := ioutil.ReadAll(reply.Body)
	if err != nil {
		return nil, err
	}
	return respBody, nil
}

func (b *botImpl) run() {
	go func () {
		var lastUpdateID int64
		for {
			select {
			case _ = <-b.stopBot:
				b.logger.Infof("Update-polling goroutine exiting")
				return
			default:
				time.Sleep(500 * time.Millisecond)
			}
			getUpdatesRequest, err := b.factory.NewGetUpdates(lastUpdateID + 1, 0, 0)
			if err != nil {
				b.logger.Errorf("Failed to prepare update request. Error: %v.", err)
				continue
			}
			response, err := post(getUpdatesRequest)
			var updates receive.UpdateResultType
			if err != nil {
				b.logger.Errorf("Failed to pull updates. Error: %v.", err)
				continue
			}
			err = json.Unmarshal(response, &updates)
			if err != nil {
				b.logger.Errorf("Failed to unmarshal. Error: %v.\n", err)
				continue
			}
			if !updates.Ok {
				b.logger.Errorf("Bad updates object...")
				continue
			}
			if len(updates.Updates) <= 0 {
				continue
			}

			/*indented, err := json.MarshalIndent(&updates, "", "    ")
			if err != nil {
				fmt.Printf("Failed to indent... Error: %v.\n", err)
			} else {
				fmt.Println(string(indented))
			}*/

			for updateIndex := range updates.Updates {
				lastUpdate := updates.Updates[updateIndex]
				index := lastUpdate.ID
				if index > lastUpdateID {
					lastUpdateID = index
				}
				response, err := b.OnUpdate(&lastUpdate)
				if err != nil {
					b.logger.Errorf("Failed to process update id '%d'. Error: %v.", lastUpdate.ID, err)
					continue
				}
				if response == nil {
					continue
				}
				responseSentResult, err := post(response)
				if err != nil {
					b.logger.Errorf("Failed to send response for update id '%d'. Error: %v.", lastUpdate.ID, err)
					continue
				}
				b.logger.Infof("Sent response: %s", string(responseSentResult))
			}
		}
	} ()
}

func NewPollingBot(factory *send.RequestFactory, onUpdate UpdateCallback, logger Logger) Bot {
	stopUpdatesChan := make(chan struct{})
	bot := botImpl{stopBot: stopUpdatesChan, logger: logger, factory: factory, OnUpdate: onUpdate}
	bot.run()
	return &bot
}

type StdoutLogger struct {
}

func (l *StdoutLogger) print(prefix string, format string, args ...interface{}) {
	fmt.Printf(prefix + format + "\n", args...)
}

func (l *StdoutLogger) Warningf(format string, args ...interface{}) {
	l.print("[WARNING] ", format, args...)
}
func (l* StdoutLogger) Debugf(format string, args ...interface{}) {
	l.print("[DEBUG] ", format, args...)
}

func (l* StdoutLogger) Errorf(format string, args ...interface{}) {
	l.print("[ERROR] ", format, args...)
}

func (l* StdoutLogger) Infof(format string, args ...interface{}) {
	l.print("[INFO] ", format, args...)
}

func NewStdoutLogger() Logger {
	return &StdoutLogger{}
}