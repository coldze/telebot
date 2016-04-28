package telebot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"time"
)

type UpdateCallback func(update *receive.UpdateType) (*send.SendType, error)

type Bot interface {
	Send(*send.SendType) error
	Stop()
}

type botImpl struct {
	stopBot  chan struct{}
	logger   Logger
	factory  *send.RequestFactory
	OnUpdate UpdateCallback
	period   time.Duration
}

func (b *botImpl) Stop() {
	b.stopBot <- struct{}{}
}

func (b *botImpl) Send(*send.SendType) error {
	return errors.New("Not implemented.")
}

func post(message *send.SendType) ([]byte, error) {
	if message == nil {
		return nil, fmt.Errorf("Message is nil. Nothing to send.")
	}

	var reply *http.Response
	var err error
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
	return ioutil.ReadAll(reply.Body)
}

func poll(message *send.SendType) (*receive.UpdateResultType, error) {
	response, err := post(message)
	if err != nil {
		return nil, err
	}
	var updates receive.UpdateResultType
	err = json.Unmarshal(response, &updates)
	if err != nil {
		return nil, err
	}
	return &updates, nil
}

func sendResponse(message *send.SendType) (*receive.SendResult, error) {
	response, err := post(message)
	if err != nil {
		return nil, err
	}
	var sendResult receive.SendResult
	err = json.Unmarshal(response, &sendResult)
	if err != nil {
		/*if message.Type == send.SEND_TYPE_GET {
			err = nil
		}*/
		return nil, err
	}
	return &sendResult, nil
}

func (b *botImpl) run() {
	go func() {
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

func (b *botImpl) singUp(listenUrl string, port int64, sslPrivateKey string, sslPublicKey string) error {
	testFunc := func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			b.logger.Errorf("Failed to read update object")
			return
		}
		b.logger.Infof("Update body: %s", string(body))
		var update receive.UpdateType
		err = json.Unmarshal(body, &update)
		if err != nil {
			b.logger.Errorf("Failed to unmarshal update object")
			return
		}
		response, err := b.OnUpdate(&update)
		if err != nil {
			b.logger.Errorf("Failed to process update id '%d'. Error: %v.", update.ID, err)
			return
		}
		if response == nil {
			return
		}
		responseSentResult, err := sendResponse(response)
		if err != nil {
			b.logger.Errorf("Failed to send response for update id '%d'. Error: %v.", update.ID, err)
			return
		}
		if !responseSentResult.Ok {
			b.logger.Errorf("Failed to send response for update id '%d'. Received error: code - '%d', description '%s'.", update.ID, responseSentResult.ErrorCode, responseSentResult.Description)
		} else {
			b.logger.Infof("Response sent.")
		}
	}
	u, err := url.Parse(listenUrl)
	if err != nil {
		panic(err)
	}

	fmt.Println(u.Path)

	http.HandleFunc(u.Path, testFunc)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	f, err := os.Open(sslPublicKey)
	if err != nil {
		return err
	}
	defer f.Close()
	fw, err := w.CreateFormFile("certificate", sslPublicKey)
	if err != nil {
		return err
	}
	if _, err = io.Copy(fw, f); err != nil {
		return err
	}
	if fw, err = w.CreateFormField("url"); err != nil {
		return err
	}
	if _, err = fw.Write([]byte(listenUrl)); err != nil {
		return err
	}

	w.Close()

	req, err := http.NewRequest("POST", b.factory.SetWebhookURL, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return err
	}
	resText, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	b.logger.Infof("Web-hook sign-up result: %s", string(resText))
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", res.Status)
	}
	return http.ListenAndServeTLS(fmt.Sprintf(":%d", port), sslPublicKey, sslPrivateKey, nil)
}

func (b *botImpl) pollIteration(currentUpdateID int64) (lastUpdateID int64) {
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
		response, err := b.OnUpdate(&lastUpdate)
		if err != nil {
			b.logger.Errorf("Failed to process update id '%d'. Error: %v.", lastUpdate.ID, err)
			continue
		}
		if response == nil {
			continue
		}
		responseSentResult, err := sendResponse(response)
		if err != nil {
			b.logger.Errorf("Failed to send response for update id '%d'. Error: %v.", lastUpdate.ID, err)
			continue
		}
		if !responseSentResult.Ok {
			b.logger.Errorf("Failed to send response for update id '%d'. Received error: code - '%d', description '%s'.", lastUpdate.ID, responseSentResult.ErrorCode, responseSentResult.Description)
		} else {
			b.logger.Infof("Response sent.")
		}
	}
	return
}

func NewPollingBot(factory *send.RequestFactory, onUpdate UpdateCallback, pollPeriodMs int64, logger Logger) Bot {
	stopUpdatesChan := make(chan struct{})
	bot := botImpl{stopBot: stopUpdatesChan, logger: logger, factory: factory, OnUpdate: onUpdate, period: time.Duration(pollPeriodMs) * time.Millisecond}
	bot.run()
	return &bot
}

func NewWebHookBot(factory *send.RequestFactory, onUpdate UpdateCallback, url string, listenPort int64, sslPrivateKey string, sslPublicKey string, logger Logger) (Bot, error) {
	stopUpdatesChan := make(chan struct{})
	bot := botImpl{stopBot: stopUpdatesChan, logger: logger, factory: factory, OnUpdate: onUpdate}
	err := bot.singUp(url, listenPort, sslPrivateKey, sslPublicKey)
	if err != nil {
		return nil, err
	}
	return &bot, nil
}
