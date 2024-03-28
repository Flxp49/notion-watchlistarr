package routine

import (
	"log/slog"
	"time"

	"github.com/flxp49/notion-watchlist-radarr-sonarr/internal/notion"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/internal/radarr"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/internal/sonarr"
	"github.com/flxp49/notion-watchlist-radarr-sonarr/internal/util"
)

// This function fetches movies set to download from the notion watchlist
func RadarrSync(Logger *slog.Logger, N *notion.NotionClient, R *radarr.RadarrClient, t time.Duration) {
start:
	Logger.Info("RadarrSync", "Status", "Fetching titles from database")
	movies, err := N.QueryDB("Movie")
	if err != nil {
		Logger.Error("RadarrSync", "Failed to query watchlist DB", err)
		goto start
	}
	Logger.Info("RadarrSync", "Status", "Fetched titles from DB", "No of titles fetched", len(movies.Results))
	for _, m := range movies.Results {
		tmdbid, err := R.LookupMovieByImdbid(m.Properties.Imdbid.Rich_text[0].Plain_text)
		if err != nil {
			Logger.Error("RadarrSync", "Failed to fetch tmdbid via radarr of imdbid", m.Properties.Imdbid.Rich_text[0].Plain_text, "Error", err)
			continue
		}
		// set monitor property
		if m.Properties.MonitorProfile.Select.Name == "" {
			monitorProfile, err := N.GetNotionMonitorProp(R.DefaultMonitorProfile, "Movie")
			if err != nil {
				Logger.Error("RadarrSync", "Could not get notion monitor property for default monitor profile", R.DefaultMonitorProfile)
				goto end
			}
			m.Properties.MonitorProfile.Select.Name = monitorProfile
		}
		//get rootpath and qualityprofile properties for notion db
		qualityProp, rootPathProp, err := N.GetNotionQualityAndRootProps(R.DefaultQualityProfile, R.DefaultRootPath, "Movie")
		if err != nil {
			Logger.Error("RadarrSync", "Failed to fetch notion DB property", err)
			goto end
		}
		// set root folder property
		if m.Properties.RootFolder.Select.Name == "" {
			m.Properties.RootFolder.Select.Name = rootPathProp
		}
		// set qualty profile property
		if m.Properties.QualityProfile.Select.Name == "" {
			m.Properties.QualityProfile.Select.Name = qualityProp
		}
		Logger.Info("RadarrSync", "Status", "Adding Title", "Title", m.Properties.Name.Title[0].Plain_text)
		err = R.AddMovie(m.Properties.Name.Title[0].Plain_text, N.Qpid[m.Properties.QualityProfile.Select.Name], tmdbid.Tmdbid, N.Rpid[m.Properties.RootFolder.Select.Name], true, true, notion.MonitorProfiles[m.Properties.MonitorProfile.Select.Name])
		//check for exists error (movie already exists in radarr)
		exists, err := util.ExistingTitleErrorHandle(err)
		if err != nil {
			Logger.Error("RadarrSync", "Error adding title", m.Properties.Name.Title[0].Plain_text, "Error", err)
			continue
		}
		if !exists {
			Logger.Info("RadarrSync", "Status", "Added title", "Title", m.Properties.Name.Title[0].Plain_text)
			continue
		}
		// movie exists
		// check if downloaded or not
		//? make a put request to update the movie?

		movie, err := R.GetMovie(tmdbid.Tmdbid)
		if err != nil {
			Logger.Error("RadarrSync", "Failed to fetch movie details from radarr", err)
			continue
		}
		//get rootpath and qualityprofile properties for notion db
		qualityProp, rootPathProp, err = N.GetNotionQualityAndRootProps(movie[0].QualityProfileId, movie[0].RootFolderPath, "Movie")
		if err != nil {
			Logger.Error("RadarrSync", "Failed to fetch notion DB property", err)
			goto end
		}
		if movie[0].HasFile {
			err = N.UpdateDownloadStatus(m.Pgid, false, "Downloaded", qualityProp, rootPathProp)
			if err != nil {
				Logger.Error("RadarrSync", "Failed to update download status in notion watchlist", err)
				continue
			}
		} else {
			//check for queue status
			queueDetails, err := R.GetQueueDetails(movie[0].MovieID)
			if err != nil {
				Logger.Error("RadarrSync", "Failed to fetch queue details from radarr", err)
				continue
			}
			downloadStatus, err := util.GetDownloadStatus(queueDetails)
			if err != nil {
				Logger.Error("RadarrSync", "Failed to get download status", err)
				continue
			}
			err = N.UpdateDownloadStatus(m.Pgid, false, downloadStatus, qualityProp, rootPathProp)
			if err != nil {
				Logger.Error("RadarrSync", "Failed to update notion watchlist", err)
				continue
			}
		}
	}
	time.Sleep(t * time.Second)
	goto start
end:
	Logger.Error("RadarrSync", "Shutting down RadarrSync routine", "Error")
}

