package sonarr

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var Sonarr *SonarrClient

func TestMain(m *testing.M) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	Sonarr = InitSonarrClient(os.Getenv("SONARR_KEY"), os.Getenv("SONARR_HOST"))
	os.Exit(m.Run())
}
func TestLookupSeriesByTvdbid(t *testing.T) {
	series, err := Sonarr.LookupSeriesByTmdbid(106541)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(series)
}

func TestAddSeries(t *testing.T) {
	err := Sonarr.AddSeries("The Witcher: Blood Origin", 4, 399987, "D:\\Media\\Shows", true, true, true, "None")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetSeries(t *testing.T) {
	series, err := Sonarr.GetSeries(399987)
	// series, err := Sonarr.GetSeries(422028)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(series)
}

func TestGetRootFolder(t *testing.T) {
	rootFolder, err := Sonarr.GetRootFolder()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rootFolder)
}
func TestGetQualityProfiles(t *testing.T) {
	qualityProfiles, err := Sonarr.GetQualityProfiles()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(qualityProfiles)
}
