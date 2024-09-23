package types

type OOCMessage struct {
	SenderKey string `json:"sender_key"`
	Message   string `json:"message"`
}

type TopicResponse struct {
	Code string `json:"code"`
}
