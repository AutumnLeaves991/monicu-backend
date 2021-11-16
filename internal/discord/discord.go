package discord

import (
	"context"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"pkg.mon.icu/monicu/internal/storage"
)

type Config struct {
	guilds       uint64Set
	chans        uint64Set
	ignoreRegexp *regexp.Regexp
}

func NewConfig(guilds, channels []uint64, ignoreRegexp *regexp.Regexp) *Config {
	return &Config{
		guilds:       newUint64Set(guilds),
		chans:        newUint64Set(channels),
		ignoreRegexp: ignoreRegexp,
	}
}

type Discord struct {
	ctx     context.Context
	logger  *zap.Logger
	session *discordgo.Session
	config  *Config
	storage *storage.Storage
}

func NewDiscord(ctx context.Context, log *zap.Logger, auth string, config *Config, store *storage.Storage) (*Discord, error) {
	s, err := discordgo.New(auth)
	if err != nil {
		return nil, err
	}
	return &Discord{ctx: ctx, logger: log, session: s, config: config, storage: store}, nil
}

func (d *Discord) addHandlers() {
	d.session.AddHandlerOnce(d.onReady)
	d.session.AddHandler(d.onMessageCreate)
}

func (d *Discord) Connect() error {
	d.addHandlers()
	return d.session.Open()
}

func (d *Discord) Close() error {
	return d.session.Close()
}
