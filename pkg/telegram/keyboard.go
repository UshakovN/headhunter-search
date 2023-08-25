package telegram

import tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type InlineKeyboard struct {
	markup tg.InlineKeyboardMarkup
}

type InlineKeyboardButton struct {
	Text    string
	Command string
}

type InlineKeyboardButtonsMarkup int

const (
	InRowButtonsMarkup InlineKeyboardButtonsMarkup = 0
	InColButtonsMarkup InlineKeyboardButtonsMarkup = 1
)

func NewInlineKeyboard(markup InlineKeyboardButtonsMarkup, buttons ...InlineKeyboardButton) *InlineKeyboard {
	var rows [][]tg.InlineKeyboardButton

	switch markup {

	case InRowButtonsMarkup:
		row := make([]tg.InlineKeyboardButton, 0, len(buttons))
		for _, button := range buttons {
			row = append(row, tg.NewInlineKeyboardButtonData(button.Text, button.Command))
		}
		rows = append(rows, row)

	case InColButtonsMarkup:
		for _, button := range buttons {
			rows = append(rows, []tg.InlineKeyboardButton{
				tg.NewInlineKeyboardButtonData(button.Text, button.Command),
			})
		}
	}

	return &InlineKeyboard{
		markup: tg.NewInlineKeyboardMarkup(rows...),
	}
}
