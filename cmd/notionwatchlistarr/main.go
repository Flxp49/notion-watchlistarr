package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/flxp49/notion-watchlistarr/internal/config"
	"github.com/flxp49/notion-watchlistarr/internal/notion"
	"github.com/flxp49/notion-watchlistarr/internal/radarr"
	"github.com/flxp49/notion-watchlistarr/internal/routine"
	"github.com/flxp49/notion-watchlistarr/internal/sonarr"
	"github.com/flxp49/notion-watchlistarr/server"
	"github.com/joho/godotenv"
)

func main() {
	// init log file
	f, err := os.OpenFile("notionwatchlistarrsync.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	var programLevel = new(slog.LevelVar)
	Logger := slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{Level: programLevel}))
	err = godotenv.Load()
	if err != nil {
		Logger.Error("Error loading .env file", "Error", err)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		Logger.Error("Error loading env variables", "Error", err)
		os.Exit(1)
	}
	fmt.Printf("%+v\n", cfg)

	if cfg.LogDebug {
		programLevel.Set(slog.LevelDebug)
	}

	R := radarr.InitRadarrClient(cfg.RadarrKey, cfg.RadarrHost)
	S := sonarr.InitSonarrClient(cfg.SonarrKey, cfg.SonarrHost)
	N := notion.InitNotionClient(cfg.NotionSecret, cfg.NotionDBID)

	// To manage Root Paths and Quality Profiles and update Notion DB with it.
	Rpid := make(map[string]string)
	Qpid := make(map[string]int)

Start:
	if !waitForService(cfg.RadarrInit, cfg.RadarrHost, cfg.SonarrInit, cfg.SonarrHost) {
		Logger.Error("Radarr / Sonarr servies not available, Retrying...")
		time.Sleep(time.Second * 30)
		goto Start
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
	if cfg.RadarrInit {
		go routine.RadarrSync(Logger, N, R, time.Duration(cfg.ArrSyncinternvalSec))
		go routine.RadarrWatchlistSync(Logger, N, R, time.Duration(cfg.WatchlistSyncIntervalHr))
	}
	if cfg.SonarrInit {
		go routine.SonarrSync(Logger, N, S, time.Duration(cfg.ArrSyncinternvalSec))
		go routine.SonarrWatchlistSync(Logger, N, S, time.Duration(cfg.WatchlistSyncIntervalHr))
	}

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
