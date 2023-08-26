package storage

import (
	"context"
	"fmt"
	"main/internal/model"
	"main/pkg/postgres"
	"main/pkg/retries"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
)

const (
	retryCount = 5
	retryWait  = 1 * time.Second
)

type storage struct {
	ctx    context.Context
	client postgres.Client
}

func NewStorage(ctx context.Context, client postgres.Client) Storage {
	return &storage{
		ctx:    ctx,
		client: client,
	}
}

func (s *storage) ChatSubscriptions(ctx context.Context, chatID int64) ([]*model.ChatSubscription, error) {
	query := sanitizeQuery(
		`SELECT 
            subscription_id,
            chat_id,
            user_id,
            area,
            keywords,
            experience,
            created_at
        FROM chat_subscriptions WHERE chat_id = $1`)

	var (
		rows pgx.Rows
		err  error
	)
	if err = retries.DoWithRetries(retryCount, retryWait, func() error {
		rows, err = s.client.Query(ctx, query, postgres.SingleQuote(chatID))
		if err != nil {
			return fmt.Errorf("cannot do postgres query: %s: %v", query, err)
		}
		return nil

	}); err != nil {
		return nil, err
	}
	var (
		subs []*model.ChatSubscription
		ok   bool
	)
	for {
		sub := &model.ChatSubscription{}

		if ok, err = scanQueriedRow(rows,
			&sub.SubscriptionID,
			&sub.ChatID,
			&sub.UserID,
			&sub.Area,
			&sub.Keywords,
			&sub.Experience,
			&sub.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("cannot scan queried row: %s: %v", query, err)
		}
		if !ok {
			break
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

func (s *storage) PutChatSubscription(ctx context.Context, sub *model.ChatSubscription) error {
	query := sanitizeQuery(
		`INSERT INTO chat_subscriptions(
            chat_id,
            user_id,
            area,
            keywords,
            experience,
            created_at
        ) VALUES ($1, $2, $3, $4, $5, $6)`,
	)
	return retries.DoWithRetries(retryCount, retryWait, func() error {
		if _, err := s.client.Exec(ctx, query,
			postgres.MultiQuote(
				sub.ChatID,
				sub.UserID,
				sub.Area,
				sub.Keywords,
				sub.Experience,
				sub.CreatedAt,
			)...,
		); err != nil {
			return fmt.Errorf("cannot do postgres exec: %s: %v", query, err)
		}
		return nil
	})
}

func (s *storage) ChatsSubscriptions(ctx context.Context, callback func(sub *model.ChatSubscription)) error {
	query := sanitizeQuery(
		`SELECT
            subscription_id,
            chat_id,
            user_id,
            area,
            keywords,
            experience,
            created_at
        FROM chat_subscriptions`)

	var (
		rows pgx.Rows
		err  error
		ok   bool
	)
	if err = retries.DoWithRetries(retryCount, retryWait, func() error {
		rows, err = s.client.Query(ctx, query)
		if err != nil {
			return fmt.Errorf("cannot do postgres query: %s: %v", query, err)
		}
		return nil

	}); err != nil {
		return err
	}
	for {
		sub := &model.ChatSubscription{}

		if ok, err = scanQueriedRow(rows,
			&sub.SubscriptionID,
			&sub.ChatID,
			&sub.UserID,
			&sub.Area,
			&sub.Keywords,
			&sub.Experience,
			&sub.CreatedAt,
		); err != nil {
			return fmt.Errorf("cannot scan queried row: %v", err)
		}
		if !ok {
			break
		}
		callback(sub)
	}
	return nil
}

func (s *storage) ChatSubscriptionsSets(ctx context.Context, callback func(subSet *model.ChatSubscriptionSet)) error {
	query := sanitizeQuery(
		`SELECT DISTINCT
            subscription_ids,
            chat_ids,
            user_ids,
            area,
            keywords,
            experience
        FROM (
            SELECT
                subscription_ids,
                user_id,
                chat_id,
                area,
                experience,
                TRIM(REGEXP_REPLACE(LOWER(keywords), '\s+', ' ', 'g')) as keywords,
                ARRAY_AGG(subscription_id) OVER (PARTITION BY area, LOWER(keywords)) as subscriptions_ids,
                ARRAY_AGG(chat_id) OVER (PARTITION BY area, LOWER(keywords)) as chat_ids,
                ARRAY_AGG(user_id) OVER (PARTITION BY area, LOWER(keywords)) as user_ids
            FROM chat_subscriptions
        ) as QUERY;`)

	var (
		rows pgx.Rows
		err  error
		ok   bool
	)
	if err = retries.DoWithRetries(retryCount, retryWait, func() error {
		rows, err = s.client.Query(ctx, query)
		if err != nil {
			return fmt.Errorf("cannot do postgres query: %s: %v", query, err)
		}
		return nil

	}); err != nil {
		return err
	}
	for {
		subSet := &model.ChatSubscriptionSet{}

		if ok, err = scanQueriedRow(rows,
			&subSet.SubscriptionIDs,
			&subSet.ChatIDs,
			&subSet.UserIDs,
			&subSet.Area,
			&subSet.Keywords,
			&subSet.Experience,
		); err != nil {
			return fmt.Errorf("cannot callback queried row: %v", err)
		}
		if !ok {
			break
		}
		callback(subSet)
	}
	return nil
}

func (s *storage) SentVacancies(ctx context.Context) ([]*model.ChatSentVacancy, error) {
	query := sanitizeQuery(
		`SELECT
            sent_id, 
            sv.subscription_id,
            chat_id,
            vacancy_id, 
            sv.created_at
    FROM chat_sent_vacancies AS sv 
        INNER JOIN chat_subscriptions AS s 
    ON sv.subscription_id = s.subscription_id`)

	var (
		rows pgx.Rows
		err  error
	)
	if err = retries.DoWithRetries(retryCount, retryWait, func() error {
		rows, err = s.client.Query(ctx, query)
		if err != nil {
			return fmt.Errorf("cannot do postgres query: %s: %v", query, err)
		}
		return nil

	}); err != nil {
		return nil, err
	}
	var (
		sv []*model.ChatSentVacancy
		ok bool
	)
	for {
		s := &model.ChatSentVacancy{}

		if ok, err = scanQueriedRow(rows,
			&s.SentID,
			&s.SubscriptionID,
			&s.ChatID,
			&s.VacancyID,
			&s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("cannot scan queried row: %v", err)
		}
		if !ok {
			break
		}
		sv = append(sv, s)
	}
	return sv, nil
}

func (s *storage) PutSentVacancy(ctx context.Context, sv *model.ChatSentVacancy) error {
	query := sanitizeQuery(
		`INSERT INTO chat_sent_vacancies(
            subscription_id,
            vacancy_id,
            created_at
        ) VALUES ($1, $2, $3)`)

	return retries.DoWithRetries(retryCount, retryWait, func() error {
		if _, err := s.client.Exec(ctx, query,
			postgres.MultiQuote(
				sv.SubscriptionID,
				sv.VacancyID,
				sv.CreatedAt,
			)...,
		); err != nil {
			return fmt.Errorf("cannot do postgres exec: %s: %v", query, err)
		}
		return nil
	})
}

func (s *storage) DeleteChatSubscription(ctx context.Context, subID int64) error {
	query := sanitizeQuery(
		`DELETE 
            FROM chat_subscriptions 
        WHERE subscription_id = $1`)

	return retries.DoWithRetries(retryCount, retryWait, func() error {
		if _, err := s.client.Exec(ctx, query, postgres.SingleQuote(subID)); err != nil {
			return fmt.Errorf("cannot do postgres exec: %s: %v", query, err)
		}
		return nil
	})
}

func scanQueriedRow(rows pgx.Rows, fields ...any) (bool, error) {
	var hasRow bool
	if rows.Next() {
		if err := rows.Scan(fields...); err != nil {
			return false, fmt.Errorf("cannot scan queried row: %v", err)
		}
		hasRow = true
	}
	return hasRow, nil
}

func sanitizeQuery(query string) string {
	query = regexQuery.ReplaceAllLiteralString(query, "")
	query = strings.TrimSpace(query)
	return query
}

var regexQuery = regexp.MustCompile(`\r\t\n`)
