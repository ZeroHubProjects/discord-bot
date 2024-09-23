package discord

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ZeroHubProjects/discord-bot/internal/config"
	"github.com/ZeroHubProjects/discord-bot/internal/ss13"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const retryMarker = ":hourglass:"

type messageHandler struct {
	ss13Config   config.SS13Config
	oocChannelID string
	discord      *discordgo.Session
	logger       *zap.SugaredLogger
}

func RunDOOC(ss13Cfg config.SS13Config, cfg config.DiscordConfig, dg *discordgo.Session, logger *zap.SugaredLogger, wg *sync.WaitGroup) {
	defer wg.Done()
	handler := &messageHandler{ss13Config: ss13Cfg, oocChannelID: cfg.OOCChannelID, discord: dg, logger: logger}
	logger.Debug("registering handler and listening for messages...")
	dg.AddHandler(handler.handle)

	// keep processing unsent messages if any
	for {
		msgs, err := dg.ChannelMessages(cfg.OOCChannelID, 50, "", "", "")
		if err != nil {
			logger.Errorf("failed to get messages from the ooc channel: %w", err)
		}
		for _, msg := range msgs {
			if strings.HasPrefix(msg.Content, retryMarker) {
				handler.retryMessage(msg)
			}
		}
		time.Sleep(time.Minute)
	}
}

func (h *messageHandler) handle(sess *discordgo.Session, msg *discordgo.MessageCreate) {
	defer func() {
		if err := recover(); err != nil {
			h.logger.Errorf("dooc message handler panicked: %v", err)
		}
	}()
	// ignore own messages
	if msg.Author.ID == sess.State.User.ID {
		return
	}
	// currently only process dooc messages
	if msg.ChannelID != h.oocChannelID {
		return
	}

	// TODO(rufus): permission check

	// delete old message from the user
	err := h.discord.ChannelMessageDelete(msg.ChannelID, msg.ID)
	if err != nil {
		h.logger.Errorf("failed to delete a DOOC message from the channel: %v", err)
	}

	// post a new formatted message with the same content
	formattedMessage := fmt.Sprintf("<t:%d:t> DOOC **%s**: %s", time.Now().Unix(), msg.Author.Username, msg.Content)
	doocMessage, err := h.discord.ChannelMessageSend(msg.ChannelID, formattedMessage)
	if err != nil {
		h.logger.Errorf("failed to send dooc message to discord: %v", err)
	}

	// relay messsage into the game
	err = h.sendDOOCMessageToSS13(msg.Author.Username, msg.Content, h.ss13Config.ServerAddress, h.ss13Config.AccessKey)
	if err != nil {
		h.logger.Errorf("failed to send dooc message into the game: %v", err)
		h.discord.ChannelMessageEdit(doocMessage.ChannelID, doocMessage.ID, fmt.Sprintf(":hourglass: %s", doocMessage.Content))
	}
}

func (h *messageHandler) sendDOOCMessageToSS13(sender, message, serverAddress, accessKey string) error {
	h.logger.Debugf("sending dooc:  %s: %s", sender, message)
	request := fmt.Sprintf("dooc&sender_key=%s&message=%s&key=%s", sender, message, accessKey)
	resp, err := ss13.SendRequest(serverAddress, []byte(request))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	h.logger.Debugf("dooc topic response: %s", string(resp))
	return nil
}

func (h *messageHandler) retryMessage(m *discordgo.Message) {
	if !strings.HasPrefix(m.Content, retryMarker) {
		h.logger.Errorf("message %s was retried without having a retry marker, contents: [%s]", m.ID, m.Content)
		return
	}
	msgContent := m.Content[len(retryMarker)+1:]
	err := h.sendDOOCMessageToSS13(m.Author.Username, msgContent, h.ss13Config.ServerAddress, h.ss13Config.AccessKey)
	if err != nil {
		h.logger.Debugf("dooc message retry failed: %v", err)
		return
	}
	// successful retry, message was sent!
	h.discord.ChannelMessageEdit(m.ChannelID, m.ID, msgContent)
}
