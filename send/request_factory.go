package send

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coldze/telebot"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send/internal/requests"
	"io"
	"mime/multipart"
	"os"
)

const (
	PARSE_MODE_HTML = iota + 1
	PARSE_MODE_MARKDOWN
)

const (
	parse_mode_html     = "HTML"
	parse_mode_markdown = "Markdown"

	bot_query_fmt = "https://api.telegram.org/bot%s/"

	cmd_get_updates             = "%sgetUpdates"
	cmd_set_web_hook            = "%ssetWebhook"
	cmd_get_me                  = "%sgetMe"
	cmd_send_message            = "%ssendMessage"
	cmd_forward_message         = "%sforwardMessage"
	cmd_send_photo              = "%ssendPhoto"
	cmd_send_audio              = "%ssendAudio"
	cmd_send_document           = "%ssendDocument"
	cmd_send_sticker            = "%ssendSticker"
	cmd_send_video              = "%ssendVideo"
	cmd_send_voice              = "%ssendVoice"
	cmd_send_location           = "%ssendLocation"
	cmd_send_venue              = "%ssendVenue"
	cmd_send_contact            = "%ssendContact"
	cmd_send_chat_action        = "%ssendChatAction"
	cmd_get_user_profile_photos = "%sgetUserProfilePhotos"
	cmd_get_file                = "https://api.telegram.org/file/bot%s/%s"
	cmd_kick_chat_member        = "%skickChatMember"
	cmd_unban_chat_member       = "%sunbanChatMember"
	cmd_answer_callback_query   = "%sanswerCallbackQuery"

	content_type_application_json = "application/json"
)

type RequestFactory struct {
	sendStickerURL  string
	sendMessageURL  string
	sendPhotoURL    string
	getUpdatesURL   string
	defaultCallback OnSentCallback
	setWebhookURL   string
}

func writeFieldString(writer *multipart.Writer, fieldName string, value string) error {
	return writeFieldBytes(writer, fieldName, []byte(value))
}

func writeFieldBytes(writer *multipart.Writer, fieldName string, value []byte) error {
	fieldWriter, err := writer.CreateFormField(fieldName)
	if err != nil {
		return err
	}
	_, err = fieldWriter.Write(value)
	return err
}

func (f *RequestFactory) NewSendRaw(url string, message interface{}) (*SendType, error) {
	request, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	return &SendType{URL: f.sendStickerURL, Parameters: request}, nil
}

func (f *RequestFactory) newPostSendType(url string, message interface{}, contentType string, callback OnSentCallback) (*SendType, error) {
	requestMessage, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	return f.newPostSendTypeBytes(url, requestMessage, contentType, callback)
}

func (f *RequestFactory) newPostSendTypeBytes(url string, message []byte, contentType string, callback OnSentCallback) (*SendType, error) {
	if callback == nil {
		return &SendType{URL: url, Parameters: message, Type: SEND_TYPE_POST, ContentType: contentType, Callback: f.defaultCallback}, nil
	}
	return &SendType{URL: url, Parameters: message, Type: SEND_TYPE_POST, ContentType: contentType, Callback: callback}, nil
}

func (f *RequestFactory) NewSendSticker(chatID string, sticker string, disableNotification bool, replyToMessageID int64, markup interface{}) (*SendType, error) {
	stickerMessage := send_requests.SendSticker{
		ChatID:              chatID,
		Sticker:             sticker,
		DisableNotification: disableNotification,
		ReplyToMessageID:    replyToMessageID}
	return f.newPostSendType(f.sendStickerURL, stickerMessage, content_type_application_json, nil)
}

func (f *RequestFactory) NewUnsubscribe() (*SendType, error) {
	return f.NewSubscribe("", "")
}

func (f *RequestFactory) NewSubscribe(url string, sslPublicKey string) (*SendType, error) {
	var buf bytes.Buffer
	bufferWriter := multipart.NewWriter(&buf)
	if len(sslPublicKey) > 0 {
		sslCertificate, err := os.Open(sslPublicKey)
		if err != nil {
			return nil, err
		}
		defer sslCertificate.Close()
		fieldWriter, err := bufferWriter.CreateFormFile("certificate", sslPublicKey)
		if err != nil {
			return nil, err
		}
		if _, err = io.Copy(fieldWriter, sslCertificate); err != nil {
			return nil, err
		}
	}
	err := writeFieldString(bufferWriter, "url", url)
	if err != nil {
		return nil, err
	}
	bufferWriter.Close()
	return f.newPostSendTypeBytes(f.setWebhookURL, buf.Bytes(), bufferWriter.FormDataContentType(), nil)
}

