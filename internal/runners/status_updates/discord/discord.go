package discord

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/ZeroHubProjects/discord-bot/internal/runners/status_updates/ss13"
	"github.com/ZeroHubProjects/discord-bot/internal/util"
	"github.com/carlmjohnson/requests"
	"go.uber.org/zap"
)

const (
	serverName  = "ZeroOnyx"
	githubLink  = "Temporarily Private"
	serverColor = "16725342" // discord accepts color in decimal, this is #FF355E aka Radical Red
)

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}
type Message struct {
	ID     string  `json:"id"`
	Author User    `json:"author"`
	Embeds []Embed `json:"embeds"`
}
type Embed struct {
	Title string `json:"title"`
}

func UpdateServerStatus(token, ss13ServerAddress, channelID string, logger *zap.SugaredLogger, ctx context.Context) error {
	serverStatus, err := ss13.GetServerStatus(ss13ServerAddress, ctx)
	if err != nil {
		return fmt.Errorf("failed to get server status: %w", err)
	}

	statusMessage, err := findStatusMessage(channelID, token, ctx)
	if err != nil {
		return fmt.Errorf("failed to find status message: %w", err)
	}
	message := getStatusMessagePayload(ss13ServerAddress, serverStatus)
	if statusMessage == nil {
		if err := postStatusMessage(message, channelID, token, ctx); err != nil {
			return fmt.Errorf("failed to post status message: %w", err)
		}
	} else {
		if err := updateStatusMessage(message, statusMessage.ID, channelID, token, ctx); err != nil {
			return fmt.Errorf("failed to update status message: %w", err)
		}
	}
	return nil
}

func findStatusMessage(channelID, token string, ctx context.Context) (*Message, error) {
	if channelID == "" {
		return nil, fmt.Errorf("empty channel ID")
	}
	if token == "" {
		return nil, fmt.Errorf("empty access token")
	}
	messages, err := getMessagesByChannelID(channelID, token, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get status channel messages: %w", err)
	}
	for _, message := range messages {
		if len(message.Embeds) > 0 && message.Embeds[0].Title == serverName {
			return &message, nil
		}
	}
	return nil, nil
}

func postStatusMessage(payload, channelID, token string, ctx context.Context) error {
	postMessageApiURL := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID)
	return requests.
		URL(postMessageApiURL).
		Header("Content-Type", "application/json").
		Header("Authorization", "Bot "+token).
		BodyBytes([]byte(payload)).
		AddValidator(util.PrintErrBodyValidationHandler()).
		Fetch(ctx)
}

func updateStatusMessage(message, messageID, channelID, token string, ctx context.Context) error {
	patchMessageApiURL := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages/%s", channelID, messageID)
	return requests.
		URL(patchMessageApiURL).
		Header("Content-Type", "application/json").
		Header("Authorization", "Bot "+token).
		BodyBytes([]byte(message)).
		AddValidator(util.PrintErrBodyValidationHandler()).
		Patch().
		Fetch(ctx)
}

func getMessagesByChannelID(channelID, token string, ctx context.Context) ([]Message, error) {
	var messages []Message
	messagesApiURL := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID)
	err := requests.
		URL(messagesApiURL).
		Header("Authorization", "Bot "+token).
		ToJSON(&messages).
		AddValidator(util.PrintErrBodyValidationHandler()).
		Fetch(ctx)
	return messages, err
}

var currentUnixTimestamp = func() int64 { return time.Now().Unix() }
var statusMessageTmplFuncs = template.FuncMap{"currentUnixTimestamp": currentUnixTimestamp, "join": strings.Join}

func getStatusMessagePayload(serverAddress string, serverStatus ss13.ServerStatus) string {
	// prepare description of the embed from the template
	descPayloadParams := descriptionPayloadParams{
		Players:       serverStatus.Players,
		RoundTime:     serverStatus.RoundTime,
		Map:           serverStatus.Map,
		Evac:          serverStatus.Evac == 1,
		ServerAddress: "byond://" + serverAddress,
		GitHubLink:    githubLink,
	}
	descriptionTmpl := template.Must(template.
		New("statusMessageDescription").
		Funcs(statusMessageTmplFuncs).
		Parse(statusMessageDescriptionTemplate))

	var descBuf bytes.Buffer
	if err := descriptionTmpl.Execute(&descBuf, descPayloadParams); err != nil {
		err := fmt.Errorf("failed to execute description template: %w", err)
		panic(err)
	}

	// prepare the final embed payload
	payloadParams := statusMessagePayloadParams{
		Title:       template.JSEscapeString(serverName),
		Description: template.JSEscapeString(descBuf.String()),
		Color:       serverColor,
	}
	payloadTmpl := template.Must(template.
		New("statusMessagePayload").
		Parse(statusMessagePayloadTemplate))
	var payloadBuf bytes.Buffer
	if err := payloadTmpl.Execute(&payloadBuf, payloadParams); err != nil {
		err := fmt.Errorf("failed to execute payload template: %w", err)
		panic(err)
	}
	return payloadBuf.String()
}
