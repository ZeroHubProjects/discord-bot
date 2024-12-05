package webhooks

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (h *webhookHandler) handleAhelp(data string, authorized bool) response {
	if !h.ahelpEnabled {
		return getResponse(http.StatusServiceUnavailable, codeWebhookDisabled, "Ahelp webhook is currently disabled")
	}
	if !authorized {
		return ForbiddenResponse
	}
	if data == "" {
		return getResponse(http.StatusBadRequest, codeEmptyData, "A request `data` was expected, but is missing")
	}
	var msg AhelpMessage
	err := json.Unmarshal([]byte(data), &msg)
	if err != nil {
		err := fmt.Errorf("failed to unmarshal data: %w", err)
		return getResponse(http.StatusBadRequest, codeMalformedData, err.Error())
	}
	if msg.SenderKey == "" || msg.Message == "" {
		return getResponse(http.StatusBadRequest, codeMalformedData, "Both `sender_key` and `message` are required in the `data`")
	}
	err = h.enqueueAhelpMessage(msg)
	if err != nil {
		h.logger.Errorf("failed to enqueue message: %v", err)
		return getResponse(http.StatusServiceUnavailable, codeInternalServerError, err.Error())
	}
	return SuccessResponse
}

// ahelp messages are enqueued so there is some buffer to accomodate for interruptions
// and allow webhook to immediately return to the game
func (h *webhookHandler) enqueueAhelpMessage(msg AhelpMessage) error {
	if h.ahelpMessageQueue == nil {
		return fmt.Errorf("ahelp message queue is nil but ahelp message is handled")
	}
	select {
	case h.ahelpMessageQueue <- msg:
		return nil
	default:
		return fmt.Errorf("queue is full, %d messages", len(h.ahelpMessageQueue))
	}
}
