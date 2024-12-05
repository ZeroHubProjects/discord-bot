package relay

import (
	"fmt"
	"sync"
	"time"

	"github.com/ZeroHubProjects/discord-bot/internal/webhooks"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const emoteInterval = time.Minute

type EmoteRelay struct {
	Queue     chan webhooks.EmoteMessage
	ChannelID string
	Discord   *discordgo.Session
	Logger    *zap.SugaredLogger
}

func (r *EmoteRelay) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	r.Logger.Debug("listening to the queue...")

	for {
		r.runEmoteRelay()
		r.Logger.Debugf("restarting in %v...", emoteInterval)
		time.Sleep(emoteInterval)
	}
}

func (r *EmoteRelay) runEmoteRelay() {
	for {
		msg := <-r.Queue
		formattedMessage := fmt.Sprintf("<t:%d:t> **%s** (%s): %s", time.Now().Unix(), msg.SenderKey, msg.Name, msg.Message)
		_, err := r.Discord.ChannelMessageSend(r.ChannelID, formattedMessage)
		if err != nil {
			r.Logger.Errorf("failed to send emote message to discord: %v", err)
		}
	}
}
