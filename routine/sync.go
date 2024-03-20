package routine

import (
	"log/slog"
	"time"

	"github.com/flxp49/notion-watchlist-radarr-sonarr/notion"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/radarr"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/sonarr"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/util"
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
			// set monitor property
			if v.Properties.MonitorProfile.Select.Name == "" {
				monitorProfile, err := N.GetNotionMonitorProp(R.DefaultMonitorProfile)
				if err != nil {
					Logger.Error("RadarrSync", "Could not get notion monitor property for default monitor profile", R.DefaultMonitorProfile)
					continue
				}
				v.Properties.MonitorProfile.Select.Name = monitorProfile
			}
			// set root folder property
			if v.Properties.RootFolder.Select.Name == "" {
				rootFolder, err := N.GetNotionRootPathProp(R.DefaultRootPath)
				if err != nil {
					Logger.Error("RadarrSync", "Could not get notion root path property for default root path", R.DefaultRootPath)
					continue
				}
				v.Properties.RootFolder.Select.Name = rootFolder
			}
			// set qualty profile property
			if v.Properties.QualityProfile.Select.Name == "" {
				qualityProfile, err := N.GetNotionQualityProfileProp(R.DefaultQualityProfile)
				if err != nil {
					Logger.Error("RadarrSync", "Could not get notion quality profile property for default quality profile", R.DefaultQualityProfile)
					continue
				}
				v.Properties.QualityProfile.Select.Name = qualityProfile
			}
			Logger.Info("RadarrSync", "Adding Title", v.Properties.Name.Title[0].Plain_text)
			err = R.AddMovie(v.Properties.Name.Title[0].Plain_text, N.Qpid[v.Properties.QualityProfile.Select.Name], v.Properties.Tmdbid.Number, N.Rpid[v.Properties.RootFolder.Select.Name], true, true, notion.MonitorProfiles[v.Properties.MonitorProfile.Select.Name])
			//check for exists error (movie already exists in radarr)
			exists, err := util.ExistingTitleErrorHandle(err)
			if err != nil {
				Logger.Error("RadarrSync", "Error adding title", v.Properties.Name.Title[0].Plain_text, "error", err)
				N.UpdateDownloadStatus(v.Pgid, false, "Error", "", "")
				continue
			}
			if !exists {
				Logger.Info("RadarrSync", "Added title", v.Properties.Name.Title[0].Plain_text)
				continue
			}
			// movie exists
			// check if downloaded or not
			//? make a put request to update the movie?
			movie, err := R.GetMovie(v.Properties.Tmdbid.Number)
			if err != nil {
				Logger.Error("RadarrSync", "Failed to fetch movie details from radarr for tmdbid", v.Properties.Tmdbid.Number)
				continue
			}
			qualityProp, _ := N.GetNotionQualityProfileProp(movie[0].QualityProfileId)
			rootFolderProp, _ := N.GetNotionRootPathProp(movie[0].RootFolderPath)
			if movie[0].HasFile {
				N.UpdateDownloadStatus(v.Pgid, true, "Downloaded", qualityProp, rootFolderProp)
			} else {
				N.UpdateDownloadStatus(v.Pgid, true, "Downloading", qualityProp, rootFolderProp)
			}
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
			// set monitor property
			if v.Properties.MonitorProfile.Select.Name == "" {
				monitorProfile, err := N.GetNotionMonitorProp(S.DefaultMonitorProfile)
				if err != nil {
					Logger.Error("SonarrSync", "Could not get notion monitor property for default monitor profile", S.DefaultMonitorProfile)
					continue
				}
				v.Properties.MonitorProfile.Select.Name = monitorProfile
			}
			// set root folder property
			if v.Properties.RootFolder.Select.Name == "" {
				rootFolder, err := N.GetNotionRootPathProp(S.DefaultRootPath)
				if err != nil {
					Logger.Error("SonarrSync", "Could not get notion root path property for default root path", S.DefaultRootPath)
					continue
				}
				v.Properties.RootFolder.Select.Name = rootFolder
			}
			// set quality profile property
			if v.Properties.QualityProfile.Select.Name == "" {
				qualityProfile, err := N.GetNotionQualityProfileProp(S.DefaultQualityProfile)
				if err != nil {
					Logger.Error("SonarrSync", "Could not get notion quality profile property for default quality profile", S.DefaultQualityProfile)
					continue
				}
				v.Properties.QualityProfile.Select.Name = qualityProfile
			}
			tvdbid, err := S.LookupSeriesByTmdbid(v.Properties.Tmdbid.Number)
			if err != nil || len(tvdbid) == 0 {
				Logger.Error("SonarrSync", "Failed to fetch tvdbid of tmdbid", v.Properties.Tmdbid.Number)
				continue
			}
			Logger.Info("SonarrSync", "Adding Title", v.Properties.Name.Title[0].Plain_text)
			err = S.AddSeries(v.Properties.Name.Title[0].Plain_text, N.Qpid[v.Properties.QualityProfile.Select.Name], tvdbid[0].TvdbId, N.Rpid[v.Properties.RootFolder.Select.Name], true, true, true, notion.MonitorProfiles[v.Properties.MonitorProfile.Select.Name]) //todo
			//check for exists error (series already exists in radarr)
			exists, err := util.ExistingTitleErrorHandle(err)
			if err != nil {
				Logger.Error("SonarrSync", "Error adding title", v.Properties.Name.Title[0].Plain_text, "error", err)
				N.UpdateDownloadStatus(v.Pgid, false, "Error", "", "")
				continue
			}
			if !exists {
				Logger.Info("SonarrSync", "Added title", v.Properties.Name.Title[0].Plain_text)
				continue
			}
			// series exists
			// check if downloaded or not
			//? make a put request to update the movie?
			series, err := S.GetSeries(tvdbid[0].TvdbId)
			if err != nil {
				Logger.Error("SonarrSync", "Failed to fetch movie details from sonarr for tmdbid", v.Properties.Tmdbid.Number)
				continue
			}
			qualityProp, _ := N.GetNotionQualityProfileProp(series[0].QualityProfileId)
			rootFolderProp, _ := N.GetNotionRootPathProp(series[0].RootFolderPath)
			if series[0].Statistics.PercentOfEpisodes == 100 {
				N.UpdateDownloadStatus(v.Pgid, true, "Downloaded", qualityProp, rootFolderProp)
			} else {
				N.UpdateDownloadStatus(v.Pgid, true, "Downloading", qualityProp, rootFolderProp)
			}
		}
		time.Sleep(10 * time.Second)
	}
}
