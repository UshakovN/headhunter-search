package telegram

import (
	"strings"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type parseMode string

const (
	HTMLParseMode       parseMode = "HTML"
	MarkdownV2ParseMode parseMode = "MarkdownV2"
)

func (p parseMode) String() string {
	return string(p)
}

type Message struct {
	MessageID    int64
	ChatID       int64
	UserID       int64
	UserName     string
	Text         string
	Command      string
	Date         int64
	fromCallback bool
}

type SendMessage struct {
	ChatID   int64
	Text     string
	Keyboard *InlineKeyboard
}

type EditMessage struct {
	MessageID int64
	ChatID    int64
	Text      string
	Keyboard  *InlineKeyboard
}

func (m *Message) IsCommand() bool {
	return m.Command != ""
}

func (m *Message) IsText() bool {
	return !m.IsCommand()
}

func (m *Message) FromCallback() bool {
	return m.fromCallback
}

func apiCallbackToModel(cb *tg.CallbackQuery) *Message {
	var (
		chatID   int64
		userID   int64
		userName string
		date     int64
	)
	if m := cb.Message; m != nil {
		if chat := m.Chat; chat != nil {
			chatID = chat.ID
		}
		if from := m.From; from != nil {
			userID = from.ID
			userName = from.UserName
		}
		date = int64(m.Date)
	}
	data := strings.TrimSpace(cb.Data)

	return &Message{
		ChatID:       chatID,
		UserID:       userID,
		UserName:     userName,
		Text:         data,
		Command:      strings.TrimPrefix(data, "/"),
		Date:         date,
		fromCallback: true,
	}
}

func apiMessageToModel(msg *tg.Message) *Message {
	var (
		chatID   int64
		userID   int64
		userName string
	)
	if chat := msg.Chat; chat != nil {
		chatID = chat.ID
	}
	if from := msg.From; from != nil {
		userID = from.ID
		userName = from.UserName
	}
	return &Message{
		MessageID: int64(msg.MessageID),
		ChatID:    chatID,
		UserID:    userID,
		UserName:  userName,
		Text:      strings.TrimSpace(msg.Text),
		Command:   msg.Command(),
		Date:      int64(msg.Date),
	}
}

func (m *SendMessage) apiInlineKeyboard() any {
	if m.Keyboard == nil {
		return nil
	}
	return m.Keyboard.markup
}

func (m *EditMessage) apiInlineKeyboard() *tg.InlineKeyboardMarkup {
	if m.Keyboard == nil {
		return nil
	}
	return &m.Keyboard.markup
}

func (m *SendMessage) ToEditMessage(messageID int64) *EditMessage {
	return &EditMessage{
		MessageID: messageID,
		ChatID:    m.ChatID,
		Text:      m.Text,
		Keyboard:  m.Keyboard,
	}
}
