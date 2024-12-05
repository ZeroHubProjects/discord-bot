package webhooks

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (h *webhookHandler) handleEmote(data string, authorized bool) response {
	if !h.emoteEnabled {
		return getResponse(http.StatusServiceUnavailable, codeWebhookDisabled, "Emote webhook is currently disabled")
	}
	if !authorized {
		return ForbiddenResponse
	}
	if data == "" {
		return getResponse(http.StatusBadRequest, codeEmptyData, "A request `data` was expected, but is missing")
	}
	var msg EmoteMessage
	err := json.Unmarshal([]byte(data), &msg)
	if err != nil {
		err := fmt.Errorf("failed to unmarshal data: %w", err)
		return getResponse(http.StatusBadRequest, codeMalformedData, err.Error())
	}
	if msg.SenderKey == "" || msg.Name == "" || msg.Message == "" {
		return getResponse(http.StatusBadRequest, codeMalformedData, "`sender_key`, `name`, and `message` are required in the `data`")
	}
	err = h.enqueueEmoteMessage(msg)
	if err != nil {
		h.logger.Errorf("failed to enqueue message: %v", err)
		return getResponse(http.StatusServiceUnavailable, codeInternalServerError, err.Error())
	}
	return SuccessResponse
}

// emote messages are enqueued so there is some buffer to accomodate for interruptions
// and allow webhook to immediately return to the game
func (h *webhookHandler) enqueueEmoteMessage(msg EmoteMessage) error {
	if h.emoteMessageQueue == nil {
		return fmt.Errorf("emote message queue is nil but emote message is handled")
	}
	select {
	case h.emoteMessageQueue <- msg:
		return nil
	default:
		return fmt.Errorf("queue is full, %d messages", len(h.emoteMessageQueue))
	}
}
