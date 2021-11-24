package storage

import (
	"context"
	"database/sql"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"pkg.mon.icu/monicu/internal/storage/model"
)

type Storage struct {
	ctx    context.Context
	logger *zap.SugaredLogger
	//pool   *pgxpool.Pool
	db *gorm.DB
}

func NewStorage(ctx context.Context, l *zap.SugaredLogger) *Storage {
	return &Storage{ctx: ctx, logger: l}
}

func (s *Storage) Connect(dsn string) error {
	var err error
	//s.pool, err = pgxpool.Connect(s.ctx, dsn)
	//if err != nil {
	//	return err
	//}
	s.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{SkipDefaultTransaction: true})
	if err := s.db.AutoMigrate(&model.Channel{}, &model.Emoji{}, &model.Guild{}, &model.Image{}, &model.Post{}, &model.Reaction{}, &model.User{}, &model.UserReaction{}); err != nil {
		return err
	}
	return err
}

func (s *Storage) Transaction(fn func(db *gorm.DB) error, opts ...*sql.TxOptions) error {
	return s.db.Transaction(fn, opts...)
}

//
//func (s *Storage) Begin(ctx context.Context, fn func(pgx.Tx) error) error {
//	return s.pool.BeginFunc(ctx, fn)
//}
//
//func (s *Storage) QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{}) (pgconn.CommandTag, error) {
//	return s.pool.QueryFunc(ctx, sql, args, scans, func(pgx.QueryFuncRow) error { return nil })
//}

func (s *Storage) Close() error {
	//s.pool.Close()
	db, err := s.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
