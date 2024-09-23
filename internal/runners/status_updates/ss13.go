package statusupdates

import (
	"encoding/json"
	"fmt"

	"github.com/ZeroHubProjects/discord-bot/internal/ss13"
)

type serverStatus struct {
	Players   []string `json:"playerlist"`
	RoundTime string   `json:"roundtime"`
	Map       string   `json:"map"`
	Evac      int      `json:"evac"`
}

func getServerStatus(serverAddress string) (serverStatus, error) {
	var result serverStatus
	resp, err := ss13.SendRequest(serverAddress, []byte("discordstatus"))
	if err != nil {
		return result, fmt.Errorf("failed to send request: %w", err)
	}
	if resp == nil {
		result.RoundTime = "Unknown... (Server restarting or stopped)"
		result.Map = "Unknown..."
		return result, nil
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return result, fmt.Errorf("failed to unmarshal: %w", err)
	}
	return result, nil
}
