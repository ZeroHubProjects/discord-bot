package webhooks_server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

func recoverMiddleware(logger *zap.SugaredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Errorf("handler panicked: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(InternalErrorResponse)
				}
			}()
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func webhookRequestToString(request url.Values) string {
	parts := []string{}
	for k := range request {
		val := request.Get(k)
		if k == "key" {
			val = "(redacted)"
		}
		if k == "data" {
			val = url.QueryEscape(val)
		}

		parts = append(parts, fmt.Sprintf("[%s: %s]", k, val))
	}
	if len(parts) < 1 {
		parts = []string{"(empty)"}
	}
	return strings.Join(parts, ", ")
}
