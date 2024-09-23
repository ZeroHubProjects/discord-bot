package statusupdates

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

var interval = time.Minute

func Run(ss13ServerAddress, statusChannelID string, dg *discordgo.Session, logger *zap.SugaredLogger, wg *sync.WaitGroup) {
	defer wg.Done()

	statusUpdater := statusUpdater{
		discord:           dg,
		ss13ServerAddress: ss13ServerAddress,
		statusChannelID:   statusChannelID,
		logger:            logger,
	}

	for {
		logger.Debugf("updating with %v interval...", interval)
		runStatusUpdatesLoop(&statusUpdater, logger)
		time.Sleep(interval)
	}
}

func runStatusUpdatesLoop(statusUpdater *statusUpdater, logger *zap.SugaredLogger) {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("status updater panicked: %v", err)
		}
	}()

	for {
		err := statusUpdater.updateServerStatus()
		if err != nil {
			logger.Errorf("failed to update server status: %v", err)
		}
		time.Sleep(interval)
	}
}
