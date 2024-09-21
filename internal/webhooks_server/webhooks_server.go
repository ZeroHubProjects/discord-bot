package webhooks_server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/ZeroHubProjects/discord-bot/internal/config"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var globalConfig config.Config
var logger *zap.SugaredLogger

func Run(cfg config.Config, wg *sync.WaitGroup) {
	defer wg.Done()

	globalConfig = cfg
	logger = cfg.Logger.Named("webhooks")

	// message queue workers so messages have some buffer even during service interruptions
	go runOOCProcessingLoop()

	// router
	r := chi.NewRouter()
	// TODO(rufus): api documentation
	// TODO(rufus): full request logging with credentials filtering
	r.Get("/", WebhookRequestHandler)

	logger.Debugf("Webhooks server listening on port %d...", cfg.Modules.Webhooks.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Modules.Webhooks.Port), recoverMiddleware(r))
}

func recoverMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("handler panicked: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(InternalErrorResponse)
			}
		}()
		h.ServeHTTP(w, r)
	})
}
