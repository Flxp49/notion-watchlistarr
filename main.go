package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/flxp49/notion-watchlist-radarr-sonarr/notion"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/radarr"
	"github.com/joho/godotenv"
)

var exit = make(chan bool)

func main() {
	// init log file
	f, err := os.OpenFile("notionSyncLogFile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	// setup logging
	logger := slog.New(slog.NewTextHandler(f, nil))
	// load env (temp)
	err = godotenv.Load()
	if err != nil {
		logger.Error("Error loading .env file")
		os.Exit(1)
	}
	R := radarr.InitRadarrClient(os.Getenv("RADARRKEY"), os.Getenv("RADARRHOST"))
	N := notion.InitNotionClient("Emad", os.Getenv("NOTIONINTEG"), os.Getenv("DBID"))

	// Defaults
	radarrDefaultRootPath := "Movie: " + os.Getenv("RADARRDEFAULTROOTPATH")
	radarrDefaultQualityProfile := "Movie: " + os.Getenv("RADARRDEFAULTQUALITYPROFILE")

	radarrSucc := true

	// Root Paths
	var rps []string
	rpid := make(map[string]string)
	// Radarr root path
	radarrRootPaths, err := R.GetRootFolder()
	if len(radarrRootPaths) == 0 || err != nil {
		logger.Warn("Failed to fetch Radarr root path", err)
		radarrSucc = false
	} else {
		for _, r := range radarrRootPaths {
			rps = append(rps, "Movie: "+r.Path)
			rpid["Movie: "+r.Path] = r.Path
		}
	}
	// Quality Profiles
	var qps []string
	qpid := make(map[string]int)
	// Radarr quality profile
	radarrQualityProfiles, err := R.GetQualityProfiles()
	if len(radarrQualityProfiles) == 0 || err != nil {
		logger.Error("Failed to fetch Radarr quality profiles", err)
		radarrSucc = false
	} else {
		for _, v := range radarrQualityProfiles {
			qps = append(qps, "Movie: "+v.Name)
			qpid["Movie: "+v.Name] = v.Id
		}
		logger.Info("Quality profiles fetched")
	}

	if !radarrSucc {
		logger.Error("Failed to fetch quality profiles & rootpaths from Radarr")
	} else {
		// Add properties to the DB
		err = N.AddDBProperties(qps, rps)
		if err != nil {
			logger.Error("Failed to add properties to DB", err)
			os.Exit(1)
		}
		logger.Info("Database updated with new properties")
	}
	if radarrSucc {
		go rdrr(R, N, qpid, rpid, radarrDefaultRootPath, radarrDefaultQualityProfile, logger)
	}
	<-exit
	logger.Error("Shutting down due to termination of Radarr subroutine")
	os.Exit(1)
}

func rdrr(R *radarr.RadarrClient, N *notion.NotionClient, qpid map[string]int, rpid map[string]string, defaultRootPath string, defaultQualityProfile string, logger *slog.Logger) {

	for {
		logger.Info("Radarr: Fetching titles")
		data, err := N.QueryDB("Movie")
		if err != nil {
			logger.Error("Radarr: ", "Failed to query watchlist DB", err)
		}
		logger.Info(fmt.Sprintf("Radarr: Fetched titles from DB: %d", len(data.Results)))
		for _, v := range data.Results {
			if v.Properties.RootFolder.Select.Name == "" {
				v.Properties.RootFolder.Select.Name = defaultRootPath
			}
			if v.Properties.QualityProfile.Select.Name == "" {
				v.Properties.QualityProfile.Select.Name = defaultQualityProfile
			}
			logger.Info(fmt.Sprintf("Radarr: Adding Title: %s", v.Properties.Name.Title[0].Plain_text))
			err = R.AddMovie(v.Properties.Name.Title[0].Plain_text, qpid[v.Properties.QualityProfile.Select.Name], v.Properties.Tmdbid.Number, rpid[v.Properties.RootFolder.Select.Name], true, true)
			if err != nil {
				logger.Error(fmt.Sprintf("Radarr: Error adding title: %s", v.Properties.Name.Title[0].Plain_text), err)
				N.UpdateDownloadStatus(v.Pgid, "Error", "", "")
				continue
			}
			N.UpdateDownloadStatus(v.Pgid, "Queued", v.Properties.QualityProfile.Select.Name, v.Properties.RootFolder.Select.Name)
			logger.Info(fmt.Sprintf("Radarr: Added title: %s", v.Properties.Name.Title[0].Plain_text))
		}
		time.Sleep(5 * time.Second)
	}
}
