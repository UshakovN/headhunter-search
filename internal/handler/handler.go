package handler

import (
	"context"
	"fmt"
	"main/internal/chats"
	"main/internal/fetcher"
	"main/internal/model"
	"main/internal/storage"
	"main/internal/task"
	"main/pkg/cache"
	"main/pkg/telegram"
	"main/pkg/timer"
	"main/pkg/utils"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Handler struct {
	ctx           context.Context
	bot           telegram.Bot
	fetcher       fetcher.Fetcher
	storage       storage.Storage
	taskQueue     task.Queue
	chatsTrees    *chats.Trees
	chatsTimers   timer.Timer[int64]
	chatsPending  cache.KeyCache[int64]
	chatsSubVacs  cache.MemCache[int64, *vacancy]
	chatsSentVacs cache.MemCache[int64, cache.KeyCache[string]]
}

func NewHandler(ctx context.Context, bot telegram.Bot, fetcher fetcher.Fetcher, storage storage.Storage) (*Handler, error) {
	const workers = 100

	h := &Handler{
		ctx:          ctx,
		bot:          bot,
		fetcher:      fetcher,
		storage:      storage,
		taskQueue:    task.NewQueue(workers),
		chatsTimers:  timer.NewTimer[int64](),
		chatsSubVacs: cache.NewMemCache[int64, *vacancy](),
		chatsPending: cache.NewKeyCache[int64](),
	}
	if err := h.prepareComponents(ctx); err != nil {
		return nil, fmt.Errorf("handler cannot prepare components: %v", err)
	}
	return h, nil
}

func (h *Handler) prepareComponents(ctx context.Context) error {
	if err := h.bot.Start(); err != nil {
		return fmt.Errorf("telegram bot cannot start: %v", err)
	}
	if err := h.setChatsSentVacs(ctx); err != nil {
		return fmt.Errorf("cannot set sent vacancies: %v", err)
	}
	h.setChatsTrees()

	return nil
}

func (h *Handler) setChatsSentVacs(ctx context.Context) error {
	vacancies, err := h.storage.SentVacancies(ctx)
	if err != nil {
		return fmt.Errorf("cannot got sent vacancies from storage: %v", err)
	}
	m := cache.NewMemCache[int64, cache.KeyCache[string]]()

	for _, v := range vacancies {
		m.GetPut(v.ChatID, cache.NewKeyCache[string]()).Put(v.VacancyID)
	}
	h.chatsSentVacs = m

	return nil
}

func (h *Handler) HandleTasksContinuously(ctx context.Context) {
	h.taskQueue.ContinuouslyHandle(ctx)
}

func (h *Handler) HandleSubscriptionsContinuously(ctx context.Context) {
	const (
		okWait     = 5 * time.Second
		errWait    = 10 * time.Second
		groupLimit = 100
	)
	g, _ := errgroup.WithContext(ctx)
	g.SetLimit(groupLimit)

	for {
		if err := h.storage.ChatsSubscriptions(ctx, func(s *model.ChatSubscription) {
			g.Go(func() error {
				if err := h.sendVacanciesForSubscription(ctx, s); err != nil {
					return fmt.Errorf("cannot send vacancies for subscription: %v", err)
				}
				return nil
			})
		}); err != nil {
			log.Errorf("cannot got users subscriptions from storage: %v. sleep %s before next handling", err, errWait.String())
			time.Sleep(errWait)
			continue
		}
		if err := g.Wait(); err != nil {
			log.Errorf("cannot handle users subscriptions from storage: %v. sleep %s before next handling", err, errWait.String())
			time.Sleep(errWait)
			continue
		}
		time.Sleep(okWait)
	}
}

func (h *Handler) sendVacanciesForSubscription(ctx context.Context, s *model.ChatSubscription) error {
	const (
		messageTimeout = 10 * time.Second
		requestPeriod  = 14
	)
	req := fetcher.NewVacanciesRequest(
		s.Keywords,
		s.Area,
		s.Experience,
		requestPeriod,
	)
	resp, err := h.fetcher.Fetch(ctx, req)
	if err != nil {
		return fmt.Errorf("cannot fetch vacancies for request: %v", err)
	}
	for _, item := range resp.Items {
		// if vacancy it is wrong
		if isWrongVacancy(item) {
			continue
		}
		// wait timer for chat id
		h.chatsTimers.Wait(s.ChatID)

		// if chat id exist in pending chats
		if h.chatsPending.Exist(s.ChatID) {
			return nil
		}
		// if vacancy id already sent to chat id
		if h.chatsSentVacs.Exist(s.ChatID) && h.chatsSentVacs.Get(s.ChatID).Exist(item.Id) {
			continue
		}
		msg := newVacancyMessage(s.ChatID, s.Keywords, item)

		if _, err = h.bot.SendMessage(msg); err != nil {
			return fmt.Errorf("cannot send vacancy telegram bot message: %v", err)
		}
		// put sent vacancy id for chat id
		h.chatsSentVacs.GetPut(s.ChatID, cache.NewKeyCache[string]()).Put(item.Id)

		if err = h.storage.PutSentVacancy(ctx, &model.ChatSentVacancy{
			VacancyID:      item.Id,
			SubscriptionID: s.SubscriptionID,
			CreatedAt:      utils.NowTimeUTC(),
		}); err != nil {
			return fmt.Errorf("cannot put sent vacancy to storage: %v", err)
		}
		// set timer for chat id
		h.chatsTimers.Set(s.ChatID, messageTimeout)
	}
	return nil
}

func (h *Handler) HandleMessagesContinuously(ctx context.Context) {
	h.bot.HandleMessages(func(m *telegram.Message) error {
		return h.HandleMessages(ctx, m)
	})
}

func (h *Handler) Shutdown() {
	h.bot.Shutdown()
}
