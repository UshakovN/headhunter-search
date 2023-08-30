package handler

import (
	"context"
	"fmt"
	"main/internal/chats"
	"main/internal/fetcher"
	"main/internal/model"
	"main/internal/storage"
	"main/pkg/cache"
	"main/pkg/schedule"
	"main/pkg/task"
	"main/pkg/telegram"
	"main/pkg/timer"
	"main/pkg/utils"
	"time"

	log "github.com/sirupsen/logrus"
)

type Handler struct {
	ctx           context.Context
	bot           telegram.Bot
	fetcher       fetcher.Fetcher
	storage       storage.Storage
	subTasks      task.Queue
	sendTasks     task.Queue
	chatsTrees    chats.Trees
	chatsTimers   cache.MemCache[int64, timer.RefreshTimer]
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
		subTasks:     task.NewQueue(workers),
		sendTasks:    task.NewQueue(workers),
		chatsTimers:  cache.NewMemCache[int64, timer.RefreshTimer](),
		chatsSubVacs: cache.NewMemCache[int64, *vacancy](),
		chatsPending: cache.NewKeyCache[int64](),
	}
	if err := h.prepareComponents(ctx); err != nil {
		return nil, fmt.Errorf("handler cannot prepare components: %v", err)
	}
	go h.subTasks.ContinuouslyHandle(ctx)
	go h.sendTasks.ContinuouslyHandle(ctx)

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

func (h *Handler) HandleSubscriptions(ctx context.Context) error {
	if err := h.storage.ChatsSubscriptions(ctx, func(sub *model.ChatSubscription) {
		h.sendTasks.Push(func() error {
			if err := h.sendSubscriptionVacancies(ctx, sub); err != nil {
				return fmt.Errorf("cannot send subscription vacancies: %v", err)
			}
			log.Infof("subscription %s for chat with id %d handled", sub.Keywords, sub.ChatID)
			return nil
		})
	}); err != nil {
		return fmt.Errorf("cannot got chats subscription from storage: %v", err)
	}
	log.Infof("chat subscriptions handled")
	return nil
}

func (h *Handler) HandleSubscriptionsContinuously(ctx context.Context) {
	schedule.DoWithSchedule("10s", "30s", true, func() error {
		log.Infof("scheduled handling subscriptions started")
		return h.HandleSubscriptions(ctx)
	})
}

func (h *Handler) fetchVacancies(ctx context.Context, s *model.ChatSubscription) ([]*fetcher.VacancyResponseItem, error) {
	const (
		maxDepth = 1000
		perPage  = 100
	)
	req := &fetcher.Request{
		Text:       s.Keywords,
		Area:       s.Area,
		Experience: s.Experience,
	}
	var (
		items []*fetcher.VacancyResponseItem
		page  int
	)
	for {
		resp, err := h.fetcher.Fetch(ctx, req.WithDefault().WithPaging(page, perPage))
		if err != nil {
			return nil, fmt.Errorf("cannot fetch vacancies for request: %v", err)
		}
		// collect response items
		if len(resp.Items) > 0 {
			items = append(items, resp.Items...)
		}
		// increment paging parameter
		page++
		// break loop if max depth exceeded
		if resp.Found == 0 || page >= resp.Pages || page*resp.PerPage >= maxDepth {
			break
		}
	}
	return items, nil
}

func (h *Handler) sendSubscriptionVacancies(ctx context.Context, s *model.ChatSubscription) error {
	const timeout = 15 * time.Second

	items, err := h.fetchVacancies(ctx, s)
	if err != nil {
		return fmt.Errorf("cannot fetch vacancies: %v", err)
	}
	for _, item := range items {
		// if vacancy it is wrong
		if isWrongVacancy(item) {
			continue
		}
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

		// wait time for chat id
		<-h.chatsTimers.GetPut(
			s.ChatID,
			timer.NewRefreshTimer(timeout, true),
		).Wait()
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
