package webhooks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

type webhookHandler struct {
	accessKey         string
	oocEnabled        bool
	ahelpEnabled      bool
	oocMessageQueue   chan OOCMessage
	ahelpMessageQueue chan AhelpMessage
	logger            *zap.SugaredLogger
}

func (h *webhookHandler) handleRequest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	h.logger.Debugf("received webhook: %s", webhookRequestToString(query))

	authorized := h.checkAuthorization(query.Get("key"))

	data, err := url.QueryUnescape(query.Get("data"))
	if err != nil {
		err := fmt.Errorf("failed to url decode data: %w", err)
		h.handleResponse(getResponse(http.StatusBadRequest, codeMalformedData, err.Error()), w)
	}

	var resp response
	switch query.Get("type") {
	case "ooc":
		resp = h.handleOOC(data, authorized)
	case "ahelp":
		resp = h.handleAhelp(data, authorized)
	case "":
		resp = WelcomeResponse
	default:
		resp = getResponse(http.StatusTeapot, codeUnknownRequestType)
	}

	h.handleResponse(resp, w)
}

func (h *webhookHandler) checkAuthorization(key string) bool {
	if key != h.accessKey {
		// it's ok to log invalid keys
		h.logger.Infof("invalid key authorization attempted: got %v", key)
		return false
	}
	return true
}

func (h *webhookHandler) handleResponse(r response, w http.ResponseWriter) {
	h.logger.Debugf("response: %d %s %s", r.statusCode, r.Code, r.Details)
	w.WriteHeader(r.statusCode)
	err := json.NewEncoder(w).Encode(r)
	if err != nil {
		h.logger.Error("failed to encode response")
	}
}
