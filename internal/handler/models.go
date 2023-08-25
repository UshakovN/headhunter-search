package handler

import (
	"fmt"
	"main/internal/fetcher"
	"main/internal/model"
	"main/pkg/telegram"
)

type subVacancy struct {
	Area       string
	Experience string
	Keywords   string
}

func (f *subVacancy) IsFilled() bool {
	return f.Area != "" && f.Experience != "" && f.Keywords != ""
}

func newStartMessage(chatID int64) *telegram.SendMessage {
	text := `Данный бот способен собирать актуальные вакансии 👀`

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
			Text:    "Контакты 🍪",
			Command: "/contacts",
		},
		telegram.InlineKeyboardButton{
			Text:    "Справка 💭",
			Command: "/man",
		})

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

func newUnsubMessage(chatID int64, subs []*model.Subscription) *telegram.SendMessage {
	text := `Отписаться от рассылки вакансий 📤
Выберите требуемую подписку 👀`

	buttons := make([]telegram.InlineKeyboardButton, 0, len(subs))

	for index, sub := range subs {
		buttons = append(buttons, telegram.InlineKeyboardButton{
			Text:    fmt.Sprintf("%d 🌠️ %s", index+1, sub.Keywords),
			Command: fmt.Sprintf("/unsub?id=%s", sub.SubscriptionID),
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

func newVacancyMessage(chatID int64, item *fetcher.VacancyResponseItem) *telegram.SendMessage {
	const (
		t = `🌠🌠🌠🌠
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
👔👔👔👔`
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
