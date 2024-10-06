package relay

import (
	"fmt"
	"sync"
	"time"

	"github.com/ZeroHubProjects/discord-bot/internal/webhooks"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const oocInterval = time.Minute

type OOCRelay struct {
	Queue     chan webhooks.OOCMessage
	ChannelID string
	Discord   *discordgo.Session
	Logger    *zap.SugaredLogger
}

func (r *OOCRelay) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	r.Logger.Debug("listening to the queue...")

	for {
		r.runOOCRelay()
		r.Logger.Debugf("restarting in %v...", oocInterval)
		time.Sleep(oocInterval)
	}
}

func (r *OOCRelay) runOOCRelay() {
	for {
		msg := <-r.Queue
		formattedMessage := fmt.Sprintf("<t:%d:t> **%s**: %s", time.Now().Unix(), msg.SenderKey, msg.Message)
		_, err := r.Discord.ChannelMessageSend(r.ChannelID, formattedMessage)
		if err != nil {
			r.Logger.Errorf("failed to send ooc message to discord: %v", err)
		}
	}
}
