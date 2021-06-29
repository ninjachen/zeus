package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"go.yym.plus/zeus/pkg/http/middleware"
	"go.yym.plus/zeus/pkg/log"
	"go.yym.plus/zeus/pkg/utils/structs"
)

type Config struct {
	Addr   string `validate:"required" default:":8080"`
	Logger struct {
		Enable *bool `default:"true"`
	}
}

type Engine struct {
	*gin.Engine
	config *Config
	server *http.Server
}

func NewServer(config *Config) (*Engine, error) {
	err := structs.SetDefaultsAndValidate(config)
	if err != nil {
		return nil, err
	}
	engine := gin.New()
	engine.Use(middleware.Recovery())
	if *config.Logger.Enable {
		engine.Use(middleware.Logger())
	}
	srv := &http.Server{
		Handler: engine,
	}

	return &Engine{Engine: engine, server: srv, config: config}, nil
}

func (self *Engine) Run() {
	go func() {
		self.server.Addr = self.config.Addr
		if err := self.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Error("http server run error")
		}
	}()
}

func (self *Engine) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := self.server.Shutdown(ctx); err != nil {
		log.WithError(err).Error("http server stop error")
	}
}
