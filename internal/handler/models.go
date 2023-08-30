package handler

import (
	"fmt"
	"main/internal/fetcher"
	"main/internal/model"
	"main/pkg/str"
	"main/pkg/telegram"
	"main/pkg/utils"
	"strings"
)

type vacancy struct {
	area       string
	experience string
	keywords   string
}

func (f *vacancy) IsFilled() bool {
	return f.area != "" && f.experience != "" && f.keywords != ""
}

func newStartMessage(chatID int64, withStop bool) *telegram.SendMessage {
	text := `Данный бот способен собирать актуальные вакансии 👀`

	buttons := []telegram.InlineKeyboardButton{
		{
			Text:    "Подписаться 📩",
			Command: "/sub",
		},
		{
			Text:    "Отписаться 📤",
			Command: "/unsub",
		},
		{
			Text:    "Контакты 🍪",
			Command: "/contacts",
		},
		{
			Text:    "Справка 💭",
			Command: "/man",
		},
	}
	if withStop {
		buttons = append(buttons, telegram.InlineKeyboardButton{
			Text:    "Продолжить рассылку ✉️",
			Command: "stop",
		})
	}
	keyboard := telegram.NewInlineKeyboard(
		telegram.InColButtonsMarkup,
		buttons...,
	)
	return &telegram.SendMessage{
		ChatID:   chatID,
		Text:     text,
		Keyboard: keyboard,
	}
}

func newContactsMessage(chatID int64) *telegram.SendMessage {
	text := `Разработчик 🍪 @ushakovn 🍪`

	keyboard := telegram.NewInlineKeyboard(telegram.InColButtonsMarkup,
		telegram.InlineKeyboardButton{
			Text:    "Назад 🔍",
			Command: "/back",
		})

	return &telegram.SendMessage{
		ChatID:   chatID,
		Text:     text,
		Keyboard: keyboard,
	}
}

func newSubMessage(chatID int64) *telegram.SendMessage {
	text := `Подписаться на рассылку вакансий 📩
Укажите следующие параметры для поиска вакансий `

	keyboard := telegram.NewInlineKeyboard(telegram.InColButtonsMarkup,
		telegram.InlineKeyboardButton{
			Text:    "Местоположение 🌎",
			Command: "/area",
		},
		telegram.InlineKeyboardButton{
			Text:    "Опыт работы 👔",
			Command: "/experience",
		},
		telegram.InlineKeyboardButton{
			Text:    "Название вакансии 🌠",
			Command: "/keywords",
		},
		telegram.InlineKeyboardButton{
			Text:    "Назад 🔍",
			Command: "/back",
		})

	return &telegram.SendMessage{
		ChatID:   chatID,
		Text:     text,
		Keyboard: keyboard,
	}
}

func newUnsubCompleteMessage(chatID int64) *telegram.SendMessage {
	keyboard := telegram.NewInlineKeyboard(telegram.InColButtonsMarkup,
		telegram.InlineKeyboardButton{
			Text:    "Назад 🔍",
			Command: "/back",
		})

	return &telegram.SendMessage{
		ChatID:   chatID,
		Keyboard: keyboard,
		Text:     fmt.Sprintf(`Вы успешно отписались от выбранной рассылки вакансий ❗️`),
	}
}

func newUnsubMessage(chatID int64, subs []*model.ChatSubscription) *telegram.SendMessage {
	text := `Отписаться от рассылки вакансий 📤
Выберите требуемую подписку 👀`

	buttons := make([]telegram.InlineKeyboardButton, 0, len(subs))

	for index, sub := range subs {
		buttons = append(buttons, telegram.InlineKeyboardButton{
			Text:    fmt.Sprintf("%d 🌠️ %s", index+1, sub.Keywords),
			Command: fmt.Sprintf("/unsub?id=%d", sub.SubscriptionID),
		})
	}
	buttons = append(buttons, telegram.InlineKeyboardButton{
		Text:    "Назад 🔍",
		Command: "/back",
	})

	keyboard := telegram.NewInlineKeyboard(
		telegram.InColButtonsMarkup,
		buttons...,
	)
	return &telegram.SendMessage{
		ChatID:   chatID,
		Text:     text,
		Keyboard: keyboard,
	}
}

