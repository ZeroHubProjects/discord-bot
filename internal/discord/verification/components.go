package verification

import "github.com/bwmarrin/discordgo"

const (
	embedTitle       = "BYOND Account Verification"
	embedColor       = 0x2554C7 // hex color as decimal, Sapphire Blue
	buttonID         = "byond_verification_button"
	buttonLabel      = "Verify"
	embedDescription = `This is a test instruction. You can test the "Verify" button if you want!`
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
	modalID          = "byond_verification_modal"
	modalTitle       = "Enter Verification Code"
	inputID          = "verification_code_input"
	inputLabel       = "Verification Code"
	inputPlaceholder = "Enter the code from the game"
)

var verificationModal = &discordgo.InteractionResponse{
	Type: discordgo.InteractionResponseModal,
	Data: &discordgo.InteractionResponseData{
		CustomID: modalID,
		Title:    modalTitle,
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
