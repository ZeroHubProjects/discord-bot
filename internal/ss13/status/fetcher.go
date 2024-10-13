package status

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ZeroHubProjects/discord-bot/internal/ss13"
	"go.uber.org/zap"
)

type ServerStatus struct {
	Players   []string  `json:"playerlist"`
	RoundTime string    `json:"roundtime"`
	Map       string    `json:"map"`
	Evac      int       `json:"evac"`
	FetchedAt time.Time `json:"-"`
}

type ServerStatusFetcher struct {
	ServerAddress      string
	Logger             *zap.SugaredLogger
	latestServerStatus *ServerStatus
	mutex              sync.Mutex
}

func (s *ServerStatusFetcher) GetServerStatus(maxAge time.Duration) (*ServerStatus, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.latestServerStatus != nil && time.Since(s.latestServerStatus.FetchedAt) <= maxAge {
		return s.latestServerStatus, nil
	}

	newStatus, err := s.fetchServerStatus()
	if err != nil {
		return nil, err
	}

	s.latestServerStatus = newStatus
	return s.latestServerStatus, nil
}

func (s *ServerStatusFetcher) fetchServerStatus() (*ServerStatus, error) {
	var result ServerStatus
	resp, err := ss13.SendRequest(s.ServerAddress, []byte("discordstatus"))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	if resp == nil {
		result.RoundTime = "Unknown... (Server restarting or stopped)"
		result.Map = "Unknown..."
		return &result, nil
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}
	result.FetchedAt = time.Now().UTC()
	return &result, nil
}
