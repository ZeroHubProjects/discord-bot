package config

import "go.uber.org/zap"

type Config struct {
	DiscordBotToken   string  `mapstructure:"discord_bot_token"`
	SS13ServerAddress string  `mapstructure:"ss13_server_address"`
	Modules           Modules `mapstructure:"modules"`
	DebugLog          bool    `mapstructure:"debug_log"`
	// so logger can be propagated together with the config
	Logger *zap.SugaredLogger `mapstructure:"-"`
}

type Modules struct {
	StatusUpdates StatusUpdates `mapstructure:"status_updates"`
	Webhooks      Webhooks      `mapstructure:"webhooks"`
}

type StatusUpdates struct {
	Enabled          bool   `mapstructure:"enabled"`
	DiscordChannelID string `mapstructure:"discord_channel_id"`
}

type Webhooks struct {
	Enabled   bool      `mapstructure:"enabled"`
	Port      int       `mapstructure:"port"`
	AccessKey string    `mapstructure:"access_key"`
	OOC       OOCBridge `mapstructure:"ooc_bridge"`
}

type OOCBridge struct {
	Enabled          bool   `mapstructure:"enabled"`
	DiscordChannelID string `mapstructure:"discord_channel_id"`
}
