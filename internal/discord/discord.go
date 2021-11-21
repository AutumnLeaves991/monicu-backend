package discord

import (
	"context"
	"regexp"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"pkg.mon.icu/monicu/internal/storage"
	"pkg.mon.icu/monicu/internal/storage/model"
)

type Config struct {
	guilds       Uint64Set
	chans        Uint64Set
	ignoreRegexp *regexp.Regexp
}

func NewConfig(guilds, channels []uint64, ignoreRegexp *regexp.Regexp) *Config {
	return &Config{
		guilds:       NewUint64Set(guilds),
		chans:        NewUint64Set(channels),
		ignoreRegexp: ignoreRegexp,
	}
}

type Discord struct {
	ctx    context.Context
	logger *zap.SugaredLogger

	session       *discordgo.Session
	handlerRemFns []func()

	config                *Config
	storage               *storage.Storage
	channelGuildRelations map[uint64]uint64
}

func NewDiscord(ctx context.Context, log *zap.SugaredLogger, auth string, config *Config, store *storage.Storage) (*Discord, error) {
	s, err := discordgo.New(auth)
	if err != nil {
		return nil, err
	}

	d := &Discord{
		ctx:                   ctx,
		logger:                log,
		session:               s,
		handlerRemFns:         make([]func(), 0, 8),
		config:                config,
		storage:               store,
		channelGuildRelations: make(map[uint64]uint64),
	}

	return d, nil
}

func (d *Discord) addHandlers() {
	d.handlerRemFns = append(d.handlerRemFns, d.session.AddHandlerOnce(d.onReady))
	for _, h := range []interface{}{
		d.onMessageUpdate,
		d.onMessageCreate,
		d.onMessageDelete,
		d.onMessageDeleteBulk,
		d.onMessageReactionAdd,
		d.onMessageReactionRemove,
		d.onMessageReactionRemoveAll,
	} {
		d.handlerRemFns = append(d.handlerRemFns, d.session.AddHandler(h))
	}
}

func (d *Discord) removeHandlers() {
	for _, removeHandler := range d.handlerRemFns {
		removeHandler()
	}
}

func (d *Discord) buildChannelGuildCache() {
	for chanID := range d.config.chans {
		chann, err := d.session.Channel(strconv.FormatUint(chanID, 10))
		if err != nil {
			d.logger.Errorf("Failed to retrieve channel %d: %s.", chanID, err)
			continue
		}

		d.channelGuildRelations[chanID] = model.MustParseSnowflake(chann.GuildID)
	}
}

func (d *Discord) Connect() error {
	d.addHandlers()
	return d.session.Open()
}

func (d *Discord) Close() error {
	d.removeHandlers()
	return d.session.Close()
}
