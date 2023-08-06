package model

import "time"

type Subscription struct {
  SubscriptionID string
  UserID         int64
  ChatID         int64
  Area           string
  Keywords       string
  Experience     string
  CreatedAt      time.Time
}

type SubscriptionSet struct {
  SubscriptionIDs []string
  UserIDs         []int64
  ChatIDs         []int64
  Area            string
  Keywords        string
  Experience      string
}

type SentSubscription struct {
  SentID         string
  SubscriptionID string
  CreatedAt      time.Time
}

type SentVacancy struct {
  SentID    string
  ChatID    int64
  VacancyID string
  CreatedAt time.Time
}