func (f *RequestFactory) NewUploadPhoto(chatID string, photo string, caption string, disableNotification bool, replyToMessageID int64, replyMarkup interface{}, callback OnSentCallback) (*SendType, error) {
	var buf bytes.Buffer
	bufferWriter := multipart.NewWriter(&buf)
	if len(photo) <= 0 {
		return nil, errors.New("No photo to upload")
	}
	photoFile, err := os.Open(photo)
	if err != nil {
		return nil, err
	}
	defer photoFile.Close()
	fieldWriter, err := bufferWriter.CreateFormFile("photo", photo)
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(fieldWriter, photoFile); err != nil {
		return nil, err
	}

	err = writeFieldString(bufferWriter, "chat_id", chatID)
	if err != nil {
		return nil, err
	}

	err = writeFieldString(bufferWriter, "caption", caption)
	if err != nil {
		return nil, err
	}

	replyMarkupSerialized, err := json.Marshal(replyMarkup)
	if err != nil {
		return nil, err
	}

	err = writeFieldString(bufferWriter, "disable_notification", fmt.Sprintf("%v", disableNotification))
	if err != nil {
		return nil, err
	}

	err = writeFieldString(bufferWriter, "reply_to_message_id", fmt.Sprintf("%v", replyToMessageID))
	if err != nil {
		return nil, err
	}

	err = writeFieldBytes(bufferWriter, "reply_markup", replyMarkupSerialized)
	if err != nil {
		return nil, err
	}

	bufferWriter.Close()
	return f.newPostSendTypeBytes(f.sendPhotoURL, buf.Bytes(), bufferWriter.FormDataContentType(), callback)
}

func (f *RequestFactory) NewResendPhoto(chatID string, photo string, caption string, disableNotification bool, replyToMessageID int64, replyMarkup interface{}, callback OnSentCallback) (*SendType, error) {
	sendPhotoRequest := send_requests.SendPhoto{
		ChatID:              chatID,
		Photo:               photo,
		DisableNotification: disableNotification,
		ReplyToMessageID:    replyToMessageID,
		ReplyMarkup:         replyMarkup}
	return f.newPostSendType(f.sendPhotoURL, sendPhotoRequest, content_type_application_json, callback)
}

func (f *RequestFactory) NewSendMessage(chatID string, message string, parseMode byte, disableWebPreview bool, disableNotifications bool, replyToMessageID int64, markup interface{}) (*SendType, error) {
	var parseModeValue string
	switch parseMode {
	case PARSE_MODE_HTML:
		parseModeValue = parse_mode_html
	case PARSE_MODE_MARKDOWN:
		parseModeValue = parse_mode_markdown
	}
	sendMessage := send_requests.SendMessageType{
		ChatID:                chatID,
		Text:                  message,
		ParseMode:             parseModeValue,
		DisableWebPagePreview: disableWebPreview,
		DisableNotification:   disableNotifications,
		ReplyToMessageID:      replyToMessageID,
		ReplyMarkup:           markup}
	return f.newPostSendType(f.sendMessageURL, sendMessage, content_type_application_json, nil)
}

func (f *RequestFactory) NewGetUpdates(offset int64, limit int64, timeout int64) (*SendType, error) {
	val := send_requests.GetUpdatesType{
		Offset:  offset,
		Limit:   limit,
		Timeout: timeout}
	return f.newPostSendType(f.getUpdatesURL, val, content_type_application_json, nil)
}

func (f *RequestFactory) String() string {
	return f.getUpdatesURL + "\n" + f.sendMessageURL + "\n" + f.sendStickerURL
}

func NewRequestFactory(botToken string, logger telebot.Logger) *RequestFactory {
	botRequestUrl := fmt.Sprintf(bot_query_fmt, botToken)
	var factory RequestFactory
	factory.sendMessageURL = fmt.Sprintf(cmd_send_message, botRequestUrl)
	factory.sendStickerURL = fmt.Sprintf(cmd_send_sticker, botRequestUrl)
	factory.getUpdatesURL = fmt.Sprintf(cmd_get_updates, botRequestUrl)
	factory.setWebhookURL = fmt.Sprintf(cmd_set_web_hook, botRequestUrl)
	factory.sendPhotoURL = fmt.Sprintf(cmd_send_photo, botRequestUrl)

	factory.defaultCallback = func(result *receive.SendResult) {
		if result.Ok {
			return
		}
		logger.Errorf("Failed to send response. Received error: code - '%d', description '%s'.", result.ErrorCode, result.Description)
	}

	return &factory
}
