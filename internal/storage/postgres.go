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

func (s *storage) UserSubscriptions(ctx context.Context, userID int64) ([]*model.Subscription, error) {
	query := sanitizeQuery(
		`SELECT 
            subscription_id,
            user_id,
            chat_id,
            area,
            keywords,
            experience,
            created_at
        FROM subscriptions WHERE user_id = $1`)

	var (
		rows pgx.Rows
		err  error
	)
	if err = retries.DoWithRetries(retryCount, retryWait, func() error {
		rows, err = s.client.Query(ctx, query, postgres.SingleQuote(userID))
		if err != nil {
			return fmt.Errorf("cannot do postgres query: %s: %v", query, err)
		}
		return nil

	}); err != nil {
		return nil, err
	}
	var (
		subs []*model.Subscription
		ok   bool
	)
	for {
		sub := &model.Subscription{}

		if ok, err = scanQueriedRow(rows,
			&sub.SubscriptionID,
			&sub.UserID,
			&sub.ChatID,
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

func (s *storage) PutUserSubscription(ctx context.Context, sub *model.Subscription) error {
	query := sanitizeQuery(
		`INSERT INTO subscriptions(
            subscription_id,
            user_id,
            chat_id,
            area,
            keywords,
            experience,
            created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
	)
	return retries.DoWithRetries(retryCount, retryWait, func() error {
		if _, err := s.client.Exec(ctx, query,
			postgres.MultiQuote(
				sub.SubscriptionID,
				sub.UserID,
				sub.ChatID,
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

func (s *storage) UsersSubscriptions(ctx context.Context, sf func(sub *model.Subscription)) error {
	query := sanitizeQuery(
		`SELECT
            subscription_id,
            user_id,
            chat_id,
            area,
            keywords,
            experience,
            created_at
        FROM subscriptions`)

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
		sub := &model.Subscription{}

		if ok, err = scanQueriedRow(rows,
			&sub.SubscriptionID,
			&sub.UserID,
			&sub.ChatID,
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
		sf(sub)
	}
	return nil
}

func (s *storage) UsersSubscriptionsSets(ctx context.Context, sf func(subset *model.SubscriptionSet)) error {
	query := sanitizeQuery(
		`SELECT DISTINCT
            subscription_ids,
            user_ids,
            chat_ids,
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
                ARRAY_AGG(user_id) OVER (PARTITION BY area, LOWER(keywords)) as user_ids,
                ARRAY_AGG(chat_id) OVER (PARTITION BY area, LOWER(keywords)) as chat_ids
            FROM subscriptions
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
		sub := &model.SubscriptionSet{}

		if ok, err = scanQueriedRow(rows,
			&sub.SubscriptionIDs,
			&sub.UserIDs,
			&sub.ChatIDs,
			&sub.Area,
			&sub.Keywords,
			&sub.Experience,
		); err != nil {
			return fmt.Errorf("cannot scan queried row: %v", err)
		}
		if !ok {
			break
		}
		sf(sub)
	}
	return nil
}

func (s *storage) SentSubscriptions(ctx context.Context) ([]*model.SentSubscription, error) {
	query := sanitizeQuery(
		`SELECT  
              sent_id, 
              subscription_id,
              created_at
        FROM sent_subscriptions`)

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
		st []*model.SentSubscription
		ok bool
	)
	for {
		s := &model.SentSubscription{}

		if ok, err = scanQueriedRow(rows,
			&s.SentID,
			&s.SubscriptionID,
			&s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("cannot scan queried row: %v", err)
		}
		if !ok {
			break
		}
		st = append(st, s)
	}
	return st, nil
}

func (s *storage) SentVacancies(ctx context.Context) ([]*model.SentVacancy, error) {
	query := sanitizeQuery(
		`SELECT
            sent_id, 
            chat_id,
            vacancy_id, 
            created_at
    FROM sent_vacancies`)

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
		sv []*model.SentVacancy
		ok bool
	)
	for {
		s := &model.SentVacancy{}

		if ok, err = scanQueriedRow(rows,
			&s.SentID,
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

func (s *storage) PutSentSubscription(ctx context.Context, st *model.SentSubscription) error {
	query := sanitizeQuery(
		`INSERT INTO sent_subscriptions(
            sent_id,
            subscription_id,
            created_at
        ) VALUES ($1, $2, $3)`)

	return retries.DoWithRetries(retryCount, retryWait, func() error {
		if _, err := s.client.Exec(ctx, query,
			postgres.MultiQuote(
				st.SentID,
				st.SubscriptionID,
				st.CreatedAt,
			)...,
		); err != nil {
			return fmt.Errorf("cannot do postgres exec: %s: %v", query, err)
		}
		return nil
	})
}

func (s *storage) PutSentVacancy(ctx context.Context, sv *model.SentVacancy) error {
	query := sanitizeQuery(
		`INSERT INTO sent_vacancies(
            sent_id, 
            chat_id,
            vacancy_id,
            created_at
        ) VALUES ($1, $2, $3, $4)`)

	return retries.DoWithRetries(retryCount, retryWait, func() error {
		if _, err := s.client.Exec(ctx, query,
			postgres.MultiQuote(
				sv.SentID,
				sv.ChatID,
				sv.VacancyID,
				sv.CreatedAt,
			)...,
		); err != nil {
			return fmt.Errorf("cannot do postgres exec: %s: %v", query, err)
		}
		return nil
	})
}

func (s *storage) DeleteUserSubscription(ctx context.Context, subID string) error {
	query := sanitizeQuery(
		`DELETE 
            FROM subscriptions 
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
