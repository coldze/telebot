package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coldze/telebot"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
	"io/ioutil"
	"net/http"
	"net/url"
)

type webHookBot struct {
	logger          telebot.Logger
	factory         *send.RequestFactory
	updateProcessor UpdateProcessor
}

func (b *webHookBot) Stop() {
}

func (b *webHookBot) Send(*send.SendType) error {
	return errors.New("Not implemented.")
}

func (b *webHookBot) singUp(listenUrl string, port int64, sslPrivateKey string, sslPublicKey string) error {
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
			b.logger.Errorf("Failed to unmarshal update object")
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
	signUp, err := b.factory.NewSignUp(listenUrl, sslPublicKey)
	if err != nil {
		return err
	}

	res, err := sendRequest(signUp)
	if err != nil {
		return err
	}
	b.logger.Infof("Web-hook sign-up result: %s", string(res))
	return http.ListenAndServeTLS(fmt.Sprintf(":%d", port), sslPublicKey, sslPrivateKey, nil)
}

func NewWebHookBot(factory *send.RequestFactory, onUpdate UpdateCallback, url string, listenPort int64, sslPrivateKey string, sslPublicKey string, logger telebot.Logger) (Bot, error) {
	updateProcessor := &SyncUpdateProcessor{logger: logger, onUpdate: onUpdate}
	bot := webHookBot{logger: logger, factory: factory, updateProcessor: updateProcessor}
	err := bot.singUp(url, listenPort, sslPrivateKey, sslPublicKey)
	if err != nil {
		return nil, err
	}
	return &bot, nil
}
