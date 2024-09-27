package status

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

const (
	serverName  = "ZeroOnyx"
	githubLink  = "Temporarily Private"
	serverColor = 0xFF355E // hex color as decimal, Radical Red
)

type statusUpdater struct {
	discord           *discordgo.Session
	ss13ServerAddress string
	statusChannelID   string
	logger            *zap.SugaredLogger
}

func (s *statusUpdater) update() error {
	serverStatus, err := getServerStatus(s.ss13ServerAddress)
	if err != nil {
		return fmt.Errorf("failed to get server status: %w", err)
	}

	msgs, err := s.discord.ChannelMessages(s.statusChannelID, 10, "", "", "")
	if err != nil {
		return fmt.Errorf("failed to get messages from the channel: %w", err)
	}

	var statusMessage *discordgo.Message
	for _, msg := range msgs {
		if len(msg.Embeds) > 0 && msg.Embeds[0].Title == serverName {
			statusMessage = msg
		}
	}

	newMessageDescription, err := s.getStatusMessageDescription(serverStatus)
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
		_, err := s.discord.ChannelMessageSendComplex(s.statusChannelID, &newMessage)
		if err != nil {
			return fmt.Errorf("failed to post message: %w", err)
		}
	} else {
		newEdit := discordgo.NewMessageEdit(s.statusChannelID, statusMessage.ID).SetEmbed(embed)
		_, err := s.discord.ChannelMessageEditComplex(newEdit)
		if err != nil {
			return fmt.Errorf("failed to update message: %w", err)
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
		ServerAddress: "byond://" + s.ss13ServerAddress,
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