package relay

import (
	"fmt"
	"sync"
	"time"

	"github.com/ZeroHubProjects/discord-bot/internal/webhooks"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const ahelpInterval = time.Minute

type AhelpRelay struct {
	Queue     chan webhooks.AhelpMessage
	ChannelID string
	Discord   *discordgo.Session
	Logger    *zap.SugaredLogger
}

func (r *AhelpRelay) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	r.Logger.Debug("listening to the queue...")

	for {
		r.runAhelpRelay()
		r.Logger.Debugf("restarting in %v...", ahelpInterval)
		time.Sleep(ahelpInterval)
	}
}

func (r *AhelpRelay) runAhelpRelay() {
	for {
		msg := <-r.Queue
		targetPart := ""
		if msg.TargetKey != "" {
			targetPart = fmt.Sprintf(" -> **%s**", msg.TargetKey)
		}
		formattedMessage := fmt.Sprintf("<t:%d:t> **%s**%s: %s", time.Now().Unix(), msg.SenderKey, targetPart, msg.Message)
		_, err := r.Discord.ChannelMessageSend(r.ChannelID, formattedMessage)
		if err != nil {
			r.Logger.Errorf("failed to send ahelp message to discord: %v", err)
		}
	}
}