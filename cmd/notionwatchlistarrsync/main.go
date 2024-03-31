package main

import (
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/flxp49/notion-watchlist-radarr-sonarr/internal/notion"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/internal/radarr"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/internal/routine"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/internal/sonarr"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/server"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// init log file
	f, err := os.OpenFile("notionwatchlistarrsync.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	var programLevel = new(slog.LevelVar)
	Logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel}))
	if os.Getenv("LOG_LEVEL") != "" || os.Getenv("LOG_LEVEL") == "0" {
		programLevel.Set(slog.LevelDebug)
	} else {
		programLevel.Set(slog.LevelError)
	}

	R := radarr.InitRadarrClient(os.Getenv("RADARR_KEY"), os.Getenv("RADARR_HOST"))
	S := sonarr.InitSonarrClient(os.Getenv("SONARR_KEY"), os.Getenv("SONARR_HOST"))
	N := notion.InitNotionClient(os.Getenv("NOTION_USER"), os.Getenv("NOTION_INTEGRATION_SECRET"), os.Getenv("NOTION_DB_ID"))

	// To manage Root Paths and Quality Profiles and update Notion DB with it.
	Rpid := make(map[string]string)
	Qpid := make(map[string]int)

	// set flag for radarr routine
	radarrStart := true
	// set flag for sonarr routine
	sonarrStart := true

	// Fetch Radarr & Sonarr info
	radarrDefaultRootPath := ""
	radarrDefaultQualityProfile := ""
	radarrDefaultMonitor := ""
	if os.Getenv("RADARR_DEFAULT_ROOT_PATH") != "" {
		radarrDefaultRootPath = os.Getenv("RADARR_DEFAULT_ROOT_PATH")
	}
	if os.Getenv("RADARR_DEFAULT_QUALITY_PROFILE") != "" {
		radarrDefaultQualityProfile = os.Getenv("RADARR_DEFAULT_QUALITY_PROFILE")
	}
	if os.Getenv("RADARR_DEFAULT_MONITOR") != "" {
		radarrDefaultMonitor = os.Getenv("RADARR_DEFAULT_MONITOR")
	}
	sonarrDefaultRootPath := ""
	sonarrDefaultQualityProfile := ""
	sonarrDefaultMonitor := ""
	if os.Getenv("SONARR_DEFAULT_ROOT_PATH") != "" {
		sonarrDefaultRootPath = os.Getenv("SONARR_DEFAULT_ROOT_PATH")
	}
	if os.Getenv("SONARR_DEFAULT_QUALITY_PROFILE") != "" {
		sonarrDefaultQualityProfile = os.Getenv("SONARR_DEFAULT_QUALITY_PROFILE")
	}
	if os.Getenv("SONARR_DEFAULT_MONITOR") != "" {
		sonarrDefaultMonitor = os.Getenv("SONARR_DEFAULT_MONITOR")
	}
	ArrSyncinternvalSec := 10
	WatchlistSyncIntervalHour := 12
	if os.Getenv("ARRSYNC_INTERVAL_SEC") != "" {
		ArrSyncinternvalSec, _ = strconv.Atoi(os.Getenv("ARRSYNC_INTERVAL_SEC"))
	}
	if os.Getenv("WATCHLIST_SYNC_INTERVAL_HOUR") != "" {
		WatchlistSyncIntervalHour, _ = strconv.Atoi(os.Getenv("WATCHLIST_SYNC_INTERVAL_HOUR"))
	}

	err = R.RadarrDefaults(radarrDefaultRootPath, radarrDefaultQualityProfile, radarrDefaultMonitor, Rpid, Qpid)
	if err != nil {
		Logger.Error("Failed to fetch Sonarr defaults, Radarr routine not initialized", "Error", err)
		radarrStart = false
	}
	err = S.SonarrDefaults(sonarrDefaultRootPath, sonarrDefaultQualityProfile, sonarrDefaultMonitor, Rpid, Qpid)
	if err != nil {
		Logger.Error("Failed to fetch Sonarr defaults, Sonarr routine not initialized", "Error", err)
		sonarrStart = false
	}
	if radarrStart || sonarrStart {
		// Add properties to the DB
		err = N.AddDBProperties(Qpid, Rpid)
		if err != nil {
			Logger.Error("Failed to add properties to DB", "Error", err)
			os.Exit(1)
		}
		Logger.Info("Database updated with new properties")
		if radarrStart {
			go routine.RadarrSync(Logger, N, R, time.Duration(ArrSyncinternvalSec))
			go routine.RadarrWatchlistSync(Logger, N, R, time.Duration(WatchlistSyncIntervalHour))
		}
		if sonarrStart {
			go routine.SonarrSync(Logger, N, S, time.Duration(ArrSyncinternvalSec))
			go routine.SonarrWatchlistSync(Logger, N, S, time.Duration(WatchlistSyncIntervalHour))
		}
	} else {
		Logger.Error("Failed to start radarr and sonarr, terminating app")
		os.Exit(1)
	}

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "7879"
	}

	Server := server.NewServer(PORT, N, R, S, Logger)
	err = Server.Start()
	if err != nil {
		Logger.Error("Server failed to listen", "Error", err)
		os.Exit(1)
	}
}
