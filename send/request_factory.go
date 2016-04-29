package send

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	sendStickerURL string
	sendMessageURL string
	getUpdatesURL  string
	SetWebhookURL  string
}

func (f *RequestFactory) NewSendRaw(url string, message interface{}) (*SendType, error) {
	request, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	return &SendType{URL: f.sendStickerURL, Parameters: request}, nil
}

func (f *RequestFactory) newPostSendType(url string, message interface{}, contentType string) (*SendType, error) {
	params, ok := message.([]byte)
	if !ok {
		return nil, fmt.Errorf("Invalid input message type. Expected []byte, got: %T", message)
	}
	return &SendType{URL: url, Parameters: params, Type: SEND_TYPE_POST, ContentType: contentType}, nil
}

func (f *RequestFactory) NewSendSticker(chatID string, sticker string, notify bool, replyToMessageID int64, markup interface{}) (*SendType, error) {
	stickerMessage := send_requests.SendSticker{
		ChatID:              chatID,
		Sticker:             sticker,
		DisableNotification: notify,
		ReplyToMessageID:    replyToMessageID}
	return f.newPostSendType(f.sendStickerURL, stickerMessage, content_type_application_json)
}

func (f *RequestFactory) NewSignUp(url string, sslPublicKey string) (*SendType, error) {
	var buf bytes.Buffer
	bufferWriter := multipart.NewWriter(&buf)
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
	if fieldWriter, err = bufferWriter.CreateFormField("url"); err != nil {
		return nil, err
	}
	if _, err = fieldWriter.Write([]byte(url)); err != nil {
		return nil, err
	}
	bufferWriter.Close()
	sendType, err := f.newPostSendType(f.SetWebhookURL, buf.Bytes(), content_type_application_json)
	if err != nil {
		return nil, err
	}
	sendType.ContentType = bufferWriter.FormDataContentType()
	return sendType, nil
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
	return f.newPostSendType(f.sendMessageURL, sendMessage, content_type_application_json)
}

func (f *RequestFactory) NewGetUpdates(offset int64, limit int64, timeout int64) (*SendType, error) {
	val := send_requests.GetUpdatesType{
		Offset:  offset,
		Limit:   limit,
		Timeout: timeout}
	return f.newPostSendType(f.getUpdatesURL, val, content_type_application_json)
}

func (f *RequestFactory) String() string {
	return f.getUpdatesURL + "\n" + f.sendMessageURL + "\n" + f.sendStickerURL
}

func NewRequestFactory(botToken string) *RequestFactory {
	botRequestUrl := fmt.Sprintf(bot_query_fmt, botToken)
	var factory RequestFactory
	factory.sendMessageURL = fmt.Sprintf(cmd_send_message, botRequestUrl)
	factory.sendStickerURL = fmt.Sprintf(cmd_send_sticker, botRequestUrl)
	factory.getUpdatesURL = fmt.Sprintf(cmd_get_updates, botRequestUrl)
	factory.SetWebhookURL = fmt.Sprintf(cmd_set_web_hook, botRequestUrl)
	return &factory
}
