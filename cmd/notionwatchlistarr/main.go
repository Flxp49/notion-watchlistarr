package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

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

	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	var programLevel = new(slog.LevelVar)
	Logger := slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{Level: programLevel}))
	if os.Getenv("LOG_LEVEL") == "0" {
		programLevel.Set(slog.LevelDebug)
	} else {
		programLevel.Set(slog.LevelError)
	}

	R := radarr.InitRadarrClient(os.Getenv("RADARR_KEY"), os.Getenv("RADARR_HOST"))
	S := sonarr.InitSonarrClient(os.Getenv("SONARR_KEY"), os.Getenv("SONARR_HOST"))
	N := notion.InitNotionClient(os.Getenv("NOTION_INTEGRATION_SECRET"), os.Getenv("NOTION_DB_ID"))

	// To manage Root Paths and Quality Profiles and update Notion DB with it.
	Rpid := make(map[string]string)
	Qpid := make(map[string]int)

	radarrInit := "1"
	if os.Getenv("RADARR_INIT") != "" {
		radarrInit = os.Getenv("RADARR_INIT")
	}
	radarrDefaultRootPath := os.Getenv("RADARR_DEFAULT_ROOT_PATH")
	radarrDefaultQualityProfile := os.Getenv("RADARR_DEFAULT_QUALITY_PROFILE")
	radarrDefaultMonitor := os.Getenv("RADARR_DEFAULT_MONITOR")
	sonarrInit := "1"
	if os.Getenv("SONARR_INIT") != "" {
		sonarrInit = os.Getenv("SONARR_INIT")
	}
	sonarrDefaultRootPath := os.Getenv("SONARR_DEFAULT_ROOT_PATH")
	sonarrDefaultQualityProfile := os.Getenv("SONARR_DEFAULT_QUALITY_PROFILE")
	sonarrDefaultMonitor := os.Getenv("SONARR_DEFAULT_MONITOR")

	ArrSyncinternvalSec := 10
	WatchlistSyncIntervalHour := 12
	if os.Getenv("ARRSYNC_INTERVAL_SEC") != "" {
		ArrSyncinternvalSec, _ = strconv.Atoi(os.Getenv("ARRSYNC_INTERVAL_SEC"))
	}
	if os.Getenv("WATCHLIST_SYNC_INTERVAL_HOUR") != "" {
		WatchlistSyncIntervalHour, _ = strconv.Atoi(os.Getenv("WATCHLIST_SYNC_INTERVAL_HOUR"))
	}

Start:
	if !waitForService(radarrInit, sonarrInit) {
		Logger.Error("Radarr / Sonarr servies not available, Retrying...")
		time.Sleep(time.Second * 30)
		goto Start
	}
	if radarrInit == "1" {
		err = R.RadarrDefaults(radarrDefaultRootPath, radarrDefaultQualityProfile, radarrDefaultMonitor, Rpid, Qpid)
		if err != nil {
			Logger.Error("Failed to fetch Radarr defaults, Retrying...", "Error", err)
			time.Sleep(time.Second * 30)
			goto Start
		}
	}
	if sonarrInit == "1" {
		err = S.SonarrDefaults(sonarrDefaultRootPath, sonarrDefaultQualityProfile, sonarrDefaultMonitor, Rpid, Qpid)
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
	if radarrInit == "1" {
		go routine.RadarrSync(Logger, N, R, time.Duration(ArrSyncinternvalSec))
		go routine.RadarrWatchlistSync(Logger, N, R, time.Duration(WatchlistSyncIntervalHour))
	}
	if sonarrInit == "1" {
		go routine.SonarrSync(Logger, N, S, time.Duration(ArrSyncinternvalSec))
		go routine.SonarrWatchlistSync(Logger, N, S, time.Duration(WatchlistSyncIntervalHour))
	}

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "7879"
	}

	Server := server.NewServer(PORT, N, R, S, Logger, radarrInit == "1", sonarrInit == "1")
	err = Server.Start()
	if err != nil {
		Logger.Error("Server failed to listen", "Error", err)
		os.Exit(1)
	}
}

func waitForService(radarrInit string, sonarrInit string) bool {
	status := false
	client := http.Client{}
	if radarrInit == "1" {
		resp, err := client.Get(os.Getenv("RADARR_HOST"))
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		status = resp.StatusCode == http.StatusOK
	}
	if sonarrInit == "1" {
		resp, err := client.Get(os.Getenv("SONARR_HOST"))
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		status = resp.StatusCode == http.StatusOK
	}
	return status
}
