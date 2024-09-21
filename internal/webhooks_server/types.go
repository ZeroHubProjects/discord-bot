package webhooks_server

type webhookOOCMessage struct {
	Ckey    string `json:"ckey"`
	Message string `json:"message"`
}

type discordMessage struct {
	Content string `json:"content"`
}
