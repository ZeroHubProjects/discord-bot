package metrics

import (
	"runtime/debug"
	"sync"
	"time"

	"github.com/ZeroHubProjects/discord-bot/internal/database"
	"github.com/ZeroHubProjects/discord-bot/internal/ss13"
	"go.uber.org/zap"
)

const interval = time.Minute

type PlayerCountRecorder struct {
	Database      *database.Database
	StatusFetcher *ss13.ServerStatusFetcher
	Logger        *zap.SugaredLogger
}

func (r *PlayerCountRecorder) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		r.runRecorder()
		r.Logger.Debugf("restarting in %v...", interval)
		time.Sleep(interval)
	}
}

func (r *PlayerCountRecorder) runRecorder() {
	defer func() {
		if err := recover(); err != nil {
			r.Logger.Errorf("panicked: %v\nstack trace: %s", err, string(debug.Stack()))
		}
	}()

	r.Logger.Debugf("updating with %v interval...", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		serverStatus, err := r.StatusFetcher.GetServerStatus(interval)
		if err != nil {
			r.Logger.Errorf("failed to fetch server status: %v", err)
			continue
		}
		err = r.Database.InsertPlayerCount(len(serverStatus.Players))
		if err != nil {
			r.Logger.Errorf("failed to insert: %v", err)
			continue
		}
	}
}
