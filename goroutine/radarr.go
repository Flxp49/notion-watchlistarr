package goroutine

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/flxp49/notion-watchlist-radarr-sonarr/notion"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/radarr"
)

func Rdrr(Logger *slog.Logger, N *notion.NotionClient, R *radarr.RadarrClient) {

	for {
		Logger.Info("Radarr: Fetching titles")
		data, err := N.QueryDB("Movie")
		if err != nil {
			Logger.Error("Radarr: ", "Failed to query watchlist DB", err)
			continue
		}
		Logger.Info(fmt.Sprintf("Radarr: Fetched titles from DB: %d", len(data.Results)))
		for _, v := range data.Results {
			if v.Properties.RootFolder.Select.Name == "" {
				v.Properties.RootFolder.Select.Name = R.DefaultRootPath
			}
			if v.Properties.QualityProfile.Select.Name == "" {
				v.Properties.QualityProfile.Select.Name = R.DefaultQualityProfile
			}
			Logger.Info(fmt.Sprintf("Radarr: Adding Title: %s", v.Properties.Name.Title[0].Plain_text))
			err = R.AddMovie(v.Properties.Name.Title[0].Plain_text, N.Qpid[v.Properties.QualityProfile.Select.Name], v.Properties.Tmdbid.Number, N.Rpid[v.Properties.RootFolder.Select.Name], true, true)
			if err != nil {
				Logger.Error(fmt.Sprintf("Radarr: Error adding title: %s", v.Properties.Name.Title[0].Plain_text), err)
				N.UpdateDownloadStatus(v.Pgid, false, "Error", "", "")
				continue
			}
			Logger.Info(fmt.Sprintf("Radarr: Added title: %s", v.Properties.Name.Title[0].Plain_text))
		}
		time.Sleep(10 * time.Second)
	}
}
