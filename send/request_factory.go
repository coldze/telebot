package send

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/coldze/primitives/custom_error"
	"github.com/coldze/primitives/logs"
	"github.com/coldze/telebot/receive"
	"github.com/coldze/telebot/send/internal/requests"
	"github.com/coldze/telebot/send/requests"
)

const (
	PARSE_MODE_HTML     = "HTML"
	PARSE_MODE_MARKDOWN = "Markdown"
)

const (
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
	sendStickerURL         string
	sendMessageURL         string
	sendPhotoURL           string
	getUpdatesURL          string
	answerCallbackQueryURL string
	defaultCallback        OnSentCallback
	setWebhookURL          string
}

func writeFieldString(writer *multipart.Writer, fieldName string, value string) custom_error.CustomError {
	customErr := writeFieldBytes(writer, fieldName, []byte(value))
	if customErr == nil {
		return nil
	}
	return custom_error.NewErrorf(customErr, "Failed to write field string")
}

func writeFieldBytes(writer *multipart.Writer, fieldName string, value []byte) custom_error.CustomError {
	fieldWriter, err := writer.CreateFormField(fieldName)
	if err != nil {
		return custom_error.MakeErrorf("Failed to create form field. Error: %v", err)
	}
	_, err = fieldWriter.Write(value)
	if err == nil {
		return nil
	}
	return custom_error.MakeErrorf("Failed to write field. Error: %v", err)
}

func (f *RequestFactory) NewSendRaw(url string, message interface{}) ([]*SendType, custom_error.CustomError) {
	request, err := json.Marshal(message)
	if err != nil {
		return nil, custom_error.MakeErrorf("Failed to marshal message. Error: %v", err)
	}
	return []*SendType{
		&SendType{
			URL:        f.sendStickerURL,
			Parameters: request,
		},
	}, nil
}

func (f *RequestFactory) newPostSendType(url string, message interface{}, contentType string, callback OnSentCallback) ([]*SendType, custom_error.CustomError) {
	requestMessage, err := json.Marshal(message)
	if err != nil {
		return nil, custom_error.MakeErrorf("Failed to marshal message. Error: %v", err)
	}
	res, customErr := f.newPostSendTypeBytes(url, requestMessage, contentType, callback)
	if customErr == nil {
		return res, nil
	}
	return nil, custom_error.NewErrorf(customErr, "Failed to post-send")
}

func (f *RequestFactory) newPostSendTypeBytes(url string, message []byte, contentType string, callback OnSentCallback) ([]*SendType, custom_error.CustomError) {
	cb := callback
	if cb == nil {
		cb = f.defaultCallback
	}
	return []*SendType{
		&SendType{
			URL:         url,
			Parameters:  message,
			Type:        SEND_TYPE_POST,
			ContentType: contentType,
			Callback:    cb,
		},
	}, nil
}

func getChatIDString(chatID interface{}) (string, custom_error.CustomError) {
	if chatID == nil {
		return "", custom_error.MakeErrorf("Failed to convert chat-ID. Empty chat-id specified.")
	}
	v, ok := chatID.(string)
	if ok {
		return v, nil
	}
	vInt, ok := chatID.(int64)
	if ok {
		return fmt.Sprintf("%d", vInt), nil
	}
	return "", custom_error.MakeErrorf("Failed to convert chat-ID. Unknown type: %T", chatID)
}

