package webhooks

type OOCMessage struct {
	SenderKey string `json:"sender_key"`
	Message   string `json:"message"`
}
