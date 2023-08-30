package telegram

import (
	"context"
	"fmt"
	"main/pkg/retries"
	"time"

	log "github.com/sirupsen/logrus"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	retryCount = 3
	retryWait  = time.Second
)

type Bot interface {
	Start() error
	StartWithWebhook(link string) error
	SendMessage(m *SendMessage, options ...MessageOption) (int64, error)
	EditMessage(m *EditMessage, options ...MessageOption) (int64, error)
	DeleteMessage(chatID int64, messageID int64) error
	HandleMessages(handler func(m *Message) error)
	Shutdown()
}

type bot struct {
	ctx     context.Context
	token   string
	api     *tg.BotAPI
	updates tg.UpdatesChannel
	started chan struct{}
}

func NewBot(ctx context.Context, token string) (Bot, error) {
	api, err := tg.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("cannot create new bot api: %v", err)
	}
	return &bot{
		ctx: ctx,
		api: api,
	}, nil
}

func (b *bot) Start() error {
	const timeout = 60

	b.updates = b.api.GetUpdatesChan(tg.UpdateConfig{
		Offset:  0,
		Limit:   0,
		Timeout: timeout,
	})
	return nil
}

func (b *bot) StartWithWebhook(link string) error {
	wh, err := tg.NewWebhook(link)
	if err != nil {
		return fmt.Errorf("cannot create new telegram webhook: %v", err)
	}
	if _, err := b.api.Request(wh); err != nil {
		return fmt.Errorf("cannot do request to webhook: %v", err)
	}
	info, err := b.api.GetWebhookInfo()
	if err != nil {
		return fmt.Errorf("cannot get webhook info: %v", err)
	}
	if info.LastErrorDate != 0 {
		return fmt.Errorf("telegram callback failed: %s", info.LastErrorMessage)
	}
	b.updates = b.api.ListenForWebhook("/" + b.token)

	return nil
}

func (b *bot) HandleMessages(handler func(m *Message) error) {
	var (
		m   *Message
		err error
	)
	for update := range b.updates {
		if msg := update.Message; msg != nil {
			m = apiMessageToModel(msg)
		}
		if cb := update.CallbackQuery; cb != nil {
			m = apiCallbackToModel(cb)
		}
		if m != nil {
			if err = handler(m); err != nil {
				log.Errorf("cannot handle telegram message: %v", err)
			}
		}
	}
}

type messageOption struct {
	parseMode parseMode
}

type MessageOption func(o *messageOption)

func WithParseMode(mode parseMode) MessageOption {
	return func(o *messageOption) {
		o.parseMode = mode
	}
}

func newMessageOption() *messageOption {
	return &messageOption{
		parseMode: HTMLParseMode,
	}
}

func callMessageOptions(options ...MessageOption) *messageOption {
	mo := newMessageOption()

	for _, option := range options {
		option(mo)
	}
	return mo
}

func (b *bot) SendMessage(m *SendMessage, options ...MessageOption) (int64, error) {
	var (
		msg tg.Message
		id  int64
		err error
	)
	mo := callMessageOptions(options...)

	err = retries.DoWithRetries(retryCount, retryWait, func() error {
		if msg, err = b.api.Send(tg.MessageConfig{
			BaseChat: tg.BaseChat{
				ChatID:      m.ChatID,
				ReplyMarkup: m.apiInlineKeyboard(),
			},
			Text:      m.Text,
			ParseMode: mo.parseMode.String(),
		}); err != nil {
			return fmt.Errorf("%w: cannot send telegram message: %v", retries.ErrDoRetry, err)
		}
		id = int64(msg.MessageID)
		return nil
	})
	return id, err
}

func (b *bot) EditMessage(m *EditMessage, options ...MessageOption) (int64, error) {
	var (
		msg tg.Message
		id  int64
		err error
	)
	mo := callMessageOptions(options...)

	err = retries.DoWithRetries(retryCount, retryWait, func() error {
		if msg, err = b.api.Send(tg.EditMessageTextConfig{
			BaseEdit: tg.BaseEdit{
				ChatID:      m.ChatID,
				MessageID:   int(m.MessageID),
				ReplyMarkup: m.apiInlineKeyboard(),
			},
			Text:      m.Text,
			ParseMode: mo.parseMode.String(),
		}); err != nil {
			return fmt.Errorf("%w: cannot edit telegram message: %v", retries.ErrDoRetry, err)
		}
		id = int64(msg.MessageID)
		return nil
	})
	return id, err
}

func (b *bot) DeleteMessage(chatID int64, messageID int64) error {
	return retries.DoWithRetries(retryCount, retryWait, func() error {
		if _, err := b.api.Request(tg.DeleteMessageConfig{
			ChatID:    chatID,
			MessageID: int(messageID),
		}); err != nil {
			return fmt.Errorf("%w: cannot delete telegram message: %v", retries.ErrDoRetry, err)
		}
		return nil
	})
}

func (b *bot) Shutdown() {
	b.api.StopReceivingUpdates()
}
