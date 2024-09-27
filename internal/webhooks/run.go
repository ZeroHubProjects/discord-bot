package webhooks

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

const interval = time.Minute

type WebhookServer struct {
	Port               int
	SS13AccessKey      string
	OOCMessagesEnabled bool
	Logger             *zap.SugaredLogger
}

func (s *WebhookServer) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		s.runServer()
		s.Logger.Debugf("restarting in %v...", interval)
		time.Sleep(interval)
	}
}

func (s *WebhookServer) runServer() {
	defer func() {
		if err := recover(); err != nil {
			s.Logger.Errorf("panicked: %v", err)
		}
	}()

	handler := webhookHandler{
		accessKey:  s.SS13AccessKey,
		oocEnabled: s.OOCMessagesEnabled,
		logger:     s.Logger,
	}

	r := chi.NewRouter()
	r.Use(middleware.SetHeader("Content-Type", "application/json"))
	r.Use(recoverMiddleware(s.Logger))
	// TODO(rufus): api documentation
	r.Get("/", handler.handleRequest)

	s.Logger.Debugf("listening on port %d...", s.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", s.Port), r)
}
