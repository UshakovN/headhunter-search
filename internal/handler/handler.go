package handler

import (
  "context"
  "fmt"
  "main/internal/fetcher"
  "main/internal/model"
  "main/internal/storage"
  "main/pkg/telegram"
  "main/pkg/utils"
  "time"

  log "github.com/sirupsen/logrus"
  "golang.org/x/sync/errgroup"
)

type Handler struct {
  ctx     context.Context
  bot     telegram.Bot
  fetcher fetcher.Fetcher
  storage storage.Storage

  pendingChats      *utils.Map[int64]
  sentSubscriptions *utils.Map[string]
  sentVacancies     *utils.Map[string]
}

func NewHandler(ctx context.Context, bot telegram.Bot, fetcher fetcher.Fetcher, storage storage.Storage) (*Handler, error) {
  h := &Handler{
    ctx:          ctx,
    bot:          bot,
    fetcher:      fetcher,
    storage:      storage,
    pendingChats: utils.NewMap[int64](),
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
  m := utils.NewMap[string]()

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
  m := utils.NewMap[string]()

  for _, s := range states {
    m.Put(s.SubscriptionID)
  }
  h.sentSubscriptions = m

  return nil
}

func (h *Handler) HandleSubscriptions(ctx context.Context) {
  const (
    okWait   = 1 * time.Minute
    errWait  = 10 * time.Second
    egrLimit = 30
  )
  g, _ := errgroup.WithContext(ctx)
  g.SetLimit(egrLimit)

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
    startPeriod = 14
    nextPeriod  = 1
    messageWait = time.Second
  )
  var (
    period int
  )
  if !h.sentSubscriptions.Exist(s.SubscriptionID) {
    period = startPeriod
  } else {
    period = nextPeriod
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
    m := formVacancyMessage(item)

    if err = h.bot.SendMessage(&telegram.SendMessage{
      ChatID:    s.ChatID,
      Text:      m,
      ParseMode: telegram.ParseModeMarkdownV2,
    }); err != nil {
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
    time.Sleep(messageWait)
  }
  return nil
}

func formVacancyMessage(item *fetcher.VacancyResponseItem) string {
  const (
    t = `
Вакансия: %s
Город: %s
Зарплата: %d - %d (%s)
Статус: %s
Компания: %s (%s)
Обязанности: %s
Требуемые навыки: %s
Требуемый опыт работы: %s
Тип занятости: %s
Опубликована: %s
Ссылка на вакансию: %s
`
  )
  return fmt.Sprintf(t,
    item.Name,
    item.Area.Name,
    item.Salary.From,
    item.Salary.To,
    item.Salary.Currency,
    item.Type.Name,
    item.Employer.Name,
    item.Employer.Url,
    item.Snippet.Responsibility,
    item.Snippet.Requirement,
    item.Experience.Name,
    item.Employment.Name,
    item.PublishedAt,
    item.Url,
  )
}

func (h *Handler) HandleMessages(_ context.Context) {
  var (
    err error
  )
  h.bot.HandleMessages(func(m *telegram.Message) error {
    if m.IsCommand() {
      err = h.handleTextCommand(m)
    }
    if m.IsText() {
      err = h.handleTextMessage(m)
    }
    return err
  })
}

func (h *Handler) handleTextMessage(m *telegram.Message) error {
  const (
    sub = `Вы успешно подписались на рассылку вакансий по ключевой фразе: 
%s.`
    warn = `Сообщение не распознано. Воспользуйтесь справкой /man.`

    wrong = `Сообщение имеет неверный формат. 
Запрос для поиска вакансий: Город // Опыт работы // Ключевые фразы. 
(Город: Москва -> 1 / Санкт-Петербург -> 2). 
(Опыт работы: 1-3 года -> 1 / 3-6 лет -> 2 / без опыта -> 3).`
  )
  if !h.pendingChats.Exist(m.ChatID) {
    return h.bot.SendMessage(&telegram.SendMessage{
      ChatID:    m.ChatID,
      Text:      warn,
      ParseMode: telegram.ParseModeMarkdownV2,
    })
  }
  parsed, err := parseTextMessage(m.Text)
  if err != nil {
    return h.bot.SendMessage(&telegram.SendMessage{
      ChatID:    m.ChatID,
      Text:      wrong,
      ParseMode: telegram.ParseModeMarkdownV2,
    })
  }
  if err := h.storage.PutUserSubscription(h.ctx, &model.Subscription{
    SubscriptionID: utils.NewUUID(),
    UserID:         m.UserID,
    ChatID:         m.ChatID,
    Keywords:       parsed.Keywords,
    Area:           parsed.AreaCode,
    Experience:     parsed.ExperienceCode,
    CreatedAt:      utils.NowTimeUTC(),
  }); err != nil {
    return fmt.Errorf("cannot put subscription in storage: %v", err)
  }
  defer h.pendingChats.Delete(m.ChatID)

  return h.bot.SendMessage(&telegram.SendMessage{
    ChatID:    m.ChatID,
    Text:      fmt.Sprintf(sub, parsed.Keywords),
    ParseMode: telegram.ParseModeMarkdownV2,
  })
}

func (h *Handler) handleTextCommand(m *telegram.Message) error {
  const (
    undefined = `Команда не распознана. 
Воспользуйтесь справкой /man.`

    start = `Данный бот способен собирать актуальные вакансии по ключевым фразам. 
Для подписки используйте команду /sub. 
Чтобы отписаться используйте /unsub. 
Справка /man.`

    sub = `Введите запрос для поиска вакансий: 
Город // Опыт работы // Ключевые фразы. 
(Город: Москва -> 1 / Санкт-Петербург -> 2). 
(Опыт работы: 1-3 года -> 1 / 3-6 лет -> 2 / без опыта -> 3).`

    unsub = `Введите идентификатор подписки или ключевую фразу.`

    man = `Для подписки используйте команду /sub. 
Чтобы отписаться используйте /unsub. 
Сбросить бота /reset.`

    reset = `Состояние бота сброшено. 
Чтобы начать воспользуйтесь /start.`
  )
  var (
    text string
  )
  switch m.Command {
  case "start":
    text = start
  case "sub":
    text = sub
    h.pendingChats.Put(m.ChatID)
  case "unsub":
    text = unsub
  case "man":
    text = man
  case "reset":
    text = reset
    h.pendingChats.Delete(m.ChatID)
  default:
    text = undefined
  }
  return h.bot.SendMessage(&telegram.SendMessage{
    ChatID:    m.ChatID,
    Text:      text,
    ParseMode: telegram.ParseModeMarkdownV2,
  })
}