func (f *RequestFactory) newFileUpload(url string, r *requests.SendFile, callback OnSentCallback) ([]*SendType, custom_error.CustomError) {
	chatID, customErr := getChatIDString(r.ChatID)
	if customErr != nil {
		return nil, custom_error.NewErrorf(customErr, "Failed to upload file.")
	}
	var buf bytes.Buffer
	bufferWriter := multipart.NewWriter(&buf)
	if len(r.FileName) <= 0 {
		return nil, custom_error.MakeErrorf("No file to upload")
	}
	uploadingFile, err := os.Open(r.FileName)
	if err != nil {
		return nil, custom_error.MakeErrorf("Failed to open file. Fieldname: '%v'. Filename: '%v'. Error: %v", r.FieldName, r.FileName, err)
	}
	defer uploadingFile.Close()
	fieldWriter, err := bufferWriter.CreateFormFile(r.FieldName, r.FileName)
	if err != nil {
		return nil, custom_error.MakeErrorf("Failed to create form from file. Fieldname: '%v'. Filename: '%v'. Error: %v", r.FieldName, r.FileName, err)
	}
	if _, err = io.Copy(fieldWriter, uploadingFile); err != nil {
		return nil, custom_error.MakeErrorf("Failed to create form from file. Fieldname: '%v'. Filename: '%v'. Error: %v", r.FieldName, r.FileName, err)
	}
	getChatIDString(r.ChatID)

	customErr = writeFieldString(bufferWriter, "chat_id", chatID)
	if customErr != nil {
		return nil, custom_error.NewErrorf(customErr, "Failed to write chat_id.")
	}

	customErr = writeFieldString(bufferWriter, "caption", r.Caption)
	if customErr != nil {
		return nil, custom_error.NewErrorf(customErr, "Failed to write caption.")
	}

	replyMarkupSerialized, err := json.Marshal(r.ReplyMarkup)
	if err != nil {
		return nil, custom_error.MakeErrorf("Failed to marshal marked up reply. Error: %v", err)
	}

	customErr = writeFieldString(bufferWriter, "disable_notification", fmt.Sprintf("%v", r.DisableNotifications))
	if customErr != nil {
		return nil, custom_error.NewErrorf(customErr, "Failed to write disable_notification.")
	}

	customErr = writeFieldString(bufferWriter, "reply_to_message_id", fmt.Sprintf("%v", r.ReplyToMessageID))
	if customErr != nil {
		return nil, custom_error.NewErrorf(customErr, "Failed to write reply_to_message_id.")
	}

	customErr = writeFieldBytes(bufferWriter, "reply_markup", replyMarkupSerialized)
	if customErr != nil {
		return nil, custom_error.NewErrorf(customErr, "Failed to write reply_markup.")
	}

	bufferWriter.Close()
	res, customErr := f.newPostSendTypeBytes(url, buf.Bytes(), bufferWriter.FormDataContentType(), callback)
	if customErr == nil {
		return res, nil
	}
	return nil, custom_error.NewErrorf(customErr, "Failed to post-send bytes.")
}

func (f *RequestFactory) NewSendSticker(r *requests.SendSticker, callback OnSentCallback) ([]*SendType, custom_error.CustomError) {
	if r == nil {
		return nil, custom_error.MakeErrorf("Failed to send sticker. Request is nil.")
	}
	res, customErr := f.newPostSendType(f.sendStickerURL, r, content_type_application_json, callback)
	if customErr == nil {
		return res, nil
	}
	return nil, custom_error.NewErrorf(customErr, "Failed to send sticker.")
}

func (f *RequestFactory) NewUnsubscribe() ([]*SendType, custom_error.CustomError) {
	res, customErr := f.NewSubscribe("", "")
	if customErr == nil {
		return res, nil
	}
	return nil, custom_error.NewErrorf(customErr, "Failed to unsubscribe")
}

func (f *RequestFactory) NewSubscribe(url string, sslPublicKey string) ([]*SendType, custom_error.CustomError) {
	var buf bytes.Buffer
	bufferWriter := multipart.NewWriter(&buf)
	if len(sslPublicKey) > 0 {
		sslCertificate, err := os.Open(sslPublicKey)
		if err != nil {
			return nil, custom_error.MakeErrorf("Failed to open file with public-key: '%v'. Error: %v", sslPublicKey, err)
		}
		defer sslCertificate.Close()
		fieldWriter, err := bufferWriter.CreateFormFile("certificate", sslPublicKey)
		if err != nil {
			return nil, custom_error.MakeErrorf("Failed to create field certificate from public-key file. Error: %v", err)
		}
		if _, err = io.Copy(fieldWriter, sslCertificate); err != nil {
			return nil, custom_error.MakeErrorf("Failed to write certificate. Error: %v", err)
		}
	}
	customErr := writeFieldString(bufferWriter, "url", url)
	if customErr != nil {
		return nil, custom_error.NewErrorf(customErr, "Failed to write url field.")
	}
	bufferWriter.Close()
	res, customErr := f.newPostSendTypeBytes(f.setWebhookURL, buf.Bytes(), bufferWriter.FormDataContentType(), nil)
	if customErr == nil {
		return res, nil
	}
	return nil, custom_error.NewErrorf(customErr, "Failed to subscribe.")
}

