package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/coldze/primitives/custom_error"
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

func (b *webHookBot) Send([]*send.SendType) custom_error.CustomError {
	return custom_error.MakeErrorf("Not implemented.")
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}
	w.Write([]byte(fmt.Sprintf("Ping from '%v'.\nReceived at: %v.", r.RemoteAddr, time.Now().UTC())))
}

func newHandlingFunc(logger logs.Logger, updateProcessor UpdateProcessor) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Errorf("Failed to read update object. Error: %v", custom_error.MakeErrorf("Failed to read request. Error: %v", err))
			return
		}
		var update receive.UpdateType
		err = json.Unmarshal(body, &update)
		if err != nil {
			logger.Errorf("Failed to unmarshal update object. Body: %+v. Error: %v", string(body), custom_error.MakeErrorf("Failed to unmarshal request body. Error: %v", err))
			return
		}
		customErr := updateProcessor.Process(&update)
		if customErr != nil {
			logger.Errorf("Error has happened, while processing update id '%d'. Error: %v.", update.ID, custom_error.NewErrorf(customErr, "Failed to process update."))
		}
	}
}

func (b *webHookBot) singUp(listenUrl string, port int64, sslPrivateKey string, sslPublicKey string, isSelfSigned bool) custom_error.CustomError {
	handlingFunc := newHandlingFunc(b.logger, b.updateProcessor)
	u, err := url.Parse(listenUrl)
	if err != nil {
		return custom_error.MakeErrorf("Failed to parse url '%v'. Error: %v", listenUrl, err)
	}

	http.HandleFunc(u.Path, handlingFunc)
	sslSubscribeKey := ""
	if isSelfSigned {
		sslSubscribeKey = sslPublicKey
	}
	signUp, customErr := b.factory.NewSubscribe(listenUrl, sslSubscribeKey)
	if customErr != nil {
		return custom_error.NewErrorf(customErr, "Failed to subscribe.")
	}

	signUpResult := []byte{}
	for i := range signUp {
		res, customErr := sendRequest(signUp[i])
		if customErr != nil {
			return custom_error.NewErrorf(customErr, "Failed to send request.")
		}
		signUpResult = append(signUpResult, res...)
	}

	http.HandleFunc("/ping", pingHandler)
	b.logger.Infof("Web-hook sign-up result: %s", string(signUpResult))
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err == nil {
		return nil
	}
	return custom_error.MakeErrorf("Failed to start http server. Error: %v", err)
	//return http.ListenAndServeTLS(fmt.Sprintf(":%d", port), sslPublicKey, sslPrivateKey, nil)
}

func NewWebHookBot(factory *send.RequestFactory, onUpdate UpdateCallback, url string, listenPort int64, sslPrivateKey string, sslPublicKey string, isSelfSigned bool, logger logs.Logger) (Bot, custom_error.CustomError) {
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
		return nil, custom_error.NewErrorf(err, "Failed to sign-up for changes")
	}
	return &bot, nil
}
