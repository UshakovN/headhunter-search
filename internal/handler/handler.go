package handler

import (
	"context"
	"fmt"
	"main/internal/dialog"
	"main/internal/fetcher"
	"main/internal/model"
	"main/internal/storage"
	"main/internal/task"
	"main/pkg/cache"
	"main/pkg/http"
	"main/pkg/telegram"
	"main/pkg/timer"
	"main/pkg/utils"
	"strings"
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
	vacancyForms      cache.MemCache[int64, *VacancyForm]
	pendingChats      cache.KeyCache[int64]
	sentSubscriptions cache.KeyCache[string]
	sentVacancies     cache.KeyCache[string]

	usersDialogs *dialog.ChatsTrees // TODO: remove this
}

func NewHandler(ctx context.Context, bot telegram.Bot, fetcher fetcher.Fetcher, storage storage.Storage) (*Handler, error) {
	const (
		workers = 100
	)
	h := &Handler{
		ctx:               ctx,
		bot:               bot,
		fetcher:           fetcher,
		storage:           storage,
		taskQueue:         task.NewQueue(workers),
		chatsTimer:        timer.NewTimer[int64](),
		vacancyForms:      cache.NewMemCache[int64, *VacancyForm](),
		pendingChats:      cache.NewKeyCache[int64](),
		sentSubscriptions: cache.NewKeyCache[string](),
		sentVacancies:     cache.NewKeyCache[string](),

		usersDialogs: dialog.NewChatsTrees(), // TODO: remove this
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

func (h *Handler) HandleTasks(ctx context.Context) {
	h.taskQueue.ContinuouslyHandle(ctx)
}

func (h *Handler) HandleSubscriptions(ctx context.Context) {
	const (
		okWait     = 1 * time.Minute
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
		firstSentPeriod = 14
		nextSentPeriod  = 1
		messageTimeout  = 10 * time.Second
	)
	var period int

	if !h.sentSubscriptions.Exist(s.SubscriptionID) {
		period = firstSentPeriod
	} else {
		period = nextSentPeriod
	}
	req := fetcher.NewVacanciesRequest(s.Keywords, s.Area, s.Experience, period)

	resp, err := h.fetcher.Fetch(ctx, req)
	if err != nil {
		return fmt.Errorf("cannot fetch vacancies for request: %v", err)
	}
	for _, item := range resp.Items {
		if h.sentVacancies.Exist(item.Id) {
			continue
		}
		msg := newVacancyMessage(s.ChatID, item)

		if err = h.bot.SendMessage(msg); err != nil {
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
		h.chatsTimer.Wait(s.ChatID)
		h.chatsTimer.Set(s.ChatID, messageTimeout)
	}
	return nil
}

func (h *Handler) HandleTextCommandTODO(_ context.Context, m *telegram.Message) error { // TODO: remove this
	msg := h.usersDialogs.Dialog(m.ChatID).Process(m.ChatID, dialog.Link(m.Command))
	return h.bot.SendMessage(msg)
}

func (h *Handler) HandleMessages(ctx context.Context) {
	var (
		err error
	)
	h.bot.HandleMessages(func(m *telegram.Message) error {
		if m.IsCommand() {
			err = h.HandleTextCommandTODO(ctx, m) // TODO: remove this

			//err = h.handleTextCommand(ctx, m) // TODO: uncomment this
		}
		if m.IsText() {
			err = h.handleTextMessage(ctx, m)
		}
		return err
	})
}

func (h *Handler) handleTextMessage(_ context.Context, m *telegram.Message) error {
	var (
		msg *telegram.SendMessage
	)
	if !h.pendingChats.Exist(m.ChatID) {
		msg = newUndefinedMessage(m.ChatID)
	} else {
		vacancyForm := h.vacancyForms.GetPut(m.ChatID, &VacancyForm{})
		vacancyForm.Keywords = m.Text

		if vacancyForm.IsFilled() {
			msg = newFormConfirmCancelMessage(m.ChatID)
		} else {
			msg = newFormFillFieldsMessage(m.ChatID)
		}
	}
	return h.bot.SendMessage(msg)
}

func (h *Handler) handleTextCommand(ctx context.Context, m *telegram.Message) error {
	var msg *telegram.SendMessage

	// handle /start command
	if cmd := m.Command; cmd == "start" {
		msg = newStartMessage(m.ChatID)
	}
	// handle /dialog/ commands
	if cmd, ok := utils.TrimPrefixIfExist(m.Command, "dialog/"); ok {

		// handle commands
		switch cmd {
		case "sub":
			msg = newSubMessage(m.ChatID)
			h.pendingChats.Put(m.ChatID)
		case "unsub":
			subs, err := h.storage.UserSubscriptions(ctx, m.UserID)
			if err != nil {
				return fmt.Errorf("cannot got subscriptions from storage for user with id: %d: %v", m.UserID, err)
			}
			msg = newUnsubMessage(m.ChatID, subs)
		case "man":
			msg = newManMessage(m.ChatID)
		case "reset":
			msg = newResetMessage(m.ChatID)
			h.pendingChats.Delete(m.ChatID)
		case "contacts":
			msg = newContactsMessage(m.ChatID)
		}
	}
	// handle /action/ commands
	if cmd, ok := utils.TrimPrefixIfExist(m.Command, "action/"); ok {
		// handle /action/sub/form/ commands
		if cmd, ok := utils.TrimPrefixIfExist(cmd, "sub/form/"); ok {

			// handle commands without query
			switch cmd {
			case "area":
				msg = newFormAreaMessage(m.ChatID)
			case "experience":
				msg = newFormExperienceMessage(m.ChatID)
			case "keywords":
				msg = newFormKeywordsMessage(m.ChatID)
			case "confirm":
				msg = newFormConfirmMessage(m.ChatID)
				vacancyForm := h.vacancyForms.Get(m.ChatID)

				h.taskQueue.Push(func() error {
					if err := h.storage.PutUserSubscription(h.ctx, &model.Subscription{
						SubscriptionID: utils.NewUUID(),
						UserID:         m.UserID,
						ChatID:         m.ChatID,
						Keywords:       vacancyForm.Keywords,
						Area:           vacancyForm.Area,
						Experience:     vacancyForm.Experience,
						CreatedAt:      utils.NowTimeUTC(),
					}); err != nil {
						return fmt.Errorf("cannot put subscription in storage: %v", err)
					}
					return nil
				})
				h.pendingChats.Delete(m.ChatID)
			case "cancel":
				msg = newFormCancelMessage(m.ChatID)
				h.pendingChats.Delete(m.ChatID)

			default:
				// handle commands with query
				uf := h.vacancyForms.GetPut(m.ChatID, &VacancyForm{})

				switch {
				case strings.HasPrefix(cmd, "area"):
					// handle /action/sub/form/area query
					query := strings.TrimPrefix(cmd, "area?")
					uf.Area = http.MustParseQuery(query).Get("id")

				case strings.HasPrefix(cmd, "experience"):
					// handle /action/sub/form/experience query
					query := strings.TrimPrefix(cmd, "experience?")
					uf.Experience = http.MustParseQuery(query).Get("id")
				}
				// check /action/sub/form filling
				if uf.IsFilled() {
					// if /action/sub/form completely filled sent confirm message
					msg = newFormConfirmCancelMessage(m.ChatID)
				} else {
					// if /action/sub/form not filled sent fill fields message
					msg = newFormFillFieldsMessage(m.ChatID)
				}
			}
		}
		// handle /action/unsub/ commands
		if cmd, ok := utils.TrimPrefixIfExist(cmd, "unsub/"); ok {

			switch {
			case strings.HasPrefix(cmd, "sub"):
				// handle /action/unsub/subscription query
				query := strings.TrimPrefix(cmd, "sub?")
				subID := http.MustParseQuery(query).Get("id")

				if err := h.storage.DeleteUserSubscription(ctx, subID); err != nil {
					return fmt.Errorf("cannot delete subscription with id from storage: %s: %v", subID, err)
				}
			}
		}
	}
	msg = newCoalesceMessage(msg, newUndefinedMessage(m.ChatID))

	return h.bot.SendMessage(msg)
}
