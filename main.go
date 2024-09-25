package main

import (
	"sync"

	"github.com/ZeroHubProjects/discord-bot/internal/config"
	"github.com/ZeroHubProjects/discord-bot/internal/discord"
	"github.com/ZeroHubProjects/discord-bot/internal/discord/dooc"
	"github.com/ZeroHubProjects/discord-bot/internal/status"
	"github.com/ZeroHubProjects/discord-bot/internal/webhooks"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

func main() {
	godotenv.Load()

	// setup logger
	loggerLevel := zap.NewAtomicLevelAt(zap.DebugLevel)
	logger := config.SetupLogger(loggerLevel, afero.NewOsFs())
	defer logger.Sync()

	// load config
	cfg, err := config.GetConfig(afero.NewOsFs())
	if err != nil {
		logger.Fatalf("failed to get config: %v, check if you configured config.yaml?", err)
	}

	// adjust log level
	if !cfg.DebugLog {
		loggerLevel.SetLevel(zap.InfoLevel)
	}

	// create a discord session
	dg, err := discordgo.New("Bot " + cfg.Discord.BotToken)
	if err != nil {
		logger.Fatalf("can't set up discord session: %v", err)
	}

	wg := new(sync.WaitGroup)
	// status updater module
	if cfg.Modules.StatusUpdatesEnabled {
		wg.Add(1)
		go status.Run(cfg.SS13.ServerAddress, cfg.Discord.StatusChannelID, dg, logger.Named("status_updates"), wg)
	}
	// webhooks server module
	if cfg.Modules.Webhooks.Enabled {
		wg.Add(1)
		go webhooks.Run(cfg.SS13.AccessKey, cfg.Modules.Webhooks, logger.Named("webhooks"), wg)
		if cfg.Modules.Webhooks.OOCMessagesEnabled {
			wg.Add(1)
			go discord.RunOOCProcessingLoop(cfg.Discord.OOCChannelID, dg, logger.Named("webhooks.ooc.processing"), wg)
		}
	}
	// discord ooc channel processing
	if cfg.Modules.DOOCEnabled {
		wg.Add(1)
		go dooc.RunDOOC(cfg.SS13, cfg.Discord, dg, logger.Named("discord.dooc"), wg)
	}

	dg.Open()
	defer dg.Close()

	wg.Wait()
	logger.Info("all modules done working, exiting")
}
