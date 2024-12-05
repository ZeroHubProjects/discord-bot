package webhooks

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (h *webhookHandler) handleOOC(data string, authorized bool) response {
	if !h.oocEnabled {
		return getResponse(http.StatusServiceUnavailable, codeWebhookDisabled, "OOC webhook is currently disabled")
	}
	if !authorized {
		return ForbiddenResponse
	}
	if data == "" {
		return getResponse(http.StatusBadRequest, codeEmptyData, "A request `data` was expected, but is missing")
	}
	var msg OOCMessage
	err := json.Unmarshal([]byte(data), &msg)
	if err != nil {
		err := fmt.Errorf("failed to unmarshal data: %w", err)
		return getResponse(http.StatusBadRequest, codeMalformedData, err.Error())
	}
	if msg.SenderKey == "" || msg.Message == "" {
		return getResponse(http.StatusBadRequest, codeMalformedData, "Both `sender_key` and `message` are required in the `data`")
	}
	err = h.enqueueOOCMessage(msg)
	if err != nil {
		h.logger.Errorf("failed to enqueue message: %v", err)
		return getResponse(http.StatusServiceUnavailable, codeInternalServerError, err.Error())
	}
	return SuccessResponse
}

// ooc messages are enqueued so there is some buffer to accomodate for interruptions
// and allow webhook to immediately return to the game
func (h *webhookHandler) enqueueOOCMessage(msg OOCMessage) error {
	if h.oocMessageQueue == nil {
		return fmt.Errorf("ooc message queue is nil but ooc message is handled")
	}
	select {
	case h.oocMessageQueue <- msg:
		return nil
	default:
		return fmt.Errorf("queue is full, %d messages", len(h.oocMessageQueue))
	}
}
