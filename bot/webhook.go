package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coldze/telebot"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
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

func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}
	log.Printf("%+v", r.Header)
	w.Write([]byte(fmt.Sprintf("Ping from '%v'.\nReceived at: %v.", r.RemoteAddr, time.Now().UTC())))
}

func (b *webHookBot) singUp(listenUrl string, port int64, sslPrivateKey string, sslPublicKey string, isSelfSigned bool) error {
	testFunc := func(w http.ResponseWriter, r *http.Request) {
		log.Print("In callback")
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			b.logger.Errorf("Failed to read update object")
			return
		}
		log.Printf("Body: %+v", string(body))
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
	log.Printf("Callback path is: %v", u.Path)
	sslSubscribeKey := ""
	if isSelfSigned {
		sslSubscribeKey = sslPublicKey
	}
	signUp, err := b.factory.NewSubscribe(listenUrl, sslSubscribeKey)
	if err != nil {
		return err
	}

	res, err := sendRequest(signUp)
	if err != nil {
		return err
	}
	http.HandleFunc("/telebot/ping", pingHandler)
	b.logger.Infof("Web-hook sign-up result: %s", string(res))
	//return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	return http.ListenAndServeTLS(fmt.Sprintf(":%d", port), sslPublicKey, sslPrivateKey, nil)
}

func NewWebHookBot(factory *send.RequestFactory, onUpdate UpdateCallback, url string, listenPort int64, sslPrivateKey string, sslPublicKey string, isSelfSigned bool, logger telebot.Logger) (Bot, error) {
	updateProcessor := &SyncUpdateProcessor{logger: logger, onUpdate: onUpdate}
	bot := webHookBot{logger: logger, factory: factory, updateProcessor: updateProcessor}
	err := bot.singUp(url, listenPort, sslPrivateKey, sslPublicKey, isSelfSigned)
	if err != nil {
		return nil, err
	}
	return &bot, nil
}
