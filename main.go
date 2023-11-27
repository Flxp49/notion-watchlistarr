package main

import (
	"log"
	"os"
	"time"

	"github.com/flxp49/notion-watchlist-radarr-sonarr/notion"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/radarr"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	R := radarr.Initrdrr(os.Getenv("RDRRKEY"), os.Getenv("HOST"))
	N := notion.InitNotionClient("Emad", os.Getenv("RDRRNOTIONINTEG"), os.Getenv("DBID"))

	d, err := R.GetQualityProfiles()
	if len(d) == 0 || err != nil {
		log.Fatalln("No quality profiles found", err)
	}
	var qps []string
	qpid := make(map[string]int)
	for _, v := range d {
		qps = append(qps, v.Name)
		qpid[v.Name] = v.Id
	}
	log.Println("Quality profiles fetched from %Radarr%:", qps)

	rootPath, err := R.GetRootFolder()
	if err != nil {
		log.Fatalln("Failed to fetch Radarr root path", err)
	}

	err = N.AddDBProperties(qps)
	if err != nil {
		log.Fatalln("Failed to add properties to DB", err)
	}
	log.Println("Updated Database with properties")
	data, err := N.QueryDB("Movie")
	if err != nil {
		log.Fatalln("Failed to add properties to DB", err)
	}
	log.Println("Fetched titles from DB: ", len(data.Results))
	for _, v := range data.Results {
		err = R.AddMovie(v.Properties.Name.Title[0].Plain_text, qpid[v.Properties.QualityProfile.Select.Name], v.Properties.Tmdbid.Number, rootPath[0].Path, true, true)
		if err != nil {
			log.Println("Error adding title:", v.Properties.Name.Title[0].Plain_text, err)
			continue
		}
		log.Println("Added title:", v.Properties.Name.Title[0].Plain_text)
	}
	time.Sleep(5 * time.Second)
}
