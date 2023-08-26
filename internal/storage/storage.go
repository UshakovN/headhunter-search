package storage

import (
	"context"
	"main/internal/model"
)

type Storage interface {
	ChatSubscriptionsSets(ctx context.Context, callback func(subSet *model.ChatSubscriptionSet)) error
	ChatsSubscriptions(ctx context.Context, callback func(sub *model.ChatSubscription)) error
	ChatSubscriptions(ctx context.Context, chatID int64) ([]*model.ChatSubscription, error)
	PutChatSubscription(ctx context.Context, sub *model.ChatSubscription) error
	SentVacancies(ctx context.Context) ([]*model.ChatSentVacancy, error)
	PutSentVacancy(ctx context.Context, sentVacancy *model.ChatSentVacancy) error
	DeleteChatSubscription(ctx context.Context, subID int64) error
}
