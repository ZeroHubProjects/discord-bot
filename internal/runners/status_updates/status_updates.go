package statusupdates

import (
	"context"
	"sync"
	"time"

	"github.com/ZeroHubProjects/discord-bot/internal/runners/status_updates/discord"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

var interval = time.Minute

func Run(ss13ServerAddress, statusChannelID string, dg *discordgo.Session, logger *zap.SugaredLogger, wg *sync.WaitGroup) {
	defer wg.Done()

	statusUpdater := discord.StatusUpdater{
		Discord:           dg,
		SS13ServerAddress: ss13ServerAddress,
		StatusChannelID:   statusChannelID,
	}

	for {
		logger.Debugf("Running status updater with %v interval...", interval)
		runStatusUpdatesLoop(&statusUpdater, logger)
		time.Sleep(interval)
	}
}

func runStatusUpdatesLoop(statusUpdater *discord.StatusUpdater, logger *zap.SugaredLogger) {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("status updater panicked: %v", err)
		}
	}()

	for {
		ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second*5)
		err := statusUpdater.UpdateServerStatus(ctx)
		if err != nil {
			logger.Errorf("failed to update server status: %v", err)
		}
		cancelCtx()
		time.Sleep(interval)
	}
}
