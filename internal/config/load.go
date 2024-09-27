package config

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

func GetConfig(fs afero.Fs) (Config, error) {
	// configuration
	config := Config{}

	viper.SetEnvPrefix("VIPER")
	viper.MustBindEnv("CONFIG")
	configPath := viper.GetString("CONFIG")
	viper.SetConfigFile(configPath)

	viper.SetFs(fs)

	err := viper.ReadInConfig()
	if err != nil {
		return config, fmt.Errorf("failed to read config: %w", err)
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		return config, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// TODO(rufus): add config validation
	return config, err
}
