package ahelp

import (
	"encoding/json"
	"fmt"
	"regexp"
	"runtime/debug"
	"time"

	"github.com/ZeroHubProjects/discord-bot/internal/discord/verification"
	"github.com/ZeroHubProjects/discord-bot/internal/ss13"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const (
	respStatusSuccess  = "success"
	respStatusNoClient = "no_client"
)

type AhelpHandler struct {
	SS13ServerAddress   string
	SS13AccessKey       string
	AhelpChannelID      string
	Discord             *discordgo.Session
	Logger              *zap.SugaredLogger
	VerificationHandler *verification.ByondVerificationHandler
}

func (h *AhelpHandler) handleAhelpMessage(sess *discordgo.Session, msg *discordgo.MessageCreate) {
	defer func() {
		if err := recover(); err != nil {
			h.Logger.Errorf("handler panicked: %v\nstack trace: %s", err, string(debug.Stack()))
		}
	}()
	// ignore own messages
	if msg.Author.ID == sess.State.User.ID {
		return
	}
	if msg.ChannelID != h.AhelpChannelID {
		return
	}

	// delete old message from the user
	err := h.Discord.ChannelMessageDelete(msg.ChannelID, msg.ID)
	if err != nil {
		h.Logger.Errorf("failed to delete message from the channel: %v", err)
	}

	// check verification
	if h.VerificationHandler == nil {
		h.Logger.Errorf("can't check if user is verified with a nil verification handler")
		return
	}
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
	senderKey := player.DisplayKey

	// get the target key
	targetKey, err := h.extractTargetKeyFromReply(sess, msg)
	if err != nil {
		h.Logger.Errorf("failed to extract target key from the reply: %v", err)
		return
	}
	if targetKey == nil {
		formattedMessage := fmt.Sprintf(":x: No response target found, you must reply to an existing message!\n<t:%d:t> DAhelp **%s**: %s", time.Now().Unix(), senderKey, msg.Content)
		_, err := h.Discord.ChannelMessageSend(msg.ChannelID, formattedMessage)
		if err != nil {
			h.Logger.Errorf("failed to send message to discord: %v", err)
		}
		return
	}

	// post a new formatted message with the same content
	formattedMessage := fmt.Sprintf("<t:%d:t> DAhelp **%s** -> **%s**: %s", time.Now().Unix(), senderKey, *targetKey, msg.Content)
	dahelpMessage, err := h.Discord.ChannelMessageSend(msg.ChannelID, formattedMessage)
	if err != nil {
		h.Logger.Errorf("failed to send message to discord: %v", err)
	}

	// relay message to the game
	respStatus, err := h.sendAhelpMessageToSS13(senderKey, *targetKey, msg.Content)
	if err != nil {
		h.Logger.Errorf("failed to send message to the game: %v", err)
		editedMessageContent := fmt.Sprintf(":x: Failed to send message: %s\n%s", err, dahelpMessage.Content)
		h.Discord.ChannelMessageEdit(dahelpMessage.ChannelID, dahelpMessage.ID, editedMessageContent)
		return
	}
	if respStatus == respStatusNoClient {
		h.Logger.Debugf("target key not found on the server: %s", msg.ID)
		editedMessageContent := fmt.Sprintf(":x: %s is not online on the server\n%s", *targetKey, dahelpMessage.Content)
		h.Discord.ChannelMessageEdit(dahelpMessage.ChannelID, dahelpMessage.ID, editedMessageContent)
		return
	}
	if respStatus != respStatusSuccess {
		h.Logger.Debugf("unexpected topic response status: %s, msg_id: %s", respStatus, msg.ID)
		editedMessageContent := fmt.Sprintf(":x: unexpected topic response status: %s\n%s", respStatus, dahelpMessage.Content)
		h.Discord.ChannelMessageEdit(dahelpMessage.ChannelID, dahelpMessage.ID, editedMessageContent)
		return
	}
}

func (h *AhelpHandler) sendAhelpMessageToSS13(sender, target, message string) (responseCode string, err error) {
	h.Logger.Debugf("sending ahelp:  %s -> %s: %s", sender, target, message)
	request := fmt.Sprintf("ahelp&sender_key=%s&target_key=%s&message=%s&key=%s", sender, target, message, h.SS13AccessKey)
	resp, err := ss13.SendRequest(h.SS13ServerAddress, []byte(request))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	if resp == nil {
		return "", fmt.Errorf("empty response from the server")
	}
	h.Logger.Debugf("ahelp topic response: %s", string(resp))
	topicResp := map[string]string{}
	err = json.Unmarshal(resp, &topicResp)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal topic response")
	}
	return topicResp["code"], nil
}

func (h *AhelpHandler) extractTargetKeyFromReply(sess *discordgo.Session, msg *discordgo.MessageCreate) (*string, error) {
	var targetKey string
	// check if we can actually find the reply
	reference := msg.MessageReference
	if reference == nil {
		h.Logger.Debugf("target key not found as message doesn't have a reply reference, msg_id: %s", msg.ID)
		return nil, nil
	}
	repliedMsg, err := h.Discord.ChannelMessage(msg.ChannelID, reference.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch replied message: %w", err)
	}
	if repliedMsg.Author.ID != sess.State.User.ID {
		h.Logger.Debugf("target key not found as message is not from this bot, msg_id: %s", msg.ID)
		return nil, nil
	}
	// regular ahelp format -> `<t:123123123:t> **UserCkey**: message
	re := regexp.MustCompile(`<t:\d+:t> \*\*([^*]+?)\*\*: .+`)
	matches := re.FindStringSubmatch(repliedMsg.Content)
	if len(matches) == 2 {
		targetKey = matches[1]
		h.Logger.Debugf("found target key in a regular ahelp message, responding to: %s", targetKey)
		return &targetKey, nil
	}
	// direct PM ahelp format -> `<t:123123123:t> **AdminCkey** -> **UserCkey**: message
	re = regexp.MustCompile(`<t:\d+:t> \*\*([^*]+?)\*\* -> \*\*[^*]+\*\*: .+`)
	matches = re.FindStringSubmatch(repliedMsg.Content)
	if len(matches) == 2 {
		targetKey = matches[1]
		h.Logger.Debugf("found target key in a regular ahelp message, responding to: %s", targetKey)
		return &targetKey, nil
	}
	// discord ahelp format -> `<t:123123123:t> DAhelp **UserCkey**: message
	re = regexp.MustCompile(`<t:\d+:t> DAhelp \*\*([^*]+?)\*\*: .+`)
	matches = re.FindStringSubmatch(repliedMsg.Content)
	if len(matches) == 2 {
		targetKey = matches[1]
		h.Logger.Debugf("found target key in a dahelp message, responding to: %s", targetKey)
		return &targetKey, nil
	}
	// discord direct PM ahelp format -> `<t:123123123:t> DAhelp **AdminCkey** -> **UserCkey**: message
	re = regexp.MustCompile(`<t:\d+:t> DAhelp \*\*([^*]+?)\*\* -> \*\*[^*]+\*\*: .+`)
	matches = re.FindStringSubmatch(repliedMsg.Content)
	if len(matches) == 2 {
		targetKey = matches[1]
		h.Logger.Debugf("found target key in a dahelp message, responding to: %s", targetKey)
		return &targetKey, nil
	}
	h.Logger.Debugf("target key not found as message didn't match any formats, msg_id: %s", msg.ID)
	return nil, nil
}