func newManMessage(chatID int64) *telegram.SendMessage {
	text := `Бот имеет следующие команды 📑`

	keyboard := telegram.NewInlineKeyboard(telegram.InColButtonsMarkup,
		telegram.InlineKeyboardButton{
			Text:    "Подписаться 📩",
			Command: "/sub",
		},
		telegram.InlineKeyboardButton{
			Text:    "Отписаться 📤",
			Command: "/unsub",
		},
		telegram.InlineKeyboardButton{
			Text:    "Справка 💭",
			Command: "/man",
		},
		telegram.InlineKeyboardButton{
			Text:    "Назад 🔍",
			Command: "/back",
		})

	return &telegram.SendMessage{
		ChatID:   chatID,
		Text:     text,
		Keyboard: keyboard,
	}
}

func newAreaMessage(chatID int64) *telegram.SendMessage {
	text := `Выберите местоположение из доступных 🌎`

	keyboard := telegram.NewInlineKeyboard(telegram.InColButtonsMarkup,
		telegram.InlineKeyboardButton{
			Text:    "Москва",
			Command: "/area?id=1",
		},
		telegram.InlineKeyboardButton{
			Text:    "Санкт-Петербург",
			Command: "/area?id=2",
		},
		telegram.InlineKeyboardButton{
			Text:    "Назад 🔍",
			Command: "/back",
		})

	return &telegram.SendMessage{
		ChatID:   chatID,
		Text:     text,
		Keyboard: keyboard,
	}
}

func newExperienceMessage(chatID int64) *telegram.SendMessage {
	text := `Выберите опыт работы из доступных 👔`

	keyboard := telegram.NewInlineKeyboard(telegram.InColButtonsMarkup,
		telegram.InlineKeyboardButton{
			Text:    "От 1 до 3 лет",
			Command: "/experience?id=between1And3",
		},
		telegram.InlineKeyboardButton{
			Text:    "От 3 до 6 лет",
			Command: "/experience?id=between3And6",
		},
		telegram.InlineKeyboardButton{
			Text:    "Без коммерческого опыта",
			Command: "/experience?id=noExperience",
		},
		telegram.InlineKeyboardButton{
			Text:    "Назад 🔍",
			Command: "/back",
		})

	return &telegram.SendMessage{
		ChatID:   chatID,
		Text:     text,
		Keyboard: keyboard,
	}
}

func newKeywordsMessage(chatID int64) *telegram.SendMessage {
	return &telegram.SendMessage{
		ChatID: chatID,
		Text:   `Укажите название вакансии 🌠`,
	}
}

func newFillFieldsMessage(chatID int64) *telegram.SendMessage {
	text := `Укажите оставшиеся поля ✅`

	keyboard := telegram.NewInlineKeyboard(telegram.InColButtonsMarkup,
		telegram.InlineKeyboardButton{
			Text:    "Назад 🔍",
			Command: "/back",
		})

	return &telegram.SendMessage{
		ChatID:   chatID,
		Text:     text,
		Keyboard: keyboard,
	}
}

func newConfirmCancelMessage(chatID int64) *telegram.SendMessage {
	text := `Подтвердите подписку на вакансию или отмените выбор ✉️`

	keyboard := telegram.NewInlineKeyboard(telegram.InColButtonsMarkup,
		telegram.InlineKeyboardButton{
			Text:    "Подтвердить ✅",
			Command: "/confirm",
		},
		telegram.InlineKeyboardButton{
			Text:    "Отмена ❗",
			Command: "/cancel",
		},
		telegram.InlineKeyboardButton{
			Text:    "Назад 🔍",
			Command: "/back",
		})

	return &telegram.SendMessage{
		ChatID:   chatID,
		Text:     text,
		Keyboard: keyboard,
	}
}

func newCancelMessage(chatID int64) *telegram.SendMessage {
	return &telegram.SendMessage{
		ChatID: chatID,
		Text:   `Вы отменили создание подписки на рассылку вакансий ❗`,
	}
}

func newConfirmMessage(chatID int64) *telegram.SendMessage {
	return &telegram.SendMessage{
		ChatID: chatID,
		Text: `Вы подтвердили создание подписки на вакансии ✅
Список актуальных вакансий сейчас будет подобран 🍪`,
	}
}

