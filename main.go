package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/flxp49/notion-watchlist-radarr-sonarr/notion"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/radarr"
	"github.com/joho/godotenv"
)

var logger *slog.Logger
var N *notion.NotionClient
var R *radarr.RadarrClient

// eventType: "MovieAdded"
type MovieInfo struct {
	Movie struct {
		id     int
		tmdbId int
	} `json:"movie"`
	EventType string `json:"eventType"`
}

func radarrHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// body, _ := io.ReadAll(r.Body)
		// var data MovieInfo
		// err := parseJson(body, &data)
		// logger.Error("Error reading body", "body", body, "error", err)

	} else {
		w.WriteHeader(405)
	}
}

func main() {
	// init log file
	f, err := os.OpenFile("notionSyncLogFile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	logger = slog.New(slog.NewTextHandler(f, nil))

	err = godotenv.Load()
	if err != nil {
		logger.Error("Error loading .env file")
		os.Exit(1)
	}

	// R = radarr.InitRadarrClient(os.Getenv("RADARR_KEY"), os.Getenv("RADARR_HOST") )
	// N = notion.InitNotionClient("Emad", os.Getenv("NOTION_INTEGRATION_SECRET"), os.Getenv("NOTION_DB_ID"))

	// To manage Root Paths and Quality Profiles and update Notion DB with it.
	rpid := make(map[string]string)
	qpid := make(map[string]int)
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

	err = getRadarrInfo(qpid, rpid, radarrDefaultRootPath, radarrDefaultQualityProfile)
	if err != nil {
		logger.Error("Error fetching Radarr details, Radarr routine not initialized", "Error", err)
		radarrStart = false
	}

	//todo: same as the above for sonarr

	// if radarrStart {
	// Add properties to the DB
	// err = N.AddDBProperties(qps, rps)
	// if err != nil {
	// logger.Error("Failed to add properties to DB", err)
	// os.Exit(1)
	// }
	// logger.Info("Database updated with new properties")
	// }

	// if radarrStart {
	// go rdrr(R, N, qpid, rpid, radarrDefaultRootPath, radarrDefaultQualityProfile, logger)
	// }

	// http.HandleFunc("/radarr", radarrHandler)
	// PORT := os.Getenv("PORT")
	// if PORT == "" {
	// 	logger.Error("PORT not specified")
	// 	os.Exit(1)
	// }
	// err = http.ListenAndServe(":"+PORT, nil)
	// if errors.Is(err, http.ErrServerClosed) {
	// 	logger.Info("Server closed")
	// } else if err != nil {
	// 	logger.Error(fmt.Sprintf("Failed to listen on PORT %s", PORT), "error", err)
	// 	os.Exit(1)
	// }
}

func rdrr(R *radarr.RadarrClient, N *notion.NotionClient, qpid map[string]int, rpid map[string]string, logger *slog.Logger) {

	for {
		logger.Info("Radarr: Fetching titles")
		data, err := N.QueryDB("Movie")
		if err != nil {
			logger.Error("Radarr: ", "Failed to query watchlist DB", err)
		}
		logger.Info(fmt.Sprintf("Radarr: Fetched titles from DB: %d", len(data.Results)))
		for _, v := range data.Results {
			if v.Properties.RootFolder.Select.Name == "" {
				v.Properties.RootFolder.Select.Name = R.DefaultRootPath
			}
			if v.Properties.QualityProfile.Select.Name == "" {
				v.Properties.QualityProfile.Select.Name = R.DefaultQualityProfile
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

func getRadarrInfo(qpid map[string]int, rpid map[string]string, radarrDefaultRootPath string, radarrDefaultQualityProfile string) error {
	radarrRootPaths, err := R.GetRootFolder()
	if len(radarrRootPaths) == 0 || err != nil {
		logger.Error("Failed to fetch Radarr root path", "error", err)
		return errors.New("RADARR ROOT PATH ERROR")
	} else {
		for _, r := range radarrRootPaths {
			rpid["Movie: "+r.Path] = r.Path
		}
		if radarrDefaultRootPath == "" {
			R.DefaultRootPath = "Movie: " + radarrRootPaths[0].Path
		} else {
			R.DefaultRootPath = "Movie: " + radarrDefaultRootPath
		}
	}
	// Quality Profiles
	radarrQualityProfiles, err := R.GetQualityProfiles()
	if len(radarrQualityProfiles) == 0 || err != nil {
		logger.Error("Failed to fetch Radarr quality profiles", "error", err)
		return errors.New("RADARR QUALITY PATH ERROR")
	} else {
		for _, v := range radarrQualityProfiles {
			qpid["Movie: "+v.Name] = v.Id
		}
		if radarrDefaultQualityProfile == "" {
			R.DefaultQualityProfile = "Movie: " + radarrQualityProfiles[0].Name
		} else {
			R.DefaultQualityProfile = "Movie: " + radarrDefaultQualityProfile
		}
	}
	return nil
}
