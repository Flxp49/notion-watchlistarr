package radarr

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var Radarr *RadarrClient

func TestMain(m *testing.M) {
	err := godotenv.Load("../../cmd/notionwatchlistarr/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	Radarr = InitRadarrClient(os.Getenv("RADARR_KEY"), os.Getenv("RADARR_HOST"))
	os.Exit(m.Run())
}

func TestLookupMovie(t *testing.T) {
	movie, err := Radarr.LookupMovie("tt0118929")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(movie)
}

func TestGetQueueDetails(t *testing.T) {
	downloadStatus, err := Radarr.GetQueueDetails(126)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(downloadStatus)
}

func TestAddMovie(t *testing.T) {
	movie, err := Radarr.LookupMovie("tt0078788")
	if err != nil {
		t.Fatal(err)
	}
	err = Radarr.AddMovie(movie, 4, "D:\\Media\\Movies", true, true, "MovieOnly")
	if err != nil {
		t.Fatal(err)
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
