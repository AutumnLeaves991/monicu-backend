package discord

import (
	"context"
	"regexp"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"pkg.mon.icu/monicu/internal/storage"
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

type embedEdit struct {
	message  *discordgo.Message
	timer    *time.Timer
	stopChan chan struct{}
}

type Discord struct {
	ctx           context.Context
	logger        *zap.Logger
	session       *discordgo.Session
	handlerRemFns []func()
	config        *Config
	storage       *storage.Storage
	//queue          *queue.BlockingQueue
	embedEditSched map[string]*embedEdit
}

func NewDiscord(ctx context.Context, log *zap.Logger, auth string, config *Config, store *storage.Storage) (*Discord, error) {
	s, err := discordgo.New(auth)
	if err != nil {
		return nil, err
	}
	d := &Discord{
		ctx:            ctx,
		logger:         log,
		session:        s,
		handlerRemFns:  make([]func(), 0, 8),
		config:         config,
		storage:        store, /*queue: queue.New(),*/
		embedEditSched: make(map[string]*embedEdit),
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
	for _, hFn := range d.handlerRemFns {
		hFn()
	}
}

//func (d *Discord) handleTaskQueue() {
//	for {
//		task, ok := d.queue.Pop()
//		if !ok {
//			break
//		}
//
//		task.(func())() // invoke
//	}
//}

func (d *Discord) Connect() error {
	//go d.handleTaskQueue()
	d.addHandlers()
	return d.session.Open()
}

func (d *Discord) Close() error {
	//_ = d.queue.Close()
	d.removeHandlers()
	return d.session.Close()
}
