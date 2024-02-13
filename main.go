package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/flxp49/notion-watchlist-radarr-sonarr/notion"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/radarr"
	"github.com/joho/godotenv"
)

var logger *slog.Logger
var N *notion.NotionClient
var R *radarr.RadarrClient

// To manage Root Paths and Quality Profiles and update Notion DB with it.
var rpid map[string]string
var qpid map[string]int

func parseJson(body []byte, target interface{}) error {
	return json.Unmarshal(body, target)
}

func getNotionQualityProfileProp(keyVal int) (string, error) {
	for key, val := range qpid {
		if val == keyVal {
			return key, nil
		}
	}
	return "", errors.New("invalid qpid key passed")
}
func getNotionRootPathProp(keyVal string) (string, error) {
	for key, val := range rpid {
		if filepath.Clean(val) == filepath.Clean(keyVal) {
			return key, nil
		}
	}
	return "", errors.New("invalid rpid key passed")
}

type MovieInfo struct {
	Movie struct {
		Id     int `json:"id"`
		TmdbId int `json:"tmdbId"`
	} `json:"movie"`
	EventType    string `json:"eventType"`    //allowed values MovieAdded|Grab|Download|MovieDelete
	DeletedFiles bool   `json:"deletedFiles"` //only present when EventType: "MovieDelete"
}

func radarrHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		logger.Error("Radarr Webhook Error", "error", "Wrong request method")
		return
	}
	body, _ := io.ReadAll(r.Body)
	r.Body.Close()
	var movieData MovieInfo
	err := parseJson(body, &movieData)
	if err != nil {
		w.WriteHeader(400)
		logger.Error("Radarr Webhook Error", "body", body, "error", err)
		return
	}
	w.WriteHeader(200)
	logger.Info("Radarr post request", "data", movieData)
	page, err := N.QueryDBTmdb(movieData.Movie.TmdbId)
	if err != nil {
		logger.Error("Radarr Webhook Error", "error", err)
		return
	}
	if len(page.Results) != 0 { // title exists, update status in watchlist
		movie, err := R.GetMovie(movieData.Movie.TmdbId)
		if err != nil {
			logger.Error("Radarr Webhook Error", "error", err)
			return
		}
		switch movieData.EventType {
		case "MovieAdded":
			//query for movie in watchlist
			movieQualityProp, err := getNotionQualityProfileProp(movie[0].QualityProfileId)
			if err != nil {
				logger.Error("Radarr Webhook Error: Failed to fetch notion DB movie quality profile property value", "error", err)
				return
			}
			rootPathProp, err := getNotionRootPathProp(movie[0].RootFolderPath)
			if err != nil {
				logger.Error("Radarr Webhook Error: Failed to fetch notion DB rootfolder property value", "error", err)
				return
			}
			if movie[0].HasFile {
				err = N.UpdateDownloadStatus(page.Results[0].Pgid, true, "Downloaded", movieQualityProp, rootPathProp)
			} else {
				err = N.UpdateDownloadStatus(page.Results[0].Pgid, true, "Queued", movieQualityProp, rootPathProp)
			}
			if err != nil {
				logger.Error("Radarr Webhook Error: Failed to update download status in watchlist", "error", err)
			}

		case "Grab":
			movie, err := R.GetMovie(movieData.Movie.TmdbId)
			if err != nil {
				logger.Error("Radarr Webhook Error", "error", err)
				return
			}
			movieQualityProp, err := getNotionQualityProfileProp(movie[0].QualityProfileId)
			if err != nil {
				logger.Error("Radarr Webhook Error: Failed to fetch notion DB movie quality profile property value", "error", err)
				return
			}
			rootPathProp, err := getNotionRootPathProp(movie[0].RootFolderPath)
			if err != nil {
				logger.Error("Radarr Webhook Error: Failed to fetch notion DB rootfolder property value", "error", err)
				return
			}
			err = N.UpdateDownloadStatus(page.Results[0].Pgid, true, "Downloading", movieQualityProp, rootPathProp)
			if err != nil {
				logger.Error("Radarr Webhook Error: Failed to update download status in watchlist", "error", err)
			}
		case "Download":
			movieQualityProp, err := getNotionQualityProfileProp(movie[0].QualityProfileId)
			if err != nil {
				logger.Error("Radarr Webhook Error: Failed to fetch notion DB movie quality profile property value", "error", err)
				return
			}
			rootPathProp, err := getNotionRootPathProp(movie[0].RootFolderPath)
			if err != nil {
				logger.Error("Radarr Webhook Error: Failed to fetch notion DB rootfolder property value", "error", err)
				return
			}
			err = N.UpdateDownloadStatus(page.Results[0].Pgid, true, "Downloaded", movieQualityProp, rootPathProp)
			if err != nil {
				logger.Error("Radarr Webhook Error: Failed to update download status in watchlist", "error", err)
			}
		case "MovieDelete":
			if !movieData.DeletedFiles {
				return
			}
			err = N.UpdateDownloadStatus(page.Results[0].Pgid, false, "Not Downloaded", "", "")
			if err != nil {
				logger.Error("Radarr Webhook Error: Failed to update download status in watchlist", "error", err)
			}
		default:
			logger.Error("Radarr Webhook Error", "error", "EventType not valid in payload")
		}
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

	R = radarr.InitRadarrClient(os.Getenv("RADARR_KEY"), os.Getenv("RADARR_HOST"))
	N = notion.InitNotionClient("Emad", os.Getenv("NOTION_INTEGRATION_SECRET"), os.Getenv("NOTION_DB_ID"))

	// To manage Root Paths and Quality Profiles and update Notion DB with it.
	rpid = make(map[string]string)
	qpid = make(map[string]int)
	// monitor := make(map[string]int)
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

	err = getRadarrInfo(radarrDefaultRootPath, radarrDefaultQualityProfile)
	if err != nil {
		logger.Error("Error fetching Radarr details, Radarr routine not initialized", "Error", err)
		radarrStart = false
	}
	//todo: same as the above for sonarr

	if radarrStart { // || sonarrStart
		// Add properties to the DB
		err = N.AddDBProperties(qpid, rpid)
		if err != nil {
			logger.Error("Failed to add properties to DB", "Error", err)
			os.Exit(1)
		}
		logger.Info("Database updated with new properties")

		// if radarrStart {
		// go rdrr()
		// }
		// same for sonarr
		// if sonarrStart {
		// go snrr()
		// }
	}

	http.HandleFunc("/radarr", radarrHandler)
	PORT := os.Getenv("PORT")
	if PORT == "" {
		logger.Error("PORT not specified")
		os.Exit(1)
	}
	err = http.ListenAndServe(":"+PORT, nil)
	if errors.Is(err, http.ErrServerClosed) {
		logger.Info("Server closed")
	} else if err != nil {
		logger.Error(fmt.Sprintf("Failed to listen on PORT %s", PORT), "error", err)
		os.Exit(1)
	}
}

