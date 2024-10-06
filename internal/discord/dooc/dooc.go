package dooc

import (
	"fmt"
	"net/url"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/ZeroHubProjects/discord-bot/internal/discord/verification"
	"github.com/ZeroHubProjects/discord-bot/internal/ss13"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const retryMarker = ":hourglass:"

type DOOCHandler struct {
	SS13ServerAddress string
	SS13AccessKey     string
	OOCChannelID      string
	Discord           *discordgo.Session
	Logger            *zap.SugaredLogger
	// if VerificationHandler is nil, BYOND account verification won't be required and Discord username will be used as the sender name
	VerificationHandler *verification.ByondVerificationHandler
}

func (h *DOOCHandler) handleDOOCMessage(sess *discordgo.Session, msg *discordgo.MessageCreate) {
	defer func() {
		if err := recover(); err != nil {
			h.Logger.Errorf("handler panicked: %v\nstack trace: %s", err, string(debug.Stack()))
		}
	}()
	// ignore own messages
	if msg.Author.ID == sess.State.User.ID {
		return
	}
	if msg.ChannelID != h.OOCChannelID {
		return
	}

	// delete old message from the user
	err := h.Discord.ChannelMessageDelete(msg.ChannelID, msg.ID)
	if err != nil {
		h.Logger.Errorf("failed to delete message from the channel: %v", err)
	}

	senderName := msg.Author.Username
	// check verification
	if h.VerificationHandler != nil {
		player, err := h.VerificationHandler.GetVerifiedPlayer(msg.Author.ID)
		if err != nil {
			h.Logger.Errorf("failed to check if player is already verified: %v", err)
			return
		}
		if player == nil {
			err := h.VerificationHandler.SendUserToVerification(msg.Author.ID)
			if err != nil {
				h.Logger.Errorf("failed to send user to verification: %v", err)
			}
			return
		}
		senderName = player.DisplayKey
	}

	// post a new formatted message with the same content
	formattedMessage := fmt.Sprintf("<t:%d:t> DOOC **%s**: %s", time.Now().Unix(), senderName, msg.Content)
	doocMessage, err := h.Discord.ChannelMessageSend(msg.ChannelID, formattedMessage)
	if err != nil {
		h.Logger.Errorf("failed to send message to discord: %v", err)
	}

	// relay messsage to the game
	err = h.sendDOOCMessageToSS13(senderName, msg.Content)
	if err != nil {
		h.Logger.Errorf("failed to send message to the game: %v", err)
		h.Discord.ChannelMessageEdit(doocMessage.ChannelID, doocMessage.ID, fmt.Sprintf(":hourglass: %s", doocMessage.Content))
	}
}

func (h *DOOCHandler) sendDOOCMessageToSS13(sender, message string) error {
	h.Logger.Debugf("sending dooc:  %s: %s", sender, message)
	message = url.QueryEscape(message)
	request := fmt.Sprintf("dooc&sender_key=%s&message=%s&key=%s", sender, message, h.SS13AccessKey)
	resp, err := ss13.SendRequest(h.SS13ServerAddress, []byte(request))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	if resp == nil {
		return fmt.Errorf("empty response from the server")
	}
	h.Logger.Debugf("dooc topic response: %s", string(resp))
	return nil
}

func (h *DOOCHandler) retryMessage(msg *discordgo.Message) {
	if !strings.HasPrefix(msg.Content, retryMarker) {
		h.Logger.Errorf("message %s was retried without having a retry marker, contents: [%s]", msg.ID, msg.Content)
		return
	}

	botMsgContents := msg.Content[len(retryMarker)+1:]
	re := regexp.MustCompile(`<t:\d+:t> DOOC \*\*([^*]+?)\*\*: (.+)`)
	matches := re.FindStringSubmatch(botMsgContents)
	if len(matches) != 3 {
		h.Logger.Warnf("failed to parse message for retry: %v", botMsgContents)
		return
	}
	err := h.sendDOOCMessageToSS13(matches[1], matches[2])
	if err != nil {
		h.Logger.Debugf("message retry failed: %v", err)
		return
	}
	// successful retry, message was sent!
	h.Discord.ChannelMessageEdit(msg.ChannelID, msg.ID, botMsgContents)
}
