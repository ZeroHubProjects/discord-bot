package config

type Config struct {
	DebugLog bool           `mapstructure:"debug_log"`
	Modules  ModulesConfig  `mapstructure:"modules"`
	SS13     SS13Config     `mapstructure:"ss13"`
	Discord  DiscordConfig  `mapstructure:"discord"`
	Database DatabaseConfig `mapstructure:"database"`
}

type ModulesConfig struct {
	StatusUpdatesEnabled     bool           `mapstructure:"status_updates_enabled"`
	DOOCEnabled              bool           `mapstructure:"dooc_enabled"`
	BYONDVerificationEnabled bool           `mapstructure:"byond_verification_enabled"`
	AhelpEnabled             bool           `mapstructure:"ahelp_enabled"`
	MetricsEnabled           bool           `mapstructure:"metrics_enabled"`
	Webhooks                 WebhooksConfig `mapstructure:"webhooks"`
}

type WebhooksConfig struct {
	Enabled              bool `mapstructure:"enabled"`
	Port                 int  `mapstructure:"port"`
	OOCMessagesEnabled   bool `mapstructure:"ooc_messages_enabled"`
	EmoteMessagesEnabled bool `mapstructure:"emote_messages_enabled"`
	AhelpMessagesEnabled bool `mapstructure:"ahelp_messages_enabled"`
}

type SS13Config struct {
	ServerAddress            string `mapstructure:"server_address"`
	AlternativeServerAddress string `mapstructure:"alternative_server_address"`
	AccessKey                string `mapstructure:"access_key"`
}

type DiscordConfig struct {
	BotToken                   string   `mapstructure:"bot_token"`
	OOCChannelID               string   `mapstructure:"ooc_channel_id"`
	EmoteChannelID             string   `mapstructure:"emote_channel_id"`
	StatusChannelIDs           []string `mapstructure:"status_channel_ids"`
	BYONDVerificationChannelID string   `mapstructure:"byond_verification_channel_id"`
	AhelpChannelID             string   `mapstructure:"ahelp_channel_id"`
}

type DatabaseConfig struct {
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Address      string `mapstructure:"address"`
	Port         string `mapstructure:"port"`
	DatabaseName string `mapstructure:"database_name"`
}
