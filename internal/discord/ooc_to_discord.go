package discord

import (
	"fmt"
	"sync"
	"time"

	"github.com/ZeroHubProjects/discord-bot/internal/types"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

var (
	oocQueueSize    = 100
	oocMessageQueue = make(chan types.OOCMessage, oocQueueSize)
)

// ooc messages are enqueued so there is some buffer to accomodate for interruptions
// and allow webhook to immediately return to the game
func EnqueueOOCMessage(msg types.OOCMessage) error {
	select {
	case oocMessageQueue <- msg:
		return nil
	default:
		return fmt.Errorf("queue is full, %d messages", oocQueueSize)
	}
}

func RunOOCProcessingLoop(channelID string, discord *discordgo.Session, logger *zap.SugaredLogger, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		msg := <-oocMessageQueue
		formattedMessage := fmt.Sprintf("<t:%d:t> **%s**: %s", time.Now().Unix(), msg.SenderKey, msg.Message)
		_, err := discord.ChannelMessageSend(channelID, formattedMessage)
		if err != nil {
			logger.Errorf("failed to send ooc message to discord: %v", err)
		}
	}
}
