package verification

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func (h *ByondVerificationHandler) SendUserToVerification(userID string) error {
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
	return nil
}
