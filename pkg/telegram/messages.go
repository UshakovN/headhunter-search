package telegram

import (
  "strings"

  tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ParseMode string

const ParseModeMarkdownV2 ParseMode = "MarkdownV2"

type Message struct {
  ChatID   int64
  UserID   int64
  UserName string
  Text     string
  Command  string
  Date     int64
}

type SendMessage struct {
  ChatID    int64
  Text      string
  ParseMode ParseMode
}

func (m *Message) IsCommand() bool {
  return m.Command != ""
}

func (m *Message) IsText() bool {
  return !m.IsCommand()
}

func toMessage(m *tg.Message) *Message {
  var (
    chatID   int64
    userID   int64
    userName string
    date     int64
  )
  if chat := m.Chat; chat != nil {
    chatID = chat.ID
  }
  if from := m.From; from != nil {
    userID = from.ID
    userName = from.UserName
  }
  return &Message{
    ChatID:   chatID,
    UserID:   userID,
    UserName: userName,
    Text:     strings.TrimSpace(m.Text),
    Command:  m.Command(),
    Date:     date,
  }
}
