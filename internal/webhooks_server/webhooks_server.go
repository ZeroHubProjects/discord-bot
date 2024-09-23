package webhooks_server

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/ZeroHubProjects/discord-bot/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func Run(cfg config.WebhooksConfig, logger *zap.SugaredLogger, wg *sync.WaitGroup) {
	defer wg.Done()

	logger = logger.Named("webhooks")

	handler := webhookHandler{
		accessKey:  cfg.AccessKey,
		oocEnabled: cfg.OOCMessagesEnabled,
		logger:     logger,
	}

	r := chi.NewRouter()
	r.Use(middleware.SetHeader("Content-Type", "application/json"))
	r.Use(recoverMiddleware(logger))
	// TODO(rufus): api documentation
	r.Get("/", handler.handleRequest)

	logger.Debugf("listening on port %d...", cfg.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), r)
}
