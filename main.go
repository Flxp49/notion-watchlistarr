package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/flxp49/notion-watchlist-radarr-sonarr/api"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/notion"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/radarr"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/routine"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/sonarr"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// init log file
	f, err := os.OpenFile("notionRadarrSonarrLogFile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	var programLevel = new(slog.LevelVar)
	Logger := slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{Level: programLevel}))
	if os.Getenv("LOG_LEVEL") != "" && os.Getenv("LOG_LEVEL") == "1" {
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
	radarrDefaultMonitor := ""
	radarrDefaultQualityProfile := ""
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
	sonarrDefaultMonitor := ""
	sonarrDefaultQualityProfile := ""
	if os.Getenv("SONARR_DEFAULT_ROOT_PATH") != "" {
		radarrDefaultRootPath = os.Getenv("SONARR_DEFAULT_ROOT_PATH")
	}
	if os.Getenv("SONARR_DEFAULT_QUALITY_PROFILE") != "" {
		radarrDefaultQualityProfile = os.Getenv("SONARR_DEFAULT_QUALITY_PROFILE")
	}
	if os.Getenv("SONARR_DEFAULT_MONITOR") != "" {
		radarrDefaultMonitor = os.Getenv("SONARR_DEFAULT_MONITOR")
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
			go routine.RadarrSync(Logger, N, R)
		}
		if sonarrStart {
			go routine.SonarrSync(Logger, N, S)
		}

		// todo: init watchlist sync
	}

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "7879"
	}

	Server := api.NewServer(PORT, N, R, S, Logger)
	err = Server.Start()
	if err != nil {
		Logger.Error(fmt.Sprintf("Failed to listen on PORT %s", PORT), "error", err)
		os.Exit(1)
	}
}
