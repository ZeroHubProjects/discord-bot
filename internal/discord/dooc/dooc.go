package dooc

import (
	"fmt"
	"regexp"
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

func RunDOOC(ss13Cfg config.SS13Config, cfg config.DiscordConfig, discord *discordgo.Session, logger *zap.SugaredLogger, wg *sync.WaitGroup) {
	defer wg.Done()
	handler := &messageHandler{
		ss13Config:   ss13Cfg,
		oocChannelID: cfg.OOCChannelID,
		discord:      discord,
		logger:       logger,
	}
	logger.Debug("registering handler and listening for messages...")
	discord.AddHandler(handler.handleDOOC)

	// keep processing unsent messages if any
	for {
		msgs, err := discord.ChannelMessages(cfg.OOCChannelID, 50, "", "", "")
		if err != nil {
			logger.Errorf("failed to get messages from the channel: %w", err)
		}
		for _, msg := range msgs {
			if strings.HasPrefix(msg.Content, retryMarker) {
				handler.retryMessage(msg)
			}
		}
		time.Sleep(time.Minute)
	}
}

func (h *messageHandler) handleDOOC(sess *discordgo.Session, msg *discordgo.MessageCreate) {
	defer func() {
		if err := recover(); err != nil {
			h.logger.Errorf("handler panicked: %v", err)
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

	// delete old message from the user
	err := h.discord.ChannelMessageDelete(msg.ChannelID, msg.ID)
	if err != nil {
		h.logger.Errorf("failed to delete message from the channel: %v", err)
	}

	// TODO(rufus): verification handler

	// post a new formatted message with the same content
	formattedMessage := fmt.Sprintf("<t:%d:t> DOOC **%s**: %s", time.Now().Unix(), msg.Author.Username, msg.Content)
	doocMessage, err := h.discord.ChannelMessageSend(msg.ChannelID, formattedMessage)
	if err != nil {
		h.logger.Errorf("failed to send message to discord: %v", err)
	}

	// relay messsage to the game
	err = h.sendDOOCMessageToSS13(msg.Author.Username, msg.Content, h.ss13Config.ServerAddress, h.ss13Config.AccessKey)
	if err != nil {
		h.logger.Errorf("failed to send message to the game: %v", err)
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
	if resp == nil {
		return fmt.Errorf("empty response from the server")
	}
	h.logger.Debugf("dooc topic response: %s", string(resp))
	return nil
}

func (h *messageHandler) retryMessage(msg *discordgo.Message) {
	if !strings.HasPrefix(msg.Content, retryMarker) {
		h.logger.Errorf("message %s was retried without having a retry marker, contents: [%s]", msg.ID, msg.Content)
		return
	}

	botMsgContents := msg.Content[len(retryMarker)+1:]
	re := regexp.MustCompile(`<t:\d+:t> DOOC \*\*(.+?)\*\*: (.+)`)
	matches := re.FindStringSubmatch(botMsgContents)
	if len(matches) != 3 {
		h.logger.Warnf("failed to parse message for retry: %v", botMsgContents)
		return
	}
	err := h.sendDOOCMessageToSS13(matches[1], matches[2], h.ss13Config.ServerAddress, h.ss13Config.AccessKey)
	if err != nil {
		h.logger.Debugf("message retry failed: %v", err)
		return
	}
	// successful retry, message was sent!
	h.discord.ChannelMessageEdit(msg.ChannelID, msg.ID, botMsgContents)
}