func (f *RequestFactory) NewUploadPhoto(r *requests.SendFileBase, callback OnSentCallback) ([]*SendType, custom_error.CustomError) {
	if r == nil {
		return nil, custom_error.MakeErrorf("Failed to create upload photo. Request is nil.")
	}
	photoRequest := requests.SendFile{
		FieldName:    "photo",
		SendFileBase: *r,
	}
	res, customErr := f.newFileUpload(f.sendPhotoURL, &photoRequest, callback)
	if customErr == nil {
		return res, nil
	}
	return nil, custom_error.NewErrorf(customErr, "Failed to create upload photo.")
}

func (f *RequestFactory) NewSendUploadedPhoto(r *requests.SendUploadedPhoto, callback OnSentCallback) ([]*SendType, custom_error.CustomError) {
	if r == nil {
		return nil, custom_error.MakeErrorf("Failed to create send uploaded photo. Request is nil.")
	}
	res, customErr := f.newPostSendType(f.sendPhotoURL, r, content_type_application_json, callback)
	if customErr == nil {
		return res, nil
	}
	return nil, custom_error.NewErrorf(customErr, "Failed to create send uploaded photo.")
}

func (f *RequestFactory) NewSendMessage(r *requests.SendMessage, callback OnSentCallback) ([]*SendType, custom_error.CustomError) {
	if r == nil {
		return nil, custom_error.MakeErrorf("Failed to create send message. Request is nil.")
	}
	res, customErr := f.newPostSendType(f.sendMessageURL, r, content_type_application_json, callback)
	if customErr == nil {
		return res, nil
	}
	return nil, custom_error.NewErrorf(customErr, "Failed to create send message.")
}

func (f *RequestFactory) NewAnswerCallbackQuery(message *requests.AnswerCallbackQuery) ([]*SendType, custom_error.CustomError) {
	res, customErr := f.newPostSendType(f.answerCallbackQueryURL, message, content_type_application_json, nil)
	if customErr == nil {
		return res, nil
	}
	return nil, custom_error.NewErrorf(customErr, "Failed to create answer callback query.")
}

func (f *RequestFactory) NewGetUpdates(offset int64, limit int64, timeout int64) ([]*SendType, custom_error.CustomError) {
	val := send_requests.GetUpdates{
		Offset:  offset,
		Limit:   limit,
		Timeout: timeout}
	res, customErr := f.newPostSendType(f.getUpdatesURL, val, content_type_application_json, nil)
	if customErr == nil {
		return res, nil
	}
	return nil, custom_error.NewErrorf(customErr, "Failed to create get updates.")
}

func (f *RequestFactory) String() string {
	return f.getUpdatesURL + "\n" + f.sendMessageURL + "\n" + f.sendStickerURL
}

func NewRequestFactory(botToken string, logger logs.Logger) *RequestFactory {
	botRequestUrl := fmt.Sprintf(bot_query_fmt, botToken)
	var factory RequestFactory
	factory.sendMessageURL = fmt.Sprintf(cmd_send_message, botRequestUrl)
	factory.sendStickerURL = fmt.Sprintf(cmd_send_sticker, botRequestUrl)
	factory.getUpdatesURL = fmt.Sprintf(cmd_get_updates, botRequestUrl)
	factory.setWebhookURL = fmt.Sprintf(cmd_set_web_hook, botRequestUrl)
	factory.sendPhotoURL = fmt.Sprintf(cmd_send_photo, botRequestUrl)
	factory.answerCallbackQueryURL = fmt.Sprintf(cmd_answer_callback_query, botRequestUrl)

	factory.defaultCallback = func(result *receive.SendResult, err custom_error.CustomError) {
		if err != nil {
			logger.Errorf("Failed to send response. Internal error: %v", err)
			return
		}
		if result == nil {
			logger.Warningf("No result-sending information provided.")
			return
		}
		if result.Ok {
			return
		}
		logger.Errorf("Failed to send response. Received error: code - '%d', description '%s'.", result.ErrorCode, result.Description)
	}

	return &factory
}
