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
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// init log file
	f, err := os.OpenFile("notionSyncLogFile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	Logger := slog.New(slog.NewTextHandler(f, nil))

	R := radarr.InitRadarrClient(os.Getenv("RADARR_KEY"), os.Getenv("RADARR_HOST"))
	N := notion.InitNotionClient(os.Getenv("NOTION_USER"), os.Getenv("NOTION_INTEGRATION_SECRET"), os.Getenv("NOTION_DB_ID"))

	// To manage Root Paths and Quality Profiles and update Notion DB with it.
	Rpid := make(map[string]string)
	Qpid := make(map[string]int)
	// set flag for radarr routine
	radarrStart := true

	// Fetch Radarr & Sonarr info
	radarrDefaultRootPath := ""
	radarrDefaultQualityProfile := ""
	if os.Getenv("RADARR_DEFAULT_ROOT_PATH") != "" {
		radarrDefaultRootPath = "Movie: " + os.Getenv("RADARR_DEFAULT_ROOT_PATH")
	}
	if os.Getenv("RADARR_DEFAULT_QUALITY_PROFILE") != "" {
		radarrDefaultQualityProfile = "Movie: " + os.Getenv("RADARR_DEFAULT_QUALITY_PROFILE")
	}

	err = R.GetRadarrDefaults(radarrDefaultRootPath, radarrDefaultQualityProfile, Rpid, Qpid)
	if err != nil {
		Logger.Error("Error fetching Radarr defaults, Radarr routine not initialized", "Error", err)
		radarrStart = false
	}
	//todo: same as the above for sonarr
	if radarrStart { // || sonarrStart
		// Add properties to the DB
		err = N.AddDBProperties(Qpid, Rpid)
		if err != nil {
			Logger.Error("Failed to add properties to DB", "Error", err)
			os.Exit(1)
		}
		Logger.Info("Database updated with new properties")
		go routine.RadarrSync(Logger, N, R)
	}

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "7879"
	}

	Server := api.NewServer(PORT, N, R, Logger)
	err = Server.Start()
	if err != nil {
		Logger.Error(fmt.Sprintf("Failed to listen on PORT %s", PORT), "error", err)
		os.Exit(1)
	}
}
