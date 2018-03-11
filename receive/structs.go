package receive

const (
	ENTITY_TYPE_MENTION     = "mention"
	ENTITY_TYPE_HASH_TAG    = "hashtag"
	ENTITY_TYPE_BOT_COMMAND = "bot_command"
	ENTITY_TYPE_URL         = "url"
	ENTITY_TYPE_EMAIL       = "email"
	ENTITY_TYPE_BOLD        = "bold"
	ENTITY_TYPE_ITALIC      = "italic"
	ENTITY_TYPE_CODE        = "code"
	ENTITY_TYPE_PRE         = "pre"
	ENTITY_TYPE_TEXT_LINK   = "text_link"
)

type MessageEntityType struct {
	Type   string `json:"type"`
	Offset int64  `json:"offset"`
	Length int64  `json:"length"`
	Url    string `json:"url,omitempty"`
}

type PhotoSizeType struct {
	ID     string `json:"file_id"`
	Width  int64  `json:"width"`
	Height int64  `json:"height"`
	Size   int64  `json:"file_size,omitempty"`
}

type AudioType struct {
	ID        string `json:"file_id"`
	Duration  int64  `json:"duration"`
	Performer string `json:"performer,omitempty"`
	Title     string `json:"title,omitempty"`
	MimeType  string `json:"mime_type,omitempty"`
	Size      int64  `json:"file_size,omitempty"`
}

type DocumentType struct {
	ID       string         `json:"file_id"`
	Thumb    *PhotoSizeType `json:"thumb,omitempty"`
	FileName string         `json:"file_name,omitempty"`
	MimeType string         `json:"mime_type,omitempty"`
	Size     int64          `json:"file_size,omitempty"`
}

type StickerType struct {
	PhotoSizeType
	Emoji   *string        `json:"emoji,omitempty"`
	SetName *string        `json:"set_name,omitempty"`
	Thumb   *PhotoSizeType `json:"thumb,omitempty"`
}

type VideoType struct {
	PhotoSizeType
	Duration int64          `json:"duration"`
	Thumb    *PhotoSizeType `json:"thumb,omitempty"`
	MimeType string         `json:"mime_type,omitempty"`
}

type VoiceType struct {
	ID       string `json:"file_id"`
	Duration int64  `json:"duration"`
	MimeType string `json:"mime_type,omitempty"`
	Size     int64  `json:"file_size,omitempty"`
}

type ContactType struct {
	Phone     string `json:"phone_number"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	UserID    int64  `json:"user_id,omitempty"`
}

type LocationType struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type VenueType struct {
	Location     LocationType `json:"location"`
	Title        string       `json:"title"`
	Address      string       `json:"address"`
	FoursquareID string       `json:"foursquare_id,omitempty"`
}

type UserType struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	UserName  string `json:"username,omitempty"`
}

const (
	CHAT_TYPE_PRIVATE = "private"
)

type ChatType struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title,omitempty"`
	UserName  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

type BasicMessageType struct {
	ID                    int64               `json:"message_id"`
	From                  *UserType           `json:"from,omitempty"`
	Date                  int64               `json:"date"`
	Chat                  *ChatType           `json:"chat"`
	ForwardFrom           *UserType           `json:"forward_from,omitempty"`
	ForwardDate           int64               `json:"forward_date,omitempty"`
	Text                  string              `json:"text,omitempty"`
	Entities              []MessageEntityType `json:"entities,omitempty"`
	Audio                 *AudioType          `json:"audio,omitempty"`
	Document              *DocumentType       `json:"document,omitempty"`
	Photos                []PhotoSizeType     `json:"photo,omitempty"`
	Sticker               *StickerType        `json:"sticker,omitempty"`
	Video                 *VideoType          `json:"video,omitempty"`
	Voice                 *VoiceType          `json:"voice,omitempty"`
	Caption               string              `json:"caption,omitempty"`
	Contact               *ContactType        `json:"contact,omitempty"`
	Location              *LocationType       `json:"location,omitempty"`
	Venue                 *VenueType          `json:"venue,omitempty"`
	NewChatMember         *UserType           `json:"new_chat_member,omitempty"`
	LeftChatMember        *UserType           `json:"left_chat_member,omitempty"`
	NewChatTitle          string              `json:"new_chat_title,omitempty"`
	NewChatPhoto          []PhotoSizeType     `json:"new_chat_photo,omitempty"`
	DeleteChatPhoto       bool                `json:"delete_chat_photo,omitempty"`
	GroupChatCreated      bool                `json:"group_chat_created,omitempty"`
	SuperGroupChatCreated bool                `json:"supergroup_chat_create,omitempty"`
	ChannelChatCreated    bool                `json:"channel_chat_created,omitempty"`
	MigrateToChatID       int64               `json:"migrate_to_chat_id,omitempty"`
	MigrateFromChatID     int64               `json:"migrate_from_chat_id,omitempty"`
}

type MessageType struct {
	BasicMessageType
	ReplyTo       *BasicMessageType `json:"reply_to_message,omitempty"`
	PinnedMessage *BasicMessageType `json:"pinned_message,omitempty"`
}

type ChosenInlineResultType struct {
	ID              int64         `json:"result_id"`
	From            UserType      `json:"from"`
	Location        *LocationType `json:"location,omitempty"`
	InlineMessageID int64         `json:"inline_message_id,omitempty"`
	Query           string        `json:"query"`
}

type InlineQueryType struct {
	ID       int64         `json:"id"`
	From     UserType      `json:"from"`
	Location *LocationType `json:"location,omitempty"`
	Query    string        `json:"query"`
	Offset   string        `json:"offset"`
}

type CallbackQueryType struct {
	ID              string       `json:"id"`
	From            UserType     `json:"from"`
	Message         *MessageType `json:"message,omitempty"`
	InlineMessageID string       `json:"inline_message_id,omitempty"`
	Data            string       `json:"data"`
}

type UpdateType struct {
	ID                 int64                   `json:"update_id"`
	Message            *MessageType            `json:"message,omitempty"`
	EditedMessage      *MessageType            `json:"edited_message,omitempty"`
	ChannelPost        *MessageType            `json:"channel_post,omitempty"`
	EditedChannelPost  *MessageType            `json:"edited_channel_post,omitempty"`
	InlineQuery        *InlineQueryType        `json:"inline_query,omitempty"`
	ChosenInlineResult *ChosenInlineResultType `json:"chosen_inline_result,omitempty"`
	CallbackQuery      *CallbackQueryType      `json:"callback_query,omitempty"`
}

type UpdateResultType struct {
	Ok      bool         `json:"ok"`
	Updates []UpdateType `json:"result"`
}

type SendResult struct {
	Ok          bool        `json:"ok"`
	ErrorCode   int64       `json:"error_code,omitempty"`
	Description *string      `json:"description,omitempty"`
	Result      *MessageType `json:"result,omitempty"`
}
