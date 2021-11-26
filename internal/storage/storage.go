package storage

import (
	"context"
	"database/sql"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"moul.io/zapgorm2"
	"pkg.mon.icu/monicu/internal/storage/model"
)

type Storage struct {
	ctx    context.Context
	logger *zap.SugaredLogger
	db *gorm.DB
}

func NewStorage(ctx context.Context, l *zap.SugaredLogger) *Storage {
	return &Storage{ctx: ctx, logger: l}
}

func (s *Storage) Connect(dsn string) error {
	var err error
	s.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{SkipDefaultTransaction: true, Logger: zapgorm2.New(s.logger.Desugar())})
	if err := s.db.AutoMigrate(&model.Channel{}, &model.Emoji{}, &model.Guild{}, &model.Image{}, &model.Post{}, &model.Reaction{}, &model.User{}, &model.UserReaction{}); err != nil {
		return err
	}
	return err
}

func (s *Storage) Transaction(fn func(db *gorm.DB) error, opts ...*sql.TxOptions) error {
	return s.db.Transaction(fn, opts...)
}

func (s *Storage) Close() error {
	db, err := s.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
