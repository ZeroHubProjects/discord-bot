package verification

import (
	"fmt"
	"time"

	"github.com/ZeroHubProjects/discord-bot/internal/database"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const verificationLifetime = 5 * time.Minute

type ByondVerificationHandler struct {
	Discord   *discordgo.Session
	ChannelID string
	Logger    *zap.SugaredLogger
	Database  *database.Database
}

func (h *ByondVerificationHandler) handleInteraction(sess *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var userID string
	if interaction.User != nil {
		userID = interaction.User.ID
	} else if interaction.Member != nil && interaction.Member.User != nil {
		userID = interaction.Member.User.ID
	} else {
		h.Logger.Errorf("failed to obtain user ID from the interaction: %+v", *interaction.Interaction)
		return
	}
	if interaction.Type == discordgo.InteractionMessageComponent && interaction.MessageComponentData().CustomID == buttonID {
		// verification button, respond with modal
		err := sess.InteractionRespond(interaction.Interaction, verificationModal)
		if err != nil {
			h.Logger.Errorf("failed to respond with interaction modal")
		}
		return
	}
	if interaction.Type == discordgo.InteractionMessageComponent && interaction.MessageComponentData().CustomID == exitButtonID {
		// exit button, kick user out of the channel
		h.exitVerificationChannel(userID)
		return
	}
	if interaction.Type == discordgo.InteractionModalSubmit {
		if interaction.ModalSubmitData().CustomID == modalID {
			// modal response, fetch the code and process it
			code := interaction.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
			verified, response, err := h.handleVerificationRequest(code, userID)
			if err != nil {
				h.Logger.Errorf("failed to handle verification: %v", err)
				err := sess.InteractionRespond(interaction.Interaction, newEphemeralInteractionResponse(lzRus[lzUnknownErrorResp]))
				if err != nil {
					h.Logger.Errorf("failed to send unhandled error response")
				}
				return
			}

			resp := newEphemeralInteractionResponse(response)
			if verified {
				h.Logger.Debug("user %s successfully verified", userID)
				resp.Data.Content += "\nКанал автоматически закроется через 5 минут. Альтернативно, можете использовать кнопку под этим сообщением."
				resp.Data.Components = exitVerificationChannelComponent
				go func() {
					time.Sleep(5 * time.Minute)
					h.exitVerificationChannel(userID)
				}()
			}
			err = sess.InteractionRespond(interaction.Interaction, resp)
			if err != nil {
				h.Logger.Errorf("failed to respond to a modal submission: %v", err)
				return
			}
		}
	}
}

func (h *ByondVerificationHandler) handleVerificationRequest(code, userID string) (passed bool, response string, err error) {
	player, err := h.GetVerifiedPlayer(userID)
	if err != nil {
		return false, "", fmt.Errorf("failed to check if player is already verified: %w", err)
	}
	if player != nil {
		return true, lzRus[lzAlreadyVerifiedResp], nil
	}

	v, err := h.Database.FetchVerification(code)
	if err != nil {
		return false, "", fmt.Errorf("failed to look up verification code: %w", err)
	}
	if v == nil {
		return false, lzRus[lzVerificationNotFoundResp], nil
	}

	if time.Now().UTC().After(v.CreatedAt.Add(verificationLifetime)) {
		err := h.Database.DeleteVerification(code)
		if err != nil {
			h.Logger.Errorf("failed to delete outdated verification: %v", err)
		}
		return false, lzRus[lzVerificationExpiredResp], nil
	}

	err = h.Database.CreateVerifiedAccountEntry(*v, userID)
	if err != nil {
		return false, "", fmt.Errorf("failed to create a verified account entry: %w", err)
	}
	err = h.Database.DeleteVerification(code)
	if err != nil {
		h.Logger.Errorf("failed to delete used verification: %v", err)
	}

	return true, lzRus[lzSuccessfullyVerifiedResp], nil
}

func (h *ByondVerificationHandler) GetVerifiedPlayer(userID string) (*database.VerifiedPlayer, error) {
	return h.Database.GetVerifiedPlayer(userID)
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
		// this is not our message, let's do some cleanup before we post a new one
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
	// post a new message
	h.Logger.Debug("posting a new verification message")
	_, err = h.Discord.ChannelMessageSendComplex(h.ChannelID, &verificationMessage)
	if err != nil {
		h.Logger.Errorf("failed to send verification message: %v", err)
	}
}
