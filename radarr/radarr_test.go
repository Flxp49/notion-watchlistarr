package radarr

import (
	"log"
	"os"
	"testing"

	"github.com/flxp49/notion-watchlist-radarr-sonarr/util"
	"github.com/joho/godotenv"
)

var Radarr *RadarrClient

func TestMain(m *testing.M) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	Radarr = InitRadarrClient(os.Getenv("RADARR_KEY"), os.Getenv("RADARR_HOST"))
	os.Exit(m.Run())
}

func TestGetQueueDetails(t *testing.T) {
	downloadStatus, err := Radarr.GetQueueDetails(126)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(downloadStatus)
}

func TestAddMovie(t *testing.T) {
	err := Radarr.AddMovie("American Psycho", 4, 1359, "D:\\Media\\Movies", true, true, "MovieandCollection")
	if err != nil {
		t.Fatal(err)
	}
}
func TestAddExistingMovie(t *testing.T) {
	err := Radarr.AddMovie("Finch", 4, 522402, "D:\\Media\\Movies", true, true, "MovieandCollection")
	if err != nil {
		t.Log(err)
		t.Log(util.ExistingTitleErrorHandle(err))
	}
}

func TestGetMovie(t *testing.T) {
	movie, err := Radarr.GetMovie(557)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(movie)
}

func TestGetAllMovie(t *testing.T) {
	movie, err := Radarr.GetMovie(-1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(movie))
}

func TestGetRootFolder(t *testing.T) {
	rootFolder, err := Radarr.GetRootFolder()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rootFolder)
}
func TestGetQualityProfiles(t *testing.T) {
	qualityProfiles, err := Radarr.GetQualityProfiles()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(qualityProfiles)
}
