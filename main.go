package main

import (
	"sync"

	"github.com/ZeroHubProjects/discord-bot/internal/config"
	"github.com/ZeroHubProjects/discord-bot/internal/discord"
	statusupdates "github.com/ZeroHubProjects/discord-bot/internal/runners/status_updates"
	server "github.com/ZeroHubProjects/discord-bot/internal/webhooks_server"
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
		go statusupdates.Run(cfg.SS13.ServerAddress, cfg.Discord.StatusChannelID, dg, logger, wg)
	}
	// webhooks server module
	if cfg.Modules.Webhooks.Enabled {
		wg.Add(1)
		go server.Run(cfg.SS13.AccessKey, cfg.Modules.Webhooks, logger, wg)
	}
	// discord processing
	// NOTE(rufus): this currently doesn't accept WaitGroup because it doesn't do anything on its own
	go discord.Run(cfg.Discord, dg, logger)

	wg.Wait()
	logger.Info("all modules done working, exiting")
}
