package model

import "time"

type ChatSubscription struct {
	SubscriptionID int64
	ChatID         int64
	UserID         int64
	Area           string
	Keywords       string
	Experience     string
	CreatedAt      time.Time
}

type ChatSubscriptionSet struct {
	SubscriptionIDs []int64
	UserIDs         []int64
	ChatIDs         []int64
	Area            string
	Keywords        string
	Experience      string
}

type ChatSentVacancy struct {
	SentID         int64
	SubscriptionID int64
	ChatID         int64
	VacancyID      string
	CreatedAt      time.Time
}

type ChatTree struct {
	ChatTreeID     int64
	ChatID         int64
	SerializedTree []byte
	CreatedAt      time.Time
}
