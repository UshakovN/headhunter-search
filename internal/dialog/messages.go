package dialog

import (
	"fmt"
	"main/internal/fetcher"
	"main/internal/model"
	"main/pkg/telegram"
)

func newStartMessage(chatID int64) *telegram.SendMessage {
	text := `Данный бот способен собирать актуальные вакансии 👀`

	keyboard := telegram.NewInlineKeyboard(telegram.InColButtonsMarkup,
		telegram.InlineKeyboardButton{
			Text:    "Подписаться 📩",
			Command: string(LinkSub),
		},
		telegram.InlineKeyboardButton{
			Text:    "Отписаться 📤",
			Command: string(LinkUnsub),
		},
		telegram.InlineKeyboardButton{
			Text:    "Контакты 🍪",
			Command: string(LinkContacts),
		},
		telegram.InlineKeyboardButton{
			Text:    "Справка 💭",
			Command: string(LinkMan),
		},
		telegram.InlineKeyboardButton{
			Text:    "Назад 📑",
			Command: string(LinkBack),
		})

	return &telegram.SendMessage{
		ChatID:   chatID,
		Text:     text,
		Keyboard: keyboard,
	}
}

func newContactsMessage(chatID int64) *telegram.SendMessage {
	return &telegram.SendMessage{
		ChatID: chatID,
		Text:   `Разработчик 🍪 @ushakovn 🍪`,
	}
}

func newSubMessage(chatID int64) *telegram.SendMessage {
	text := `Подписаться на рассылку вакансий 📩
Укажите следующие параметры для поиска вакансий `

	keyboard := telegram.NewInlineKeyboard(telegram.InColButtonsMarkup,
		telegram.InlineKeyboardButton{
			Text:    "Местоположение 🌎",
			Command: string(LinkSubArea),
		},
		telegram.InlineKeyboardButton{
			Text:    "Опыт работы 👔",
			Command: string(LinkSubExperience),
		},
		telegram.InlineKeyboardButton{
			Text:    "Название вакансии 🌠",
			Command: string(LinkSubKeywords),
		})

	return &telegram.SendMessage{
		ChatID:   chatID,
		Text:     text,
		Keyboard: keyboard,
	}
}

