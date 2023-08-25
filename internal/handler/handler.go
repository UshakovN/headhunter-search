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
	ctx               context.Context
	bot               telegram.Bot
	fetcher           fetcher.Fetcher
	storage           storage.Storage
	taskQueue         task.Queue
	chatsTimer        timer.Timer[int64]
	chatsTrees        *chats.Trees
	subVacancies      cache.MemCache[int64, *subVacancy]
	pendingChats      cache.KeyCache[int64]
	sentSubscriptions cache.KeyCache[string]
	sentVacancies     cache.KeyCache[string]
}

func NewHandler(ctx context.Context, bot telegram.Bot, fetcher fetcher.Fetcher, storage storage.Storage) (*Handler, error) {
	const workers = 100

	h := &Handler{
		ctx:               ctx,
		bot:               bot,
		fetcher:           fetcher,
		storage:           storage,
		taskQueue:         task.NewQueue(workers),
		chatsTimer:        timer.NewTimer[int64](),
		subVacancies:      cache.NewMemCache[int64, *subVacancy](),
		pendingChats:      cache.NewKeyCache[int64](),
		sentSubscriptions: cache.NewKeyCache[string](),
		sentVacancies:     cache.NewKeyCache[string](),
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
	if err := h.setSentSubscriptions(ctx); err != nil {
		return fmt.Errorf("cannot set sent subscriptions: %v", err)
	}
	if err := h.setSentVacancies(ctx); err != nil {
		return fmt.Errorf("cannot set sent vacancies: %v", err)
	}
	h.setChatsTrees()

	return nil
}

func (h *Handler) setSentVacancies(ctx context.Context) error {
	vacancies, err := h.storage.SentVacancies(ctx)
	if err != nil {
		return fmt.Errorf("cannot got sent vacancies from storage: %v", err)
	}
	m := cache.NewKeyCache[string]()

	for _, v := range vacancies {
		m.Put(v.VacancyID)
	}
	h.sentVacancies = m

	return nil
}

func (h *Handler) setSentSubscriptions(ctx context.Context) error {
	states, err := h.storage.SentSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("cannot got subscriptions states from storage: %v", err)
	}
	m := cache.NewKeyCache[string]()

	for _, s := range states {
		m.Put(s.SubscriptionID)
	}
	h.sentSubscriptions = m

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
		if err := h.storage.UsersSubscriptions(ctx, func(s *model.Subscription) {
			g.Go(func() error {
				if err := h.sendVacanciesForSubscription(ctx, s); err != nil {
					return fmt.Errorf("cannot send vacancies for subscription: %v", err)
				}
				return nil
			})
		}); err != nil {
			log.Errorf("cannot got users subscriptions from storage: %v. sleep %s before next handling...", err, errWait.String())
			time.Sleep(errWait)
			continue
		}
		if err := g.Wait(); err != nil {
			log.Errorf("cannot handle users subscriptions from storage: %v. sleep %s before next handling...", err, errWait.String())
			time.Sleep(errWait)
			continue
		}
		log.Infof("subscriptions handled. sleep %s before next handling...", okWait.String())
		time.Sleep(okWait)
	}
}

func (h *Handler) sendVacanciesForSubscription(ctx context.Context, s *model.Subscription) error {
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
		// wait timer for chat id
		h.chatsTimer.Wait(s.ChatID)

		// if chat id exist in pending chats
		if h.pendingChats.Exist(s.ChatID) {
			return nil
		}
		if h.sentVacancies.Exist(item.Id) {
			continue
		}
		msg := newVacancyMessage(s.ChatID, item)

		if _, err = h.bot.SendMessage(msg); err != nil {
			return fmt.Errorf("cannot send vacancy telegram bot message: %v", err)
		}
		h.sentVacancies.Put(item.Id)
		h.sentSubscriptions.Put(s.SubscriptionID)

		if err = h.storage.PutSentVacancy(ctx, &model.SentVacancy{
			SentID:    utils.NewUUID(),
			VacancyID: item.Id,
			ChatID:    s.ChatID,
			CreatedAt: utils.NowTimeUTC(),
		}); err != nil {
			return fmt.Errorf("cannot put sent vacancy to storage: %v", err)
		}
		if err = h.storage.PutSentSubscription(ctx, &model.SentSubscription{
			SentID:         utils.NewUUID(),
			SubscriptionID: s.SubscriptionID,
			CreatedAt:      utils.NowTimeUTC(),
		}); err != nil {
			return fmt.Errorf("cannot put sent subscription to storage: %v", err)
		}
		// set timer for chat id
		h.chatsTimer.Set(s.ChatID, messageTimeout)
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