func rdrr() {

	for {
		logger.Info("Radarr: Fetching titles")
		data, err := N.QueryDB("Movie")
		if err != nil {
			logger.Error("Radarr: ", "Failed to query watchlist DB", err)
			continue
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
				N.UpdateDownloadStatus(v.Pgid, false, "Error", "", "")
				continue
			}
			logger.Info(fmt.Sprintf("Radarr: Added title: %s", v.Properties.Name.Title[0].Plain_text))
		}
		time.Sleep(10 * time.Second)
	}
}

func getRadarrInfo(radarrDefaultRootPath string, radarrDefaultQualityProfile string) error {
	// Root path
	radarrRootPaths, err := R.GetRootFolder()
	if len(radarrRootPaths) == 0 || err != nil {
		logger.Error("Failed to fetch Radarr root path", "error", err)
		return errors.New("RADARR ROOT PATH ERROR")
	}

	for _, r := range radarrRootPaths {
		rpid["Movie: "+r.Path] = r.Path
	}
	if radarrDefaultRootPath == "" {
		R.DefaultRootPath = "Movie: " + radarrRootPaths[0].Path
	} else {
		R.DefaultRootPath = "Movie: " + radarrDefaultRootPath
	}
	// Quality Profiles
	radarrQualityProfiles, err := R.GetQualityProfiles()
	if len(radarrQualityProfiles) == 0 || err != nil {
		logger.Error("Failed to fetch Radarr quality profiles", "error", err)
		return errors.New("RADARR QUALITY PATH ERROR")
	}

	for _, v := range radarrQualityProfiles {
		qpid["Movie: "+v.Name] = v.Id
	}
	if radarrDefaultQualityProfile == "" {
		R.DefaultQualityProfile = "Movie: " + radarrQualityProfiles[0].Name
	} else {
		R.DefaultQualityProfile = "Movie: " + radarrDefaultQualityProfile
	}
	return nil
}
