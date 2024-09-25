package verification

import (
	"sync"
	"time"

	"github.com/ZeroHubProjects/discord-bot/internal/config"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type ByondVerificationHandler struct {
	Discord   *discordgo.Session
	ChannelID string
	Logger    *zap.SugaredLogger
}

func RunBYONDVerification(cfg config.DiscordConfig, discord *discordgo.Session, logger *zap.SugaredLogger, wg *sync.WaitGroup) {
	defer wg.Done()

	handler := ByondVerificationHandler{
		Discord:   discord,
		ChannelID: cfg.BYONDVerificationChannelID,
		Logger:    logger,
	}

	logger.Debug("checking verification message and registering handlers...")
	handler.updateVerificationMessage()
	discord.AddHandler(handler.handleInteraction)

	for {
		// NOTE(rufus): add routine tasks as required
		time.Sleep(time.Minute)
	}
}

func (h *ByondVerificationHandler) handleInteraction(sess *discordgo.Session, interaction *discordgo.InteractionCreate) {
	// only our button clicks
	if interaction.Type != discordgo.InteractionMessageComponent || interaction.MessageComponentData().CustomID != buttonID {
		return
	}
	// respond with modal
	err := sess.InteractionRespond(interaction.Interaction, verificationModal)
	if err != nil {
		h.Logger.Errorf("failed to respond with interaction modal")
	}
}

func (h *ByondVerificationHandler) updateVerificationMessage() {
	// our message should be the last one in the channel, match the title and match the description
	msgs, err := h.Discord.ChannelMessages(h.ChannelID, 1, "", "", "")
	if err != nil {
		h.Logger.Errorf("failed to get messages from the channel: %w", err)
		return
	}
	if len(msgs) != 0 {
		// found *some* message in the channel
		lastMsg := msgs[0]
		if len(lastMsg.Embeds) > 0 && lastMsg.Embeds[0].Title == embedTitle && lastMsg.Embeds[0].Description == embedDescription {
			// message is correct, no need to do anything
			return
		}
		// this is not our message, let's do some cleanup
		msgs, err := h.Discord.ChannelMessages(h.ChannelID, 100, "", "", "")
		if err != nil {
			h.Logger.Errorf("failed to get messages from the channel: %w", err)
			return
		}
		h.Logger.Debug("cleaning up outdated messages")
		for _, msg := range msgs {
			if len(msg.Embeds) > 0 && msg.Embeds[0].Title == embedTitle {
				err := h.Discord.ChannelMessageDelete(msg.ChannelID, msg.ID)
				if err != nil {
					h.Logger.Errorf("failed to delete outdated message: %v", err)
					// don't return, it's fine, let humans handle this manually
				}
			}
		}
	}
	h.Logger.Debug("posting a new verification message")
	_, err = h.Discord.ChannelMessageSendComplex(h.ChannelID, &verificationMessage)
	if err != nil {
		h.Logger.Errorf("failed to send verification message: %v", err)
	}
}