func newVacancyMessage(chatID int64, keywords string, item *fetcher.VacancyResponseItem) *telegram.SendMessage {
	s := strings.Builder{}

	url := fmt.Sprintf("<a href=\"%s\">Новая вакансия</a>", item.AlternateUrl)
	s.WriteString(fmt.Sprintf("🌠📨🌠📨🌠 %s\n\n", url))

	s.WriteString(fmt.Sprintf("<b>🍪 Подписка</b>\n%s\n\n", str.Sanitize(keywords)))

	s.WriteString(fmt.Sprintf("<b>👔 Название</b>\n%s\n\n", str.Sanitize(item.Name)))

	if area := item.Area; area != nil && area.Name != "" {
		s.WriteString(fmt.Sprintf("<b>🌎 Город</b>\n%s\n\n", str.Sanitize(area.Name)))
	}

	if salary := item.Salary; salary != nil && salary.Currency != "" {
		curr := str.Sanitize(salary.Currency)
		curr = strings.ToUpper(curr)

		if fork := salary.From > 0 && salary.To > 0; fork {
			s.WriteString(fmt.Sprintf("<b>💶 Зарплата</b>\nОт %d до %d (%s)", salary.From, salary.To, curr))
		} else if from := salary.From; from > 0 {
			s.WriteString(fmt.Sprintf("<b>💶 Зарплата</b>\nОт %d (%s)", from, curr))
		} else if to := salary.To; to > 0 {
			s.WriteString(fmt.Sprintf("<b>💶 Зарплата</b>\nДо %d (%s)", to, curr))
		}

		if salary.Gross {
			s.WriteString(" <i>до вычета НДФЛ</i>")
		}
		s.WriteString("\n\n")
	}

	if employer := item.Employer; employer != nil && employer.Name != "" { // TODO: add employer url
		s.WriteString(fmt.Sprintf("<b>⭐ Компания</b>\n%s\n\n", str.Sanitize(employer.Name)))
	}

	if snippet := item.Snippet; snippet != nil {
		if req := snippet.Requirement; req != "" {
			s.WriteString(fmt.Sprintf("<b>👨‍💼 Требуемые навыки </b>\n%s\n\n", str.Sanitize(req)))
		}
		if resp := snippet.Responsibility; resp != "" {
			s.WriteString(fmt.Sprintf("<b>💡 Обязанности</b>\n%s\n\n", str.Sanitize(resp)))
		}
	}

	if exp := item.Experience; exp != nil {
		if exp := exp.Name; exp != "" {
			s.WriteString(fmt.Sprintf("<b>⏳ Требуемый опыт работы</b>\n%s\n\n", str.Sanitize(exp)))
		}
	}

	if hhUrl := item.AlternateUrl; hhUrl != "" {
		hhUrl := fmt.Sprintf("<a href=\"%s\">Ссылка</a>", hhUrl)
		s.WriteString(fmt.Sprintf("<b>📑 Ссылка на вакансию</b>\n%s\n\n", hhUrl))
	}

	if pub := item.PublishedAt; pub != "" {
		const (
			hhTimeLayout  = "2006-01-02T15:04:05-0700"
			msgTimeLayout = "02-01-2006 15:04"
		)
		if pub, err := utils.TimeStrCast(pub, hhTimeLayout, msgTimeLayout); err == nil {
			s.WriteString(fmt.Sprintf("<b>🕒 Вакансия опубликована</b>\n%s\n\n", pub))
		}
	}
	s.WriteString("🌠📨🌠📨🌠\n\n")

	if tags := str.BuildSentenceTags(keywords); len(tags) >= 0 {
		for tagIndex, tag := range tags {
			s.WriteString(fmt.Sprintf("<b>%s</b>", tag))

			if tagIndex < len(tags)-1 {
				s.WriteString(" ")
			}
		}
		s.WriteString("\n")
	}
	text := s.String()

	keyboard := telegram.NewInlineKeyboard(telegram.InColButtonsMarkup,
		telegram.InlineKeyboardButton{
			Text:    "Перейти в меню бота 💭",
			Command: "/start",
		})

	return &telegram.SendMessage{
		ChatID:   chatID,
		Text:     text,
		Keyboard: keyboard,
	}
}

func isWrongVacancy(item *fetcher.VacancyResponseItem) bool {
	switch {
	case
		item.Archived,
		item.AlternateUrl == "",
		item.Name == "",
		item.Snippet == nil,
		item.Employer == nil:
		return true
	default:
		return false
	}
}
