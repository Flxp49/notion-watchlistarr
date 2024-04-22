package sonarr

import (
	"log"
	"os"
	"testing"

	"github.com/flxp49/notion-watchlist-radarr-sonarr/internal/util"
	"github.com/joho/godotenv"
)

var Sonarr *SonarrClient

func TestMain(m *testing.M) {
	err := godotenv.Load("../../cmd/notionwatchlistarr/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	Sonarr = InitSonarrClient(os.Getenv("SONARR_KEY"), os.Getenv("SONARR_HOST"))
	os.Exit(m.Run())
}

func TestLookupSeriesByTvdbid(t *testing.T) {
	series, err := Sonarr.LookupSeriesByImdbid("")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(series)
}

func TestAddSeries(t *testing.T) {
	err := Sonarr.AddSeries("Doom Patrol", 4, 355622, "D:\\Media\\Shows", true, true, true, "LastSeason")
	if err != nil {
		t.Fatal(err)
	}
}

func TestAddExistingSeries(t *testing.T) {
	err := Sonarr.AddSeries("Lockwood & Co.", 4, 422028, "D:\\Media\\Shows", true, true, true, "AllEpisodes")
	if err != nil {
		t.Log(err)
		t.Log(util.ExistingTitleErrorHandle(err))
	}
}

func TestGetSeries(t *testing.T) {
	series, err := Sonarr.GetSeries(399987)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(series)
}

func TestGetAllSeries(t *testing.T) {
	series, err := Sonarr.GetSeries(-1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(series))
}

func TestGetQueueDetails(t *testing.T) {
	//the witcher blood moon
	downloadStatus, err := Sonarr.GetQueueDetails(36)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(downloadStatus)
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
func TestGetSeriesByRemoteID(t *testing.T) {
	id, err := util.GetSeriesByRemoteID("tt1751634")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(id)
}
