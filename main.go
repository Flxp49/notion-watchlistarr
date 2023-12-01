package main

import (
	"log"
	"os"
	"time"

	"github.com/flxp49/notion-watchlist-radarr-sonarr/notion"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/radarr"
	"github.com/joho/godotenv"
)

var exit = make(chan bool)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	R := radarr.InitRadarrClient(os.Getenv("RDRRKEY"), os.Getenv("RADARRHOST"))
	N := notion.InitNotionClient("Emad", os.Getenv("RDRRNOTIONINTEG"), os.Getenv("DBID"))
	radarrSucc := true

	// Root Paths
	var rps []string
	rpid := make(map[string]string)
	// Radarr root path
	radarrRootPaths, err := R.GetRootFolder()
	if len(radarrRootPaths) == 0 || err != nil {
		log.Println("Failed to fetch Radarr root path", err)
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
		log.Println("Failed to fetch Radarr quality profiles", err)
		radarrSucc = false
	} else {
		for _, v := range radarrQualityProfiles {
			qps = append(qps, "Movie: "+v.Name)
			qpid["Movie: "+v.Name] = v.Id
		}
		log.Println("Quality profiles fetched:", qps)
	}
	if !radarrSucc {
		log.Fatalln("Failed to fetch quality profiles & rootpaths from Radarr")
	}
	// Add properties to the DB
	err = N.AddDBProperties(qps, rps)
	if err != nil {
		log.Fatalln("Failed to add properties to DB", err)
	}
	log.Println("Database updated with new properties")
	if radarrSucc {
		go rdrr(R, N, qpid, rpid, radarrRootPaths[0].Path, radarrQualityProfiles[0].Id)
	}
	<-exit
	log.Fatalln("Shutting down due to termination of Radarr subroutine")
}

func rdrr(R *radarr.RadarrClient, N *notion.NotionClient, qpid map[string]int, rpid map[string]string, defaultRootPath string, defaultQualityProfile int) {

	for {
		data, err := N.QueryDB("Movie")
		if err != nil {
			log.Println("Radarr: Failed to query watchlist DB", err)
			exit <- true
		}
		log.Println("Fetched titles from DB:", len(data.Results))
		log.Println(defaultRootPath)
		log.Println(defaultQualityProfile)
		var rp string
		var qp int
		for _, v := range data.Results {
			if v.Properties.RootFolder.Select.Name == "" {
				rp = defaultRootPath
			} else {
				rp = rpid[v.Properties.RootFolder.Select.Name]
			}
			if v.Properties.QualityProfile.Select.Name == "" {
				qp = defaultQualityProfile
			} else {
				qp = qpid[v.Properties.QualityProfile.Select.Name]
			}
			err = R.AddMovie(v.Properties.Name.Title[0].Plain_text, qp, v.Properties.Tmdbid.Number, rp, true, true)
			if err != nil {
				log.Println("Error adding title:", v.Properties.Name.Title[0].Plain_text, err)
				N.UpdateDownloadStatus(v.Pgid, "Error")
				continue
			}
			log.Println("Added title:", v.Properties.Name.Title[0].Plain_text)
			N.UpdateDownloadStatus(v.Pgid, "Queued")
		}
		time.Sleep(3 * time.Second)
	}

}
