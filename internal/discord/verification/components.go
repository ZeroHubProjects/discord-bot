package verification

import (
	"github.com/bwmarrin/discordgo"
)

const (
	embedColor = 0x2554C7 // hex color as decimal, Sapphire Blue
	buttonID   = "byond_verification_button"
)

var (
	embedTitle       = lzRus[lzInstructionsTitle]
	embedDescription = lzRus[lzInstructions]
	buttonLabel      = lzRus[lzButtonLabel]
)

var verificationMessage = discordgo.MessageSend{
	Embeds: []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeRich,
			Title:       embedTitle,
			Description: embedDescription,
			Color:       embedColor,
		},
	},
	Components: []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    buttonLabel,
					Style:    discordgo.PrimaryButton,
					CustomID: buttonID,
				},
			},
		},
	},
}

const (
	modalID = "byond_verification_modal"
	inputID = "verification_code_input"
)

var (
	inputLabel       = lzRus[lzInputLabel]
	inputPlaceholder = lzRus[lzInputPlaceholder]
)

var verificationModal = &discordgo.InteractionResponse{
	Type: discordgo.InteractionResponseModal,
	Data: &discordgo.InteractionResponseData{
		CustomID: modalID,
		Title:    embedTitle,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    inputID,
						Label:       inputLabel,
						Style:       discordgo.TextInputShort,
						Placeholder: inputPlaceholder,
						Required:    true,
					},
				},
			},
		},
	},
}

func newEphemeralInteractionResponse(content string) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}
}

const exitButtonID = "exit_verification_button"

var exitButtonLabel = lzRus[lzExitChannelButtonLabel]

var exitVerificationChannelComponent = []discordgo.MessageComponent{
	discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    exitButtonLabel,
				Style:    discordgo.SecondaryButton,
				CustomID: exitButtonID,
			},
		},
	},
}
