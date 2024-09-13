package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/flxp49/notion-watchlistarr/internal/app"
	"github.com/flxp49/notion-watchlistarr/internal/config"
	"github.com/flxp49/notion-watchlistarr/internal/notion"
	"github.com/flxp49/notion-watchlistarr/internal/radarr"
	"github.com/flxp49/notion-watchlistarr/internal/sonarr"
	"github.com/flxp49/notion-watchlistarr/server"
	"github.com/joho/godotenv"
)

func main() {
	var programLevel = new(slog.LevelVar)
	Logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel}))
	err := godotenv.Load()
	if err != nil {
		Logger.Error("Error loading .env file", "Error", err)
		os.Exit(1)
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		Logger.Error("Error loading env variables", "Error", err)
		os.Exit(1)
	}

	if cfg.LogDebug {
		programLevel.Set(slog.LevelDebug)
	} else {
		programLevel.Set(slog.LevelWarn)
	}

	if !(cfg.RadarrInit || cfg.SonarrInit) {
		Logger.Error("Both Radarr and Sonarr cannot be disabled")
		os.Exit(1)
	}

	R := radarr.InitRadarrClient(cfg.RadarrKey, cfg.RadarrHost)
	S := sonarr.InitSonarrClient(cfg.SonarrKey, cfg.SonarrHost)
	N := notion.InitNotionClient(cfg.NotionSecret, cfg.NotionDBID)

	// To manage Root Paths and Quality Profiles and update Notion DB with it.
	Rpid := make(map[string]string)
	Qpid := make(map[string]int)

Start:
	if !waitForService(cfg.RadarrInit, cfg.RadarrHost, cfg.SonarrInit, cfg.SonarrHost) {
		Logger.Error("Radarr / Sonarr services not available, Retrying...")
		time.Sleep(time.Second * 30)
	}
	if cfg.RadarrInit {
		err = R.RadarrDefaults(cfg.RadarrDefaultRootPath, cfg.RadarrDefaultQualityProfile, cfg.RadarrDefaultMonitor, Rpid, Qpid)
		if err != nil {
			Logger.Error("Failed to fetch Radarr defaults, Retrying...", "Error", err)
			time.Sleep(time.Second * 30)
			goto Start
		}
	}
	if cfg.SonarrInit {
		err = S.SonarrDefaults(cfg.SonarrDefaultRootPath, cfg.SonarrDefaultQualityProfile, cfg.SonarrDefaultMonitor, Rpid, Qpid)
		if err != nil {
			Logger.Error("Failed to fetch Sonarr defaults, Retrying...", "Error", err)
			time.Sleep(time.Second * 30)
			goto Start
		}
	}
	// Add properties to the DB
	err = N.AddDBProperties(Qpid, Rpid)
	if err != nil {
		Logger.Error("Failed to add properties to DB", "Error", err)
		goto Start
	}
	Logger.Info("Database updated with new properties")

	app := app.NewApp(N, R, S, Logger, time.Duration(cfg.PollInternvalSec), time.Duration(cfg.WatchlistSyncIntervalHr), cfg.RadarrInit, cfg.SonarrInit)
	app.RunApp()

	Server := server.NewServer(cfg.Port, N, R, S, Logger, cfg.RadarrInit, cfg.SonarrInit)
	err = Server.Start()
	if err != nil {
		Logger.Error("Server failed to listen", "Error", err)
		os.Exit(1)
	}
}

func waitForService(radarrInit bool, radarrHost string, sonarrInit bool, sonarrHost string) bool {
	status := false
	client := http.Client{}
	if radarrInit {
		resp, err := client.Get(radarrHost)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		status = resp.StatusCode == http.StatusOK
	}
	if sonarrInit {
		resp, err := client.Get(sonarrHost)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		status = resp.StatusCode == http.StatusOK
	}
	return status
}
