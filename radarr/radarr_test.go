package radarr

import (
	"log"
	"os"
	"testing"

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

func TestAddMovie(t *testing.T) {
	err := Radarr.AddMovie("American Psycho", 4, 1359, "D:\\Media\\Movies", true, true, "MovieandCollection")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetMovie(t *testing.T) {
	movie, err := Radarr.GetMovie(74)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(movie)
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
