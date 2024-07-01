package config

import (
	"github.com/caarlos0/env/v11"
)

type config struct {
	Port                        string `env:"PORT" envDefault:"7879"`
	RadarrHost                  string `env:"RADARR_HOST"`
	RadarrKey                   string `env:"RADARR_KEY"`
	RadarrInit                  bool   `env:"RADARR_INIT" envDefault:"true"`
	RadarrDefaultRootPath       string `env:"RADARR_DEFAULT_ROOT_PATH"`
	RadarrDefaultQualityProfile string `env:"RADARR_DEFAULT_QUALITY_PROFILE"`
	RadarrDefaultMonitor        string `env:"RADARR_DEFAULT_MONITOR"`
	SonarrHost                  string `env:"SONARR_HOST"`
	SonarrKey                   string `env:"SONARR_KEY"`
	SonarrInit                  bool   `env:"SONARR_INIT" envDefault:"true"`
	SonarrDefaultRootPath       string `env:"SONARR_DEFAULT_ROOT_PATH"`
	SonarrDefaultQualityProfile string `env:"SONARR_DEFAULT_QUALITY_PROFILE"`
	SonarrDefaultMonitor        string `env:"SONARR_DEFAULT_MONITOR"`
	NotionSecret                string `env:"NOTION_INTEGRATION_SECRET,notEmpty"`
	NotionDBID                  string `env:"NOTION_DB_ID,notEmpty"`
	ArrSyncinternvalSec         int    `env:"ARRSYNC_INTERVAL_SEC" envDefault:"10"`
	WatchlistSyncIntervalHr     int    `env:"WATCHLIST_SYNC_INTERVAL_HOUR" envDefault:"12"`
	LogDebug                    bool   `env:"LOG_DEBUG" envDefault:"false"`
}

func LoadConfig() (config, error) {
	cfg := config{}
	err := env.Parse(&cfg)
	if err != nil {
		return config{}, err
	}
	return cfg, nil
}
