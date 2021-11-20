package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"pkg.mon.icu/monicu/internal/storage"
)

type Config struct {
	Port uint16
}

func NewConfig(port uint16) *Config {
	return &Config{Port: port}
}

type API struct {
	ctx     context.Context
	logger  *zap.SugaredLogger
	storage *storage.Storage
	router  *gin.Engine
	serv    *http.Server
}

func NewAPI(ctx context.Context, logger *zap.SugaredLogger, storage *storage.Storage, config *Config) *API {
	a := &API{
		ctx:     ctx,
		logger:  logger,
		storage: storage,
		router:  gin.New(),
	}
	a.serv = &http.Server{Addr: fmt.Sprintf(":%d", config.Port), Handler: a.router}
	return a
}

func (a *API) Listen() {
	a.registerGetPosts()
	go func() {
		if err := a.serv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				a.logger.Errorf("Server returned with error: %s.", err)
			}
		}
	}()
}

func (a *API) Close() error {
	return a.serv.Close()
}
