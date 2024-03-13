package routine

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/flxp49/notion-watchlist-radarr-sonarr/notion"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/radarr"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/sonarr"
)

// This function fetches movies set to download from the notion watchlist
func RadarrSync(Logger *slog.Logger, N *notion.NotionClient, R *radarr.RadarrClient) {

	for {
		Logger.Info("RadarrSync", "Status", "Fetching titles from database")
		data, err := N.QueryDB("Movie")
		if err != nil {
			Logger.Error("RadarrSync", "Failed to query watchlist DB", err)
			continue
		}
		Logger.Info("RadarrSync", "Fetched titles from DB:", len(data.Results))
		for _, v := range data.Results {
			if v.Properties.RootFolder.Select.Name == "" {
				v.Properties.RootFolder.Select.Name = R.DefaultRootPath
			}
			if v.Properties.QualityProfile.Select.Name == "" {
				v.Properties.QualityProfile.Select.Name = R.DefaultQualityProfile
			}
			Logger.Info(fmt.Sprintf("Radarr: Adding Title: %s", v.Properties.Name.Title[0].Plain_text))
			err = R.AddMovie(v.Properties.Name.Title[0].Plain_text, N.Qpid[v.Properties.QualityProfile.Select.Name], v.Properties.Tmdbid.Number, N.Rpid[v.Properties.RootFolder.Select.Name], true, true, "")
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

// This function fetches series set to download from the notion watchlist
func SonarrSync(Logger *slog.Logger, N *notion.NotionClient, S *sonarr.SonarrClient) {
	for {
		Logger.Info("SonarrSync", "Status", "Fetching titles from database")
		data, err := N.QueryDB("TV Series")
		if err != nil {
			Logger.Error("SonarrSync", "Failed to query watchlist DB", err)
			continue
		}
		Logger.Info("SonarrSync", "Fetched titles from DB", len(data.Results))
		for _, v := range data.Results {
			if v.Properties.RootFolder.Select.Name == "" {
				v.Properties.RootFolder.Select.Name = S.DefaultRootPath
			}
			if v.Properties.QualityProfile.Select.Name == "" {
				v.Properties.QualityProfile.Select.Name = S.DefaultQualityProfile
			}
			tvdbid, err := S.LookupSeriesByTmdbid(v.Properties.Tmdbid.Number)
			if err != nil || len(tvdbid) == 0 {
				Logger.Error("SonarrSync", "Failed to fetch tvdbid of tmdbid", v.Properties.Tmdbid.Number)
			}
			Logger.Info("SonarrSync", "Adding Title", v.Properties.Name.Title[0].Plain_text)
			err = S.AddSeries(v.Properties.Name.Title[0].Plain_text, N.Qpid[v.Properties.QualityProfile.Select.Name], tvdbid[0].TvdbId, N.Rpid[v.Properties.RootFolder.Select.Name], true, true, true, "") //todo
			if err != nil {
				Logger.Error("SonarrSync", "Error adding title", v.Properties.Name.Title[0].Plain_text, "error", err)
				N.UpdateDownloadStatus(v.Pgid, false, "Error", "", "")
				continue
			}
			Logger.Info("SonarrSync", "Added title", v.Properties.Name.Title[0].Plain_text)
		}
		time.Sleep(10 * time.Second)
	}
}