// This function fetches series set to download from the notion watchlist
func SonarrSync(Logger *slog.Logger, N *notion.NotionClient, S *sonarr.SonarrClient, t time.Duration) {
start:
	Logger.Info("SonarrSync", "Status", "Fetching titles from database")
	data, err := N.QueryDB("TV Series")
	if err != nil {
		Logger.Error("SonarrSync", "Failed to query watchlist DB", err)
		goto start
	}
	Logger.Info("SonarrSync", "Status", "Fetched titles from DB", "No of titles fetched", len(data.Results))
	for _, m := range data.Results {
		tvdbid, err := S.LookupSeriesByImdbid(m.Properties.Imdbid.Rich_text[0].Plain_text)
		if err != nil || len(tvdbid) == 0 {
			Logger.Error("SonarrSync", "Failed to fetch tvdbid via sonarr of imdbid", m.Properties.Imdbid.Rich_text[0].Plain_text, "Error", err)
			continue
		}
		// set monitor property
		if m.Properties.MonitorProfile.Select.Name == "" {
			monitorProfile, err := N.GetNotionMonitorProp(S.DefaultMonitorProfile, "TV Series")
			if err != nil {
				Logger.Error("SonarrSync", "Could not get notion monitor property for default monitor profile", S.DefaultMonitorProfile)
				goto end
			}
			m.Properties.MonitorProfile.Select.Name = monitorProfile
		}

		//get rootpath and qualityprofile properties for notion db
		qualityProp, rootPathProp, err := N.GetNotionQualityAndRootProps(S.DefaultQualityProfile, S.DefaultRootPath, "TV Series")
		if err != nil {
			Logger.Error("SonarrSync", "Failed to fetch notion DB property", err)
			goto end
		}
		// set root folder property
		if m.Properties.RootFolder.Select.Name == "" {
			m.Properties.RootFolder.Select.Name = rootPathProp
		}
		// set quality profile property
		if m.Properties.QualityProfile.Select.Name == "" {
			m.Properties.QualityProfile.Select.Name = qualityProp
		}
		Logger.Info("SonarrSync", "Status", "Adding Title", "Title", m.Properties.Name.Title[0].Plain_text)
		err = S.AddSeries(m.Properties.Name.Title[0].Plain_text, N.Qpid[m.Properties.QualityProfile.Select.Name], tvdbid[0].TvdbId, N.Rpid[m.Properties.RootFolder.Select.Name], true, true, true, notion.MonitorProfiles[m.Properties.MonitorProfile.Select.Name]) //todo
		//check for exists error (series already exists in radarr)
		exists, err := util.ExistingTitleErrorHandle(err)
		if err != nil {
			Logger.Error("SonarrSync", "Error adding title", m.Properties.Name.Title[0].Plain_text, "Error", err)
			continue
		}
		if !exists {
			Logger.Info("SonarrSync", "Status", "Added title", "Title", m.Properties.Name.Title[0].Plain_text)
			continue
		}
		// series exists
		// check if downloaded or not
		//? make a put request to update the movie?
		series, err := S.GetSeries(tvdbid[0].TvdbId)
		if err != nil {
			Logger.Error("SonarrSync", "Failed to fetch movie details from sonarr", err)
			continue
		}
		qualityProp, rootPathProp, err = N.GetNotionQualityAndRootProps(series[0].QualityProfileId, series[0].RootFolderPath, "TV Series")
		if err != nil {
			Logger.Error("SonarrSync", "Failed to fetch notion DB property", err)
			goto end
		}
		if series[0].Statistics.PercentOfEpisodes == 100 {
			err = N.UpdateDownloadStatus(m.Pgid, false, "Downloaded", qualityProp, rootPathProp)
			if err != nil {
				Logger.Error("SonarrSync", "Failed to update download status in notion watchlist", err)
				continue
			}
		} else {
			//check for download queue
			queueDetails, err := S.GetQueueDetails(series[0].SeriesID)
			if err != nil {
				Logger.Error("SonarrSync", "Failed to fetch queue details from sonarr", err)
				continue
			}
			downloadStatus, err := util.GetDownloadStatus(queueDetails)
			if err != nil {
				Logger.Error("SonarrSync", "Failed to get download status", err)
				continue
			}
			err = N.UpdateDownloadStatus(m.Pgid, false, downloadStatus, qualityProp, rootPathProp)
			if err != nil {
				Logger.Error("SonarrSync", "Failed to update notion watchlist", err)
				continue
			}
		}
	}
	time.Sleep(t * time.Second)
	goto start
end:
	Logger.Error("SonarrSync", "Shutting down SonarrSync routine", "Error")
}
