package routine

import (
	"log/slog"
	"time"

	"github.com/flxp49/notion-watchlistarr/internal/constant"
	"github.com/flxp49/notion-watchlistarr/internal/notion"
	"github.com/flxp49/notion-watchlistarr/internal/radarr"
	"github.com/flxp49/notion-watchlistarr/internal/sonarr"
)

func RadarrWatchlistSync(Logger *slog.Logger, N *notion.NotionClient, R *radarr.RadarrClient, t time.Duration) {
	//fetch all movies from radarr along with download status
	//update notion watchlist
start:
	Logger.Info("RadarrWatchlistSync", "Status", "Radarr Watchlist Sync Start")
	movies, err := R.GetMovie(-1)
	Logger.Info("RadarrWatchlistSync", "Status", "Fetched movies from radarr library", "No of movies fetched", len(movies))
	if err != nil {
		Logger.Error("RadarrWatchlistSync", "Failed to fetch movies from radarr library", err)
		time.Sleep(10 * time.Second)
		goto start
	}
	for _, movie := range movies {
	m:
		// query for each movie in watchlist
		watchlistMovie, err := N.QueryDBImdb(movie.ImdbID)
		if err != nil {
			Logger.Error("RadarrWatchlistSync", "Failed to query movie from notion watchlist", err)
			time.Sleep(5 * time.Second)
			goto m
		}
		// if movie is not present in watchlist, skip
		if len(watchlistMovie.Results) == 0 {
			continue
		}
		monitoredProfile := constant.MovieOnly
		if movie.Collection.TmdbID != 0 {
			collectionMonitored, err := R.GetCollection(movie.Collection.TmdbID)
			if err != nil {
				Logger.Error("RadarrSync", "Failed to get collection details", err)
				N.UpdateDownloadStatus(watchlistMovie.Results[0].Pgid, false, "Error", "", "", "")
				continue
			}
			if collectionMonitored {
				monitoredProfile = constant.MovieAndCollection
			}
		}
		monitoredProfileNotionProp, _ := N.GetNotionMonitorProp(monitoredProfile, "Movie")
		//get rootpath and qualityprofile properties for notion db
		qualityProp, rootPathProp, err := N.GetNotionQualityAndRootProps(movie.QualityProfileID, movie.RootFolderPath, "Movie")
		if err != nil {
			Logger.Error("RadarrWatchlistSync", "Failed to fetch notion DB property", err)
			goto m
		}
		if movie.HasFile {
			err = N.UpdateDownloadStatus(watchlistMovie.Results[0].Pgid, false, "Downloaded", qualityProp, rootPathProp, monitoredProfileNotionProp)
			if err != nil {
				Logger.Error("RadarrWatchlistSync", "Failed to update download status in watchlist", err)
				goto m
			}
		} else {
			//check for queue status
			queueStatus, err := R.GetQueueDetails(movie.ID)
			if err != nil {
				Logger.Error("RadarrWatchlistSync", "Failed to fetch queue details from radarr", err)
				goto m
			}
			err = N.UpdateDownloadStatus(watchlistMovie.Results[0].Pgid, false, queueStatus, qualityProp, rootPathProp, monitoredProfileNotionProp)
			if err != nil {
				Logger.Error("RadarrWatchlistSync", "Failed to update notion watchlist", err)
				goto m
			}
		}
	}
	Logger.Info("RadarrWatchlistSync", "Status", "Radarr Watchlist Sync End")
	time.Sleep(t * time.Hour)
	goto start

}

func SonarrWatchlistSync(Logger *slog.Logger, N *notion.NotionClient, S *sonarr.SonarrClient, t time.Duration) {
	//fetch all series from sonarr along with download status
	//update notion watchlist
start:
	Logger.Info("SonarrWatchlistSync", "Status", "Sonarr Watchlist Sync Start")
	series, err := S.GetSeries(-1)
	Logger.Info("SonarrWatchlistSync", "Status", "Fetched series from sonarr library", "No of series fetched", len(series))
	if err != nil {
		Logger.Error("SonarrWatchlistSync", "Failed to fetch series from sonarr library", err)
		time.Sleep(10 * time.Second)
		goto start
	}
	for _, serie := range series {
	m:
		// query for each series in watchlist
		watchlistSeries, err := N.QueryDBImdb(serie.ImdbID)
		if err != nil {
			Logger.Error("SonarrWatchlistSync", "Failed to query series title from notion watchlist", err)
			time.Sleep(5 * time.Second)
			goto m
		}
		// if movie is not present in watchlist, skip
		if len(watchlistSeries.Results) == 0 {
			continue
		}
		//get rootpath and qualityprofile properties for notion db
		qualityProp, rootPathProp, err := N.GetNotionQualityAndRootProps(serie.QualityProfileID, serie.RootFolderPath, "TV Series")
		if err != nil {
			Logger.Error("SonarrWatchlistSync", "Failed to fetch notion DB property", err)
			goto m
		}
		if serie.Statistics.PercentOfEpisodes == 100 {
			err = N.UpdateDownloadStatus(watchlistSeries.Results[0].Pgid, false, "Downloaded", qualityProp, rootPathProp, "")
			if err != nil {
				Logger.Error("SonarrWatchlistSync", "Failed to update notion watchlist", err)
				goto m
			}
		} else {
			//check for queue status
			queueStatus, err := S.GetQueueDetails(serie.ID)
			if err != nil {
				Logger.Error("SonarrWatchlistSync", "Failed to fetch queue details from sonarr", err)
				goto m
			}
			err = N.UpdateDownloadStatus(watchlistSeries.Results[0].Pgid, false, queueStatus, qualityProp, rootPathProp, "")
			if err != nil {
				Logger.Error("SonarrWatchlistSync", "Failed to update notion watchlist", err)
				goto m
			}
		}
	}
	Logger.Info("SonarrWatchlistSync", "Status", "Sonarr Watchlist Sync End")
	time.Sleep(t * time.Hour)
	goto start
}
