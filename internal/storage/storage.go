package storage

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

type Storage struct {
	ctx    context.Context
	logger *zap.Logger
	pool   *pgxpool.Pool
}

func NewStorage(ctx context.Context, l *zap.Logger) *Storage {
	return &Storage{ctx: ctx, logger: l}
}

func (s *Storage) Connect(dsn string) error {
	var err error
	s.pool, err = pgxpool.Connect(s.ctx, dsn)
	return err
}

func (s *Storage) Begin(ctx context.Context, fn func(pgx.Tx) error) error {
	return s.pool.BeginFunc(ctx, fn)
}

func (s *Storage) QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{}) (pgconn.CommandTag, error) {
	return s.pool.QueryFunc(ctx, sql, args, scans, func(pgx.QueryFuncRow) error { return nil })
}

func (s *Storage) Close() error {
	s.pool.Close()
	return nil
}
