package statusupdates

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	serverName  = "ZeroOnyx"
	githubLink  = "Temporarily Private"
	serverColor = 16725342 // discord accepts color in decimal, this is #FF355E aka Radical Red
)

type statusUpdater struct {
	Discord           *discordgo.Session
	SS13ServerAddress string
	StatusChannelID   string
}

func (s *statusUpdater) updateServerStatus() error {
	serverStatus, err := getServerStatus(s.SS13ServerAddress)
	if err != nil {
		return fmt.Errorf("failed to get server status: %w", err)
	}

	msgs, err := s.Discord.ChannelMessages(s.StatusChannelID, 10, "", "", "")
	if err != nil {
		return fmt.Errorf("failed to get messages from the status updates channel: %w", err)
	}

	var statusMessage *discordgo.Message
	for _, msg := range msgs {
		if len(msg.Embeds) > 0 && msg.Embeds[0].Title == serverName {
			statusMessage = msg
		}
	}

	newMessageDescription, err := s.getStatusMessageDescription(serverStatus)
	if err != nil {
		return fmt.Errorf("failed to create new status update message description: %w", err)
	}

	embed := &discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Title:       serverName,
		Description: newMessageDescription,
		Color:       serverColor,
	}

	if statusMessage == nil {
		newMessage := discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{embed}}
		_, err := s.Discord.ChannelMessageSendComplex(s.StatusChannelID, &newMessage)
		if err != nil {
			return fmt.Errorf("failed to post status message: %w", err)
		}
	} else {
		newEdit := discordgo.NewMessageEdit(s.StatusChannelID, statusMessage.ID).SetEmbed(embed)
		_, err := s.Discord.ChannelMessageEditComplex(newEdit)
		if err != nil {
			return fmt.Errorf("failed to update status message: %w", err)
		}
	}
	return nil
}

var currentUnixTimestamp = func() int64 { return time.Now().Unix() }
var statusMessageTmplFuncs = template.FuncMap{"currentUnixTimestamp": currentUnixTimestamp, "join": strings.Join}

func (s *statusUpdater) getStatusMessageDescription(serverStatus serverStatus) (string, error) {
	descPayloadParams := descriptionPayloadParams{
		Players:       serverStatus.Players,
		RoundTime:     serverStatus.RoundTime,
		Map:           serverStatus.Map,
		Evac:          serverStatus.Evac == 1,
		ServerAddress: "byond://" + s.SS13ServerAddress,
		GitHubLink:    githubLink,
	}
	descriptionTmpl := template.Must(template.
		New("statusMessageDescription").
		Funcs(statusMessageTmplFuncs).
		Parse(statusMessageDescriptionTemplate))

	var descBuf bytes.Buffer
	if err := descriptionTmpl.Execute(&descBuf, descPayloadParams); err != nil {
		return "", err
	}
	return descBuf.String(), nil
}
