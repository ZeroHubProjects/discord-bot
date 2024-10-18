package status

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/ZeroHubProjects/discord-bot/internal/ss13"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const (
	serverName  = "ZeroOnyx"
	githubLink  = "Temporarily Private"
	serverColor = 0xFF355E // hex color as decimal, Radical Red
)

type StatusUpdater struct {
	Discord           *discordgo.Session
	SS13ServerAddress string
	StatusChannelIDs  []string
	StatusFetcher     *ss13.ServerStatusFetcher
	Logger            *zap.SugaredLogger
}

func (s *StatusUpdater) update(status *ss13.ServerStatus, channelID string) error {
	if status == nil {
		return fmt.Errorf("nil server status passed")
	}

	msgs, err := s.Discord.ChannelMessages(channelID, 10, "", "", "")
	if err != nil {
		return fmt.Errorf("failed to get messages from the channel: %w", err)
	}

	var statusMessage *discordgo.Message
	for _, msg := range msgs {
		if len(msg.Embeds) > 0 && msg.Embeds[0].Title == serverName {
			statusMessage = msg
		}
	}

	newMessageDescription, err := s.getStatusMessageDescription(status)
	if err != nil {
		return fmt.Errorf("failed to create new message description: %w", err)
	}

	embed := &discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Title:       serverName,
		Description: newMessageDescription,
		Color:       serverColor,
	}

	if statusMessage == nil {
		newMessage := discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{embed}}
		_, err := s.Discord.ChannelMessageSendComplex(channelID, &newMessage)
		if err != nil {
			return fmt.Errorf("failed to post message: %w", err)
		}
	} else {
		newEdit := discordgo.NewMessageEdit(channelID, statusMessage.ID).SetEmbed(embed)
		_, err := s.Discord.ChannelMessageEditComplex(newEdit)
		if err != nil {
			return fmt.Errorf("failed to update message: %w", err)
		}
	}
	return nil
}

var currentUnixTimestamp = func() int64 { return time.Now().Unix() }
var statusMessageTmplFuncs = template.FuncMap{"currentUnixTimestamp": currentUnixTimestamp, "join": strings.Join}

func (s *StatusUpdater) getStatusMessageDescription(serverStatus *ss13.ServerStatus) (string, error) {
	if serverStatus == nil {
		return "", fmt.Errorf("nil server status passed")
	}
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
