package verification

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

const cooldown = time.Minute

// map of userID to last time they were sent to verification
var lastUserVerificationSend = map[string]time.Time{}

func (h *ByondVerificationHandler) SendUserToVerification(userID string) error {
	// check if user was recently sent to verification to avoid spamming
	if t, ok := lastUserVerificationSend[userID]; ok && !time.Now().UTC().After(t.Add(cooldown)) {
		return nil
	}
	allow := int64(discordgo.PermissionViewChannel)
	deny := int64(0)
	permType := discordgo.PermissionOverwriteTypeMember
	err := h.Discord.ChannelPermissionSet(h.ChannelID, userID, permType, allow, deny)
	if err != nil {
		return fmt.Errorf("failed to grant permissions: %w", err)
	}
	msg, err := h.Discord.ChannelMessageSend(h.ChannelID, fmt.Sprintf("<@%s>", userID))
	if err != nil {
		return fmt.Errorf("failed to ping user: %w", err)
	}
	// immediately delete the ping which creates a ghost-ping and just points the user towards the verification channel
	err = h.Discord.ChannelMessageDelete(msg.ChannelID, msg.ID)
	if err != nil {
		return fmt.Errorf("failed to delete ping: %w", err)
	}
	lastUserVerificationSend[userID] = time.Now().UTC()
	return nil
}
