package webhooks

type OOCMessage struct {
	SenderKey string `json:"sender_key"`
	Message   string `json:"message"`
}

type EmoteMessage struct {
	SenderKey string `json:"sender_key"`
	Name      string `json:"name"`
	Message   string `json:"message"`
}

type AhelpMessage struct {
	SenderKey string `json:"sender_key"`
	TargetKey string `json:"target_key"`
	Message   string `json:"message"`
}
