package statusupdates

import (
	"context"
	"sync"
	"time"

	"github.com/ZeroHubProjects/discord-bot/internal/config"
	"github.com/ZeroHubProjects/discord-bot/internal/runners/status_updates/discord"
	"go.uber.org/zap"
)

var interval = time.Minute

func Run(cfg config.Config, wg *sync.WaitGroup) {
	defer wg.Done()

	logger := cfg.Logger.Named("status_updates")

	for {
		logger.Debugf("Running status updater with %v interval...", interval)
		runStatusUpdatesLoop(cfg.DiscordBotToken, cfg.SS13ServerAddress, cfg.Modules.StatusUpdates.DiscordChannelID, logger)
		time.Sleep(interval)
	}
}

func runStatusUpdatesLoop(botToken, ss13ServerAddress, channelID string, logger *zap.SugaredLogger) {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("status updater panicked: %v", err)
		}
	}()
	for {
		ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*5)
		err := discord.UpdateServerStatus(botToken, ss13ServerAddress, channelID, logger, ctx)
		if err != nil {
			logger.Errorf("failed to update server status: %v", err)
		}
		cancelCtx()
		time.Sleep(interval)
	}
}
