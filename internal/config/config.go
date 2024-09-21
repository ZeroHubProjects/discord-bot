package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GetConfig(fs afero.Fs) (Config, error) {
	// set up logging first so even if we fail to start we can properly report this
	loggerLevel := zap.NewAtomicLevelAt(zap.DebugLevel)
	logger := setupLogger(loggerLevel, fs)

	// configuration
	config := Config{}
	config.Logger = logger

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

	if !config.DebugLog {
		loggerLevel.SetLevel(zap.InfoLevel)
	}

	// TODO(rufus): add config validation
	return config, err
}

func setupLogger(level zap.AtomicLevel, fs afero.Fs) *zap.SugaredLogger {
	logCfg := zap.NewDevelopmentEncoderConfig()

	// file logger
	err := fs.MkdirAll("logs", 0777)
	if err != nil {
		panic(fmt.Errorf("failed to set up log folder: %w", err))
	}
	var logfile afero.File
	logfilePath := fmt.Sprintf("logs/%s-output.log", time.Now().Format(time.DateOnly))
	logfile, err = fs.OpenFile(logfilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		panic(fmt.Errorf("failed to open existing log file: %w", err))
	}
	fileEncoder := zapcore.NewJSONEncoder(logCfg)
	fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(logfile), level)

	// console logger
	logCfg.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	consoleEncoder := zapcore.NewConsoleEncoder(logCfg)
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)

	return zap.New(zapcore.NewTee(fileCore, consoleCore)).Sugar()
}
