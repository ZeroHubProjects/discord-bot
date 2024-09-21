package main

import (
	"sync"

	"github.com/ZeroHubProjects/discord-bot/internal/config"
	statusupdates "github.com/ZeroHubProjects/discord-bot/internal/runners/status_updates"
	server "github.com/ZeroHubProjects/discord-bot/internal/webhooks_server"
	"github.com/joho/godotenv"
	"github.com/spf13/afero"
)

func main() {
	godotenv.Load()

	// load config
	cfg, err := config.GetConfig(afero.NewOsFs())
	if err != nil {
		cfg.Logger.Fatalf("failed to get config: %v, check if you configured config.yaml?", err)
	}
	defer cfg.Logger.Sync()

	wg := new(sync.WaitGroup)
	// status updater module
	if cfg.Modules.StatusUpdates.Enabled {
		wg.Add(1)
		go statusupdates.Run(cfg, wg)
	}
	// webhooks server module
	if cfg.Modules.Webhooks.Enabled {
		wg.Add(1)
		go server.Run(cfg, wg)
	}

	wg.Wait()
	cfg.Logger.Info("all modules done working, exiting")
}
