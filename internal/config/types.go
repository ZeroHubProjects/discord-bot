package config

type Config struct {
	SS13ServerAddress string        `mapstructure:"ss13_server_address"`
	DebugLog          bool          `mapstructure:"debug_log"`
	Modules           ModulesConfig `mapstructure:"modules"`
	Discord           DiscordConfig `mapstructure:"discord"`
}

type ModulesConfig struct {
	StatusUpdatesEnabled bool           `mapstructure:"status_updates_enabled"`
	Webhooks             WebhooksConfig `mapstructure:"webhooks"`
}

type WebhooksConfig struct {
	Enabled            bool   `mapstructure:"enabled"`
	Port               int    `mapstructure:"port"`
	AccessKey          string `mapstructure:"access_key"`
	OOCMessagesEnabled bool   `mapstructure:"ooc_messages_enabled"`
}

type DiscordConfig struct {
	BotToken        string `mapstructure:"bot_token"`
	OOCChannelID    string `mapstructure:"ooc_channel_id"`
	StatusChannelID string `mapstructure:"status_channel_id"`
}
