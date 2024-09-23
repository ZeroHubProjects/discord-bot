package discord

import (
	"github.com/ZeroHubProjects/discord-bot/internal/config"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func Run(cfg config.DiscordConfig, dg *discordgo.Session, logger *zap.SugaredLogger) {
	logger = logger.Named("discord")

	go runOOCProcessingLoop(cfg.OOCChannelID, dg, logger.Named("ooc"))
}
