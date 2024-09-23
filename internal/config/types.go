package config

type Config struct {
	DebugLog bool          `mapstructure:"debug_log"`
	Modules  ModulesConfig `mapstructure:"modules"`
	SS13     SS13Config    `mapstructure:"ss13"`
	Discord  DiscordConfig `mapstructure:"discord"`
}

type ModulesConfig struct {
	StatusUpdatesEnabled bool           `mapstructure:"status_updates_enabled"`
	Webhooks             WebhooksConfig `mapstructure:"webhooks"`
	DOOCEnabled          bool           `mapstructure:"dooc_enabled"`
}

type WebhooksConfig struct {
	Enabled            bool `mapstructure:"enabled"`
	Port               int  `mapstructure:"port"`
	OOCMessagesEnabled bool `mapstructure:"ooc_messages_enabled"`
}

type SS13Config struct {
	ServerAddress string `mapstructure:"server_address"`
	AccessKey     string `mapstructure:"access_key"`
}

type DiscordConfig struct {
	BotToken        string `mapstructure:"bot_token"`
	OOCChannelID    string `mapstructure:"ooc_channel_id"`
	StatusChannelID string `mapstructure:"status_channel_id"`
}
