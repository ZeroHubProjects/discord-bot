package main

import (
	"sync"

	"github.com/ZeroHubProjects/discord-bot/internal/config"
	"github.com/ZeroHubProjects/discord-bot/internal/discord/dooc"
	"github.com/ZeroHubProjects/discord-bot/internal/discord/relay"
	"github.com/ZeroHubProjects/discord-bot/internal/discord/verification"
	"github.com/ZeroHubProjects/discord-bot/internal/status"
	"github.com/ZeroHubProjects/discord-bot/internal/types"
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
		statusUpdater := status.StatusUpdater{
			Discord:           dg,
			SS13ServerAddress: cfg.SS13.ServerAddress,
			StatusChannelID:   cfg.Discord.StatusChannelID,
			Logger:            logger.Named("status_updates"),
		}
		go statusUpdater.Run(wg)
	}
	// webhooks server module
	if cfg.Modules.Webhooks.Enabled {
		wg.Add(1)
		server := webhooks.WebhookServer{
			Port:               cfg.Modules.Webhooks.Port,
			SS13AccessKey:      cfg.SS13.AccessKey,
			OOCMessagesEnabled: cfg.Modules.Webhooks.OOCMessagesEnabled,
			Logger:             logger.Named("webhooks"),
		}
		// OOC to discord relay
		if cfg.Modules.Webhooks.OOCMessagesEnabled {
			wg.Add(1)
			server.OOCMessageQueue = make(chan types.OOCMessage, 5)
			relay := relay.OOCRelay{
				Queue:     server.OOCMessageQueue,
				ChannelID: cfg.Discord.OOCChannelID,
				Discord:   dg,
				Logger:    logger.Named("relay.ooc"),
			}
			go relay.Run(wg)
		}
		go server.Run(wg)

	}
	// discord ooc channel processing
	if cfg.Modules.DOOCEnabled {
		wg.Add(1)
		handler := dooc.DOOCHandler{
			SS13ServerAddress:   cfg.SS13.ServerAddress,
			SS13AccessKey:       cfg.SS13.AccessKey,
			OOCChannelID:        cfg.Discord.OOCChannelID,
			Discord:             dg,
			Logger:              logger.Named("dooc"),
			VerificationHandler: nil, // disabled until verification module is complete
		}
		go handler.Run(wg)
	}
	// verification processing
	if cfg.Modules.BYONDVerificationEnabled {
		wg.Add(1)
		handler := verification.ByondVerificationHandler{
			Discord:   dg,
			ChannelID: cfg.Discord.BYONDVerificationChannelID,
			Logger:    logger.Named("verification"),
		}
		go handler.Run(wg)
	}

	dg.Open()
	defer dg.Close()

	wg.Wait()
	logger.Info("all modules done working, exiting")
}
