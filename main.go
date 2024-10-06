package main

import (
	"sync"

	"github.com/ZeroHubProjects/discord-bot/internal/config"
	"github.com/ZeroHubProjects/discord-bot/internal/database"
	"github.com/ZeroHubProjects/discord-bot/internal/discord/ahelp"
	"github.com/ZeroHubProjects/discord-bot/internal/discord/dooc"
	"github.com/ZeroHubProjects/discord-bot/internal/discord/relay"
	"github.com/ZeroHubProjects/discord-bot/internal/discord/verification"
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
	// create a database connection
	db, dberr := database.NewDatabase(cfg.Database, logger.Named("database"))
	if dberr == nil {
		defer db.Close()
	}

	wg := new(sync.WaitGroup)
	// status updater module
	if cfg.Modules.StatusUpdatesEnabled {
		statusUpdater := status.StatusUpdater{
			Discord:           dg,
			SS13ServerAddress: cfg.SS13.ServerAddress,
			StatusChannelID:   cfg.Discord.StatusChannelID,
			Logger:            logger.Named("status_updates"),
		}
		wg.Add(1)
		go statusUpdater.Run(wg)
	}
	// webhooks server module
	if cfg.Modules.Webhooks.Enabled {
		server := webhooks.WebhookServer{
			Port:                 cfg.Modules.Webhooks.Port,
			SS13AccessKey:        cfg.SS13.AccessKey,
			OOCMessagesEnabled:   cfg.Modules.Webhooks.OOCMessagesEnabled,
			AhelpMessagesEnabled: cfg.Modules.Webhooks.AhelpMessagesEnabled,
			Logger:               logger.Named("webhooks"),
		}
		// OOC to discord relay
		if cfg.Modules.Webhooks.OOCMessagesEnabled {
			server.OOCMessageQueue = make(chan webhooks.OOCMessage, 5)
			relay := relay.OOCRelay{
				Queue:     server.OOCMessageQueue,
				ChannelID: cfg.Discord.OOCChannelID,
				Discord:   dg,
				Logger:    logger.Named("relay.ooc"),
			}
			wg.Add(1)
			go relay.Run(wg)
		}
		// Ahelp to discord relay
		if cfg.Modules.Webhooks.AhelpMessagesEnabled {
			server.AhelpMessageQueue = make(chan webhooks.AhelpMessage, 5)
			relay := relay.AhelpRelay{
				Queue:     server.AhelpMessageQueue,
				ChannelID: cfg.Discord.AhelpChannelID,
				Discord:   dg,
				Logger:    logger.Named("relay.ahelp"),
			}
			wg.Add(1)
			go relay.Run(wg)
		}
		wg.Add(1)
		go server.Run(wg)
	}
	// verification processing
	var verificationHandler *verification.ByondVerificationHandler // keep a reference so other modules can use it
	if cfg.Modules.BYONDVerificationEnabled {
		if dberr != nil {
			logger.Fatalf("failed to establish a required database connection for the verification module: %v", dberr)
		}
		verificationHandler = &verification.ByondVerificationHandler{
			Discord:   dg,
			ChannelID: cfg.Discord.BYONDVerificationChannelID,
			Logger:    logger.Named("verification"),
			Database:  db,
		}
		wg.Add(1)
		go verificationHandler.Run(wg)
	}
	// discord ooc channel processing
	if cfg.Modules.DOOCEnabled {
		handler := dooc.DOOCHandler{
			SS13ServerAddress:   cfg.SS13.ServerAddress,
			SS13AccessKey:       cfg.SS13.AccessKey,
			OOCChannelID:        cfg.Discord.OOCChannelID,
			Discord:             dg,
			Logger:              logger.Named("dooc"),
			VerificationHandler: verificationHandler, // might be nil if verification is not enabled in the config
		}
		wg.Add(1)
		go handler.Run(wg)
	}
	// ahelp channel processing, verification module is required
	if cfg.Modules.AhelpEnabled && verificationHandler != nil {
		handler := ahelp.AhelpHandler{
			SS13ServerAddress:   cfg.SS13.ServerAddress,
			SS13AccessKey:       cfg.SS13.AccessKey,
			AhelpChannelID:      cfg.Discord.AhelpChannelID,
			Discord:             dg,
			Logger:              logger.Named("ahelp"),
			VerificationHandler: verificationHandler,
		}
		wg.Add(1)
		go handler.Run(wg)
	}

	dg.Open()
	defer dg.Close()

	wg.Wait()
	logger.Info("all modules done working, exiting")
}
