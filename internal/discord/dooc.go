package discord

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ZeroHubProjects/discord-bot/internal/config"
	"github.com/ZeroHubProjects/discord-bot/internal/ss13"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type messageHandler struct {
	ss13Config   config.SS13Config
	oocChannelID string
	logger       *zap.SugaredLogger
}

func RunDOOC(ss13Cfg config.SS13Config, cfg config.DiscordConfig, dg *discordgo.Session, logger *zap.SugaredLogger, wg *sync.WaitGroup) {
	handler := &messageHandler{ss13Config: ss13Cfg, oocChannelID: cfg.OOCChannelID, logger: logger}
	logger.Debug("registering dooc message handler...")
	dg.AddHandler(handler.handle)

	// and then just wait here until the end of times (or an interrupt)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}

func (h *messageHandler) handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ignore own messages
	if m.Author.ID == s.State.User.ID {
		return
	}
	// currently only process ooc messages
	if m.ChannelID != h.oocChannelID {
		return
	}

	// TODO(rufus): permission check

	// relay messsage into the game
	h.sendDOOCMessage(m.Author.Username, m.Content, h.ss13Config.ServerAddress, h.ss13Config.AccessKey)

	// delete old message from the user
	err := s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		h.logger.Errorf("failed to delete a DOOC message from the channel: %v", err)
	}

	// post a new formatted message with the same content
	formattedMessage := fmt.Sprintf("<t:%d:t> `**[DOOC]**` **%s**: %s", time.Now().Unix(), m.Author.Username, m.Content)
	_, err = s.ChannelMessageSend(m.ChannelID, formattedMessage)
	if err != nil {
		h.logger.Errorf("failed to send dooc message to discord: %v", err)
	}

}

func (h *messageHandler) sendDOOCMessage(sender, message, serverAddress, accessKey string) error {
	h.logger.Debugf("sending dooc:  %s: %s", sender, message)
	request := fmt.Sprintf("dooc&sender_key=%s&message=%s&key=%s", sender, message, accessKey)
	resp, err := ss13.SendRequest(serverAddress, []byte(request))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	h.logger.Debugf("dooc topic response: %s", string(resp))
	return nil
}