func newUnsubMessage(chatID int64, subs []*model.Subscription) *telegram.SendMessage {
	text := `Отписаться от рассылки вакансий 📤
Выберите требуемую подписку 👀`

	buttons := make([]telegram.InlineKeyboardButton, 0, len(subs))

	for index, sub := range subs {
		buttons = append(buttons, telegram.InlineKeyboardButton{
			Text:    fmt.Sprintf("%d 📌 %s", index+1, sub.Keywords),
			Command: fmt.Sprintf("action/unsub/sub?id=%s", sub.SubscriptionID),
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

func newManMessage(chatID int64) *telegram.SendMessage {
	text := `Бот имеет следующие команды 📋`

	keyboard := telegram.NewInlineKeyboard(telegram.InColButtonsMarkup,
		telegram.InlineKeyboardButton{
			Text:    "Подписаться 📩",
			Command: "/dialog/sub",
		},
		telegram.InlineKeyboardButton{
			Text:    "Отписаться 📤",
			Command: "/dialog/unsub",
		},
		telegram.InlineKeyboardButton{
			Text:    "Справка 💭",
			Command: "/dialog/man",
		})

	return &telegram.SendMessage{
		ChatID:   chatID,
		Text:     text,
		Keyboard: keyboard,
	}
}

func newCoalesceMessage(m ...*telegram.SendMessage) *telegram.SendMessage {
	for _, m := range m {
		if m != nil {
			return m
		}
	}
	return nil
}

func newUndefinedMessage(chatID int64) *telegram.SendMessage {
	text := `Команда не распознана 🔍
Воспользуйтесь справкой`

	keyboard := telegram.NewInlineKeyboard(telegram.InColButtonsMarkup,
		telegram.InlineKeyboardButton{
			Text:    "Справка 💭",
			Command: "/dialog/man",
		})

	return &telegram.SendMessage{
		ChatID:   chatID,
		Text:     text,
		Keyboard: keyboard,
	}
}

func newResetMessage(chatID int64) *telegram.SendMessage {
	return &telegram.SendMessage{
		ChatID: chatID,
		Text:   `Последние введенные команды были отменены ❗`,
	}
}

func newSubAreaMessage(chatID int64) *telegram.SendMessage {
	text := `Выберите местоположение из доступных 🌎`

	keyboard := telegram.NewInlineKeyboard(telegram.InColButtonsMarkup,
		telegram.InlineKeyboardButton{
			Text:    "Москва",
			Command: "action/sub/form/area?id=1",
		},
		telegram.InlineKeyboardButton{
			Text:    "Санкт-Петербург",
			Command: "action/sub/form/area?id=2",
		})

	return &telegram.SendMessage{
		ChatID:   chatID,
		Text:     text,
		Keyboard: keyboard,
	}
}

func newSubExperienceMessage(chatID int64) *telegram.SendMessage {
	text := `Выберите опыт работы из доступных 👔`

	keyboard := telegram.NewInlineKeyboard(telegram.InColButtonsMarkup,
		telegram.InlineKeyboardButton{
			Text:    "От 1 до 3 лет",
			Command: "action/sub/form/experience?id=between1And3",
		},
		telegram.InlineKeyboardButton{
			Text:    "От 3 до 6 лет",
			Command: "action/sub/form/experience?id=between3And6",
		},
		telegram.InlineKeyboardButton{
			Text:    "Без коммерческого опыта",
			Command: "action/sub/form/form/experience?id=noExperience",
		})

	return &telegram.SendMessage{
		ChatID:   chatID,
		Text:     text,
		Keyboard: keyboard,
	}
}

func newSubKeywordsMessage(chatID int64) *telegram.SendMessage {
	return &telegram.SendMessage{
		ChatID: chatID,
		Text:   `Укажите название вакансии 🌠`,
	}
}

func newFormFillFieldsMessage(chatID int64) *telegram.SendMessage {
	return &telegram.SendMessage{
		ChatID: chatID,
		Text:   `Укажите оставшиеся поля ✅`,
	}
}

func newFormConfirmCancelMessage(chatID int64) *telegram.SendMessage {
	text := `Подтвердите подписку на вакансию или отмените выбор ✉️`

	keyboard := telegram.NewInlineKeyboard(telegram.InColButtonsMarkup,
		telegram.InlineKeyboardButton{
			Text:    "Подтвердить ✅",
			Command: "action/sub/form/confirm",
		},
		telegram.InlineKeyboardButton{
			Text:    "Отмена ❗",
			Command: "action/sub/form/cancel",
		})

	return &telegram.SendMessage{
		ChatID:   chatID,
		Text:     text,
		Keyboard: keyboard,
	}
}

func newFormCancelMessage(chatID int64) *telegram.SendMessage {
	return &telegram.SendMessage{
		ChatID: chatID,
		Text:   `Вы отменили создание подписки на рассылку вакансий ❗`,
	}
}

func newFormConfirmMessage(chatID int64) *telegram.SendMessage {
	return &telegram.SendMessage{
		ChatID: chatID,
		Text: `Вы подтвердили создание подписки на вакансии ✅
Список актуальных вакансий сейчас будет подобран.`,
	}
}

func newUnsubSubMessage(chatID int64, keywords string) *telegram.SendMessage {
	return &telegram.SendMessage{
		ChatID: chatID,
		Text:   fmt.Sprintf(`Вы отменили подписку на вакансии ❗%s ❗`, keywords),
	}
}

func newVacancyMessage(chatID int64, item *fetcher.VacancyResponseItem) *telegram.SendMessage {
	const (
		t = `⚡
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
⚡`
	)
	text := fmt.Sprintf(t,
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
	return &telegram.SendMessage{
		ChatID: chatID,
		Text:   text,
	}
}
