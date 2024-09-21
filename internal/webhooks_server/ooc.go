package webhooks_server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ZeroHubProjects/discord-bot/internal/util"
	"github.com/carlmjohnson/requests"
)

var (
	oocQueueSize    = 100
	oocMessageQueue = make(chan webhookOOCMessage, oocQueueSize)
)

// ooc messages are enqueued so there is some buffer to accomodate for interruptions
// and allow webhook to immediately return to the game
func enqueueOOCMessage(msg webhookOOCMessage) error {
	select {
	case oocMessageQueue <- msg:
	default:
		return fmt.Errorf("queue is full, %d messages", oocQueueSize)
	}
	return fmt.Errorf("unhandled ooc queue operation")
}

func runOOCProcessingLoop() {
	for {
		err := SendOOCToDiscord(<-oocMessageQueue)
		if err != nil {
			logger.Errorf("failed to send ooc message to discord: %v", err)
		}
	}
}

func SendOOCToDiscord(msg webhookOOCMessage) error {
	formattedMessage := fmt.Sprintf("<t:%d:t> **%s**: %s", time.Now().Unix(), msg.Ckey, msg.Message)
	payload, err := json.Marshal(discordMessage{Content: formattedMessage})
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	return postDiscordMessage(payload, globalConfig.Modules.Webhooks.OOC.DiscordChannelID, globalConfig.DiscordBotToken)
}

func postDiscordMessage(payload []byte, channelID, token string) error {
	postMessageApiURL := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID)
	return requests.
		URL(postMessageApiURL).
		Header("Content-Type", "application/json").
		Header("Authorization", "Bot "+token).
		BodyBytes(payload).
		AddValidator(util.PrintErrBodyValidationHandler()).
		Fetch(context.Background())
}
