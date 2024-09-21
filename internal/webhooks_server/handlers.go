package webhooks_server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ZeroHubProjects/discord-bot/internal/util"
)

func WebhookRequestHandler(w http.ResponseWriter, r *http.Request) {
	response := handleRequest(r)

	if response.statusCode == http.StatusNoContent {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.statusCode)
	json.NewEncoder(w).Encode(response)

}

func handleRequest(r *http.Request) response {
	query := r.URL.Query()
	util.DebugPrintWebhookRequest(query, logger)

	requestType := query.Get("type")
	providedKey := query.Get("key")

	var authorized bool
	if providedKey == webhooksConfig.AccessKey {
		authorized = true
	} else {
		if providedKey != "" {
			// it's ok to log invalid keys
			logger.Infof("invalid key authorization attempted: got %v", providedKey)
		}
	}

	data, err := url.QueryUnescape(query.Get("data"))
	if err != nil {
		err := fmt.Errorf("failed to url decode data: %w", err)
		logger.Info(err)
		return GetBadRequestResponse(codeMalformedData, err.Error())
	}

	var resp response
	switch requestType {
	case "ooc":
		resp, err = handleOOC(data, authorized)
	case "":
		return WelcomeResponse
	default:
		return GetTeapotResponse(codeUnknownRequestType, fmt.Sprintf("Webhook requests of type `%s` are not supported", requestType))
	}

	if err != nil {
		logger.Infof("failed to handle webhook: %v", err)
	}
	return resp

}

func handleOOC(data string, authorized bool) (response, error) {
	if !webhooksConfig.OOC.Enabled {
		return GetServiceUnavailableResponse(codeWebhookDisabled, "OOC webhook is currently disabled"), nil
	}
	if !authorized {
		return ForbiddenResponse, nil
	}
	if data == "" {
		return GetBadRequestResponse(codeEmptyData, "A request `data` was expected, but is missing"), nil
	}
	var msg webhookOOCMessage
	err := json.Unmarshal([]byte(data), &msg)
	if err != nil {
		err := fmt.Errorf("failed to unmarshal data: %w", err)
		return GetBadRequestResponse(codeMalformedData, err.Error()), err
	}
	if msg.Ckey == "" || msg.Message == "" {
		return GetBadRequestResponse(codeMalformedData, "Both `ckey` and `message` are required in the `data`"), nil
	}
	enqueueOOCMessage(msg)
	return SuccessResponse, nil
}
