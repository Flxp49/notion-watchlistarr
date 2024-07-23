package routine

import (
	"log/slog"
	"time"

	"github.com/flxp49/notion-watchlistarr/internal/constant"
	"github.com/flxp49/notion-watchlistarr/internal/notion"
	"github.com/flxp49/notion-watchlistarr/internal/radarr"
	"github.com/flxp49/notion-watchlistarr/internal/sonarr"
	"github.com/flxp49/notion-watchlistarr/internal/util"
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
		movieLookupInfo, err := R.LookupMovie(m.Properties.Imdbid.Rich_text[0].Plain_text)
		if err != nil {
			Logger.Error("RadarrSync", "Failed to fetch lookup movie via radarr by imdbid", m.Properties.Imdbid.Rich_text[0].Plain_text, "Error", err)
			N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
			continue
		}
		//check if movie exists or not
		movieLibraryData, err := R.GetMovie(movieLookupInfo.TmdbID)
		if err != nil {
			Logger.Error("RadarrSync", "Could not get title details from Radarr", err)
			N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
			continue
		}
		if len(movieLibraryData) != 0 {
			// movie exists
			//get rootpath and qualityprofile properties for notion db
			qualityProp, rootPathProp, err := N.GetNotionQualityAndRootProps(movieLibraryData[0].QualityProfileID, movieLibraryData[0].RootFolderPath, "Movie")
			if err != nil {
				Logger.Error("RadarrSync", "Failed to fetch notion DB qualityprofile and rootpath property for downloaded title", err)
				N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
				continue
			}
			monitoredProfile := constant.MovieOnly
			if movieLibraryData[0].Collection.TmdbID != 0 {
				collectionMonitored, err := R.GetCollection(movieLibraryData[0].Collection.TmdbID)
				if err != nil {
					Logger.Error("RadarrSync", "Failed to get collection details", err)
					N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
					continue
				}
				if collectionMonitored {
					monitoredProfile = constant.MovieAndCollection
				}
			}
			monitoredProfileNotionProp, _ := N.GetNotionMonitorProp(monitoredProfile, "Movie")
			if movieLibraryData[0].HasFile {
				err = N.UpdateDownloadStatus(m.Pgid, false, "Downloaded", qualityProp, rootPathProp, monitoredProfileNotionProp)
				if err != nil {
					Logger.Error("RadarrSync", "Failed to update download status in notion watchlist", err)
					N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
					continue
				}
			} else {
				//check for queue status
				queueStatus, err := R.GetQueueDetails(movieLibraryData[0].ID)
				if err != nil {
					Logger.Error("RadarrSync", "Failed to fetch queue details from radarr", err)
					N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
					continue
				}
				err = N.UpdateDownloadStatus(m.Pgid, false, queueStatus, qualityProp, rootPathProp, monitoredProfileNotionProp)
				if err != nil {
					Logger.Error("RadarrSync", "Failed to update notion watchlist", err)
					N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
					continue
				}
			}
			continue
		}
		// movie doesnt exist so lets add it
		// set monitor property
		if m.Properties.MonitorProfile.Select.Name == "" {
			monitorProfile, err := N.GetNotionMonitorProp(R.DefaultMonitorProfile, "Movie")
			if err != nil {
				Logger.Error("RadarrSync", "Could not get notion monitor property for monitor profile", err)
				N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
				continue
			}
			m.Properties.MonitorProfile.Select.Name = monitorProfile
		}
		//get rootpath and qualityprofile properties for notion db
		qualityProp, rootPathProp, err := N.GetNotionQualityAndRootProps(R.DefaultQualityProfile, R.DefaultRootPath, "Movie")
		if err != nil {
			Logger.Error("RadarrSync", "Failed to fetch notion DB qualityprofile and rootpath property", err)
			N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
			continue
		}
		// set root folder property
		if m.Properties.RootFolder.Select.Name == "" {
			m.Properties.RootFolder.Select.Name = rootPathProp
		}
		// set qualty profile property
		if m.Properties.QualityProfile.Select.Name == "" {
			m.Properties.QualityProfile.Select.Name = qualityProp
		}
		Logger.Info("RadarrSync", "Status", "Adding Title", "Title", movieLookupInfo.Title)
		err = R.AddMovie(movieLookupInfo, N.Qpid[m.Properties.QualityProfile.Select.Name], N.Rpid[m.Properties.RootFolder.Select.Name], true, true, notion.MonitorProfiles[m.Properties.MonitorProfile.Select.Name])
		if err != nil {
			Logger.Error("RadarrSync", "Failed to add title for download", err)
			N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
		}
	}
	time.Sleep(t * time.Second)
	goto start
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
		seriesLookupInfo, err := S.LookupSeries("imdb", m.Properties.Imdbid.Rich_text[0].Plain_text)
		if err != nil {
			Logger.Warn("SonarrSync", "Failed to lookup series via sonarr by imdbid", m.Properties.Imdbid.Rich_text[0].Plain_text, "Error", err)
			// use secondary search
			fallbackTvdb, err := util.GetSeriesByRemoteID(m.Properties.Imdbid.Rich_text[0].Plain_text)
			if err != nil {
				Logger.Error("SonarrSync", "Failed to fetch tvdbid of imdbid", m.Properties.Imdbid.Rich_text[0].Plain_text, "Error", err)
				N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
				continue
			}
			seriesLookupInfo, err = S.LookupSeries("tvdb", fallbackTvdb.Series.Seriesid)
			if err != nil {
				Logger.Error("SonarrSync", "Failed to lookup series by tvdbid", m.Properties.Imdbid.Rich_text[0].Plain_text, "Error", err)
				N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
				continue
			}
		}
		//check if series exists or not
		seriesLibraryData, err := S.GetSeries(seriesLookupInfo.TvdbID)
		if err != nil {
			Logger.Error("SonarrSync", "Could not get title details from Sonarr", err)
			N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
			continue
		}
		if len(seriesLibraryData) != 0 {
			// series exists
			// check if downloaded or not
			//? make a put request to update the movie?
			qualityProp, rootPathProp, err := N.GetNotionQualityAndRootProps(seriesLibraryData[0].QualityProfileID, seriesLibraryData[0].RootFolderPath, "TV Series")
			if err != nil {
				Logger.Error("SonarrSync", "Failed to fetch notion DB qualityprofile and rootpath property for downloaded title", err)
				N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
				continue
			}
			if seriesLibraryData[0].Statistics.PercentOfEpisodes == 100 {
				err = N.UpdateDownloadStatus(m.Pgid, false, "Downloaded", qualityProp, rootPathProp, "")
				if err != nil {
					Logger.Error("SonarrSync", "Failed to update download status in notion watchlist", err)
					N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
					continue
				}
			} else {
				//check for download queue
				queueStatus, err := S.GetQueueDetails(seriesLibraryData[0].ID)
				if err != nil {
					Logger.Error("SonarrSync", "Failed to fetch queue details from sonarr", err)
					N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
					continue
				}
				err = N.UpdateDownloadStatus(m.Pgid, false, queueStatus, qualityProp, rootPathProp, "")
				if err != nil {
					Logger.Error("RadarrSync", "Failed to update notion watchlist", err)
					N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
					continue
				}
			}
			continue
		}
		// series doesnt exist so lets add it
		// set monitor property
		if m.Properties.MonitorProfile.Select.Name == "" {
			monitorProfile, err := N.GetNotionMonitorProp(S.DefaultMonitorProfile, "TV Series")
			if err != nil {
				Logger.Error("SonarrSync", "Could not get notion monitor property for monitor profile", err)
				N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
				continue
			}
			m.Properties.MonitorProfile.Select.Name = monitorProfile
		}
		//get rootpath and qualityprofile properties for notion db
		qualityProp, rootPathProp, err := N.GetNotionQualityAndRootProps(S.DefaultQualityProfile, S.DefaultRootPath, "TV Series")
		if err != nil {
			Logger.Error("SonarrSync", "Failed to fetch notion DB qualityprofile and rootpath property", err)
			N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
			continue
		}
		// set root folder property
		if m.Properties.RootFolder.Select.Name == "" {
			m.Properties.RootFolder.Select.Name = rootPathProp
		}
		// set quality profile property
		if m.Properties.QualityProfile.Select.Name == "" {
			m.Properties.QualityProfile.Select.Name = qualityProp
		}
		Logger.Info("SonarrSync", "Status", "Adding Title", "Title", seriesLookupInfo.Title)
		err = S.AddSeries(seriesLookupInfo, N.Qpid[m.Properties.QualityProfile.Select.Name], N.Rpid[m.Properties.RootFolder.Select.Name], true, true, true, notion.MonitorProfiles[m.Properties.MonitorProfile.Select.Name])
		if err != nil {
			Logger.Error("SonarrSync", "Failed to add title for download", err)
			N.UpdateDownloadStatus(m.Pgid, false, "Error", "", "", "")
		}
	}
	time.Sleep(t * time.Second)
	goto start
}
