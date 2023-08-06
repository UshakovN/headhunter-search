package storage

import (
  "context"
  "main/internal/model"
)

type Storage interface {
  UsersSubscriptionsSets(ctx context.Context, sf func(subset *model.SubscriptionSet)) error
  UsersSubscriptions(ctx context.Context, sf func(sub *model.Subscription)) error
  UserSubscriptions(ctx context.Context, userID int64) ([]*model.Subscription, error)
  PutUserSubscription(ctx context.Context, sub *model.Subscription) error
  SentSubscriptions(ctx context.Context) ([]*model.SentSubscription, error)
  PutSentSubscription(ctx context.Context, st *model.SentSubscription) error
  SentVacancies(ctx context.Context) ([]*model.SentVacancy, error)
  PutSentVacancy(ctx context.Context, sv *model.SentVacancy) error
}
