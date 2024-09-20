package util

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/carlmjohnson/requests"
	"go.uber.org/zap"
)

func PrintErrBodyValidationHandler() requests.ResponseHandler {
	return requests.ValidatorHandler(requests.DefaultValidator, func(r *http.Response) error {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		return fmt.Errorf("unexpected status code: %d, response body: %s", r.StatusCode, string(b))
	})
}

func DebugPrintWebhookRequest(request url.Values, logger *zap.SugaredLogger) {
	if logger.Level() > zap.DebugLevel {
		// debug level is disabled, skip
		return
	}
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
	logger.Debugf("received webhook: %s", strings.Join(parts, ", "))
}
