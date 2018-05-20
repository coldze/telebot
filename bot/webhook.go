package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/coldze/primitives/logs"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
)

type webHookBot struct {
	logger          logs.Logger
	factory         *send.RequestFactory
	updateProcessor UpdateProcessor
}

func (b *webHookBot) Stop() {
}

func (b *webHookBot) Send([]*send.SendType) error {
	return errors.New("Not implemented.")
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}
	w.Write([]byte(fmt.Sprintf("Ping from '%v'.\nReceived at: %v.", r.RemoteAddr, time.Now().UTC())))
}

func (b *webHookBot) singUp(listenUrl string, port int64, sslPrivateKey string, sslPublicKey string, isSelfSigned bool) error {
	testFunc := func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			b.logger.Errorf("Failed to read update object")
			return
		}
		var update receive.UpdateType
		err = json.Unmarshal(body, &update)
		if err != nil {
			b.logger.Errorf("Failed to unmarshal update object. Body: %+v. Error: %v", string(body), err)
			return
		}
		err = b.updateProcessor.Process(&update)
		if err != nil {
			b.logger.Errorf("Error has happened, while processing update id '%d'. Error: %v.", update.ID, err)
		}
	}
	u, err := url.Parse(listenUrl)
	if err != nil {
		panic(err)
	}

	http.HandleFunc(u.Path, testFunc)
	sslSubscribeKey := ""
	if isSelfSigned {
		sslSubscribeKey = sslPublicKey
	}
	signUp, err := b.factory.NewSubscribe(listenUrl, sslSubscribeKey)
	if err != nil {
		return err
	}

	signUpResult := []byte{}
	for i := range signUp {
		res, err := sendRequest(signUp[i])
		if err != nil {
			return err
		}
		signUpResult = append(signUpResult, res...)
	}

	http.HandleFunc("/ping", pingHandler)
	b.logger.Infof("Web-hook sign-up result: %s", string(signUpResult))
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	//return http.ListenAndServeTLS(fmt.Sprintf(":%d", port), sslPublicKey, sslPrivateKey, nil)
}

func NewWebHookBot(factory *send.RequestFactory, onUpdate UpdateCallback, url string, listenPort int64, sslPrivateKey string, sslPublicKey string, isSelfSigned bool, logger telebot.Logger) (Bot, error) {
	updateProcessor := &SyncUpdateProcessor{
		logger:   logger,
		onUpdate: onUpdate,
	}
	bot := webHookBot{
		logger:          logger,
		factory:         factory,
		updateProcessor: updateProcessor,
	}
	err := bot.singUp(url, listenPort, sslPrivateKey, sslPublicKey, isSelfSigned)
	if err != nil {
		return nil, err
	}
	return &bot, nil
}
