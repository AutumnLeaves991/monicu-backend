package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"pkg.mon.icu/monicu/internal/config"
	"pkg.mon.icu/monicu/internal/discord"
	"pkg.mon.icu/monicu/internal/storage"
)

type app struct {
	ctx    context.Context
	cancel context.CancelFunc

	logConf zap.Config
	logger  *zap.Logger

	config *config.Config

	storage *storage.Storage
	discord *discord.Discord
}

func newApp(ctx context.Context, lcf zap.Config, log *zap.Logger) (*app, error) {
	ctx, cancel := context.WithCancel(ctx)
	a := &app{ctx: ctx, cancel: cancel, logConf: lcf, logger: log}
	var err error

	log.Debug("Loading configuration.")
	a.config, err = config.Read()
	if err != nil {
		return nil, fmt.Errorf("couldn't load configuration: %w", err)
	}

	log.Debug("Successfully loaded configuration (also switching log level.)")
	lcf.Level.SetLevel(a.config.Logging.Level)

	log.Debug("Initializing Storage struct.")
	a.storage = storage.NewStorage(ctx, log)

	log.Debug("Initializing Discord struct.")
	a.discord, err = discord.NewDiscord(ctx, log, a.config.Discord.Auth, discord.NewConfig(a.config.Discord.Guilds, a.config.Discord.Channels, a.config.Posts.IgnoreRegexp), a.storage)
	if err != nil {
		return nil, fmt.Errorf("couldn't initialize Discord struct: %w", err)
	}

	return a, nil
}

func (a *app) Run() error {
	a.logger.Debug("Connecting to PostgreSQL storage.")
	if err := a.storage.Connect(a.config.Storage.PostgresDSN); err != nil {
		return fmt.Errorf("couldn't connect to storage: %s", err)
	}
	defer func() {
		a.logger.Debug("Closing PostgreSQL storage.")
		if err := a.storage.Close(); err != nil {
			a.logger.Sugar().Errorf("Couldn't close storage: %s.", err)
		}
		a.logger.Debug("Closed PostgreSQL storage.")
	}()
	a.logger.Debug("Successfully connected to PostgreSQL storage.")

	a.logger.Debug("Connecting to Discord API gateway.")
	if err := a.discord.Connect(); err != nil {
		return fmt.Errorf("couldn't connect to Discord: %s", err)
	}
	defer func() {
		a.logger.Debug("Closing connection with Discord API gateway.")
		if err := a.discord.Close(); err != nil {
			a.logger.Sugar().Errorf("Couldn't close Discord: %s.", err)
		}
		a.logger.Debug("Closed connection with Discord API gateway.")
	}()
	a.logger.Debug("Successfully connected to Discord API gateway.")

	a.logger.Info("Launch complete. Send SIGINT to gracefully terminate.")
	<-a.ctx.Done()
	a.logger.Info("SIGINT received, terminating.")

	return a.ctx.Err()
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	lcf := zap.NewDevelopmentConfig() // to later switch level without reallocation
	lcf.Level.SetLevel(zapcore.DebugLevel)
	lcf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	lcf.DisableCaller = true
	log, _ := lcf.Build()

	log.Info("Initializing application.")
	a, err := newApp(ctx, lcf, log)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Sugar().Fatalf("Couldn't initialize application: %s.", err)
		}

		return
	}

	log.Debug("Initialization tasks complete, continuing with launch.")
	if err := a.Run(); err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Sugar().Fatalf("Application crashed: %s.", err)
		}
	}
}
