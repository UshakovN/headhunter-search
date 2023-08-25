package postgres

import (
	"context"
	"fmt"
	"main/pkg/retries"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"

	"github.com/jackc/pgx/v4/pgxpool"
)

const connTimeout = 5 * time.Second

type Client interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	BeginTxFunc(ctx context.Context, txOptions pgx.TxOptions, f func(pgx.Tx) error) error
}

type (
	ExecFunc  func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryFunc func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
)

func NewClient(ctx context.Context, config *Config) (Client, error) {
	const (
		retryCount = 5
		retryWait  = 3 * time.Second
	)
	var (
		pgxConn *pgxpool.Pool
		err     error
	)
	strConn := config.ConnectString()

	if err = retries.DoWithRetries(retryCount, retryWait, func() error {
		ctx, cancel := context.WithTimeout(ctx, connTimeout)
		defer cancel()

		if pgxConn, err = pgxpool.Connect(ctx, strConn); err != nil {
			return fmt.Errorf("%w: cannot connect to posgtres pgx driver: %v", retries.ErrDoRetry, err)
		}
		return nil

	}); err != nil {
		return nil, fmt.Errorf("connection to posgtres pgx driver failed: %v", err)
	}
	return pgxConn, nil
}
