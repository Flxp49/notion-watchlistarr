package api

import (
	"io"
	"net/http"

	"github.com/flxp49/notion-watchlist-radarr-sonarr/util"
)

type MovieInfo struct {
	Movie struct {
		Id     int `json:"id"`
		TmdbId int `json:"tmdbId"`
	} `json:"movie"`
	EventType    string `json:"eventType"`    //allowed values MovieAdded|Grab|Download|MovieDelete
	DeletedFiles bool   `json:"deletedFiles"` //only present when EventType: "MovieDelete"
}

type SeriesInfo struct {
	Series struct {
		Id     int `json:"id"`
		TvdbId int `json:"tvdbId"`
	} `json:"series"`
	EventType    string `json:"eventType"`    //allowed values SeriesAdd|Grab|Download|SeriesDelete
	DeletedFiles bool   `json:"deletedFiles"` //only present when EventType: "SeriesDelete"
}

func (s *Server) radarrHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	r.Body.Close()
	var movieData MovieInfo
	err := util.ParseJson(body, &movieData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.Logger.Error("Radarr Webhook", "body", body, "error", err)
		return
	}
	s.Logger.Info("Radarr Webhook", "data", movieData)
	// Check if title exists in the watchlist
	page, err := s.N.QueryDBTmdb(movieData.Movie.TmdbId)
	if err != nil {
		s.Logger.Error("Radarr Webhook", "error", err)
		return
	}
	if len(page.Results) == 0 {
		return
	}

	//get movie file details
	movie, err := s.R.GetMovie(movieData.Movie.TmdbId)
	if err != nil {
		s.Logger.Error("Radarr Webhook", "error", err)
		return
	}
	//get rootpath and qualityprofile properties for notion db
	movieQualityProp, err := s.N.GetNotionQualityProfileProp(movie[0].QualityProfileId)
	if err != nil {
		s.Logger.Error("Radarr Webhook", "Failed to fetch notion DB movie quality profile property value", err)
		return
	}
	rootPathProp, err := s.N.GetNotionRootPathProp(movie[0].RootFolderPath)
	if err != nil {
		s.Logger.Error("Radarr Webhook", "Failed to fetch notion DB rootfolder property value", err)
		return
	}
	switch movieData.EventType {
	case "MovieAdded":
		//check if movie was imported manually (file already exists)
		if movie[0].HasFile {
			err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, true, "Downloaded", movieQualityProp, rootPathProp)
		} else {
			err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, true, "Queued", movieQualityProp, rootPathProp)
		}
		if err != nil {
			s.Logger.Error("Radarr Webhook", "Failed to update download status in watchlist", err)
		}

	case "Grab":
		err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, true, "Downloading", movieQualityProp, rootPathProp)
		if err != nil {
			s.Logger.Error("Radarr Webhook", "Failed to update download status in watchlist", err)
		}
	case "Download":
		err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, true, "Downloaded", movieQualityProp, rootPathProp)
		if err != nil {
			s.Logger.Error("Radarr Webhook", "Failed to update download status in watchlist", err)
		}
	case "MovieDelete":
		if !movieData.DeletedFiles {
			return
		}
		err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, false, "Not Downloaded", "", "")
		if err != nil {
			s.Logger.Error("Radarr Webhook", "Failed to update download status in watchlist", err)
		}
	default:
		s.Logger.Error("Radarr Webhook Error", "error", "EventType not valid in payload")
	}
}

func (s *Server) sonarrHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	r.Body.Close()
	var seriesData SeriesInfo
	err := util.ParseJson(body, &seriesData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.Logger.Error("Sonarr Webhook", "body", body, "error", err)
		return
	}
	// Check if title exists in the watchlist
	// for this we get tvdbid from sonarr but we need imdbid
	series, err := s.S.GetSeries(seriesData.Series.TvdbId)
	if err != nil {
		s.Logger.Error("Sonarr Webhook", "body", body, "error", err)
		return
	}
	// check if title exists in watchlist db
	page, err := s.N.QueryDBImdb(series[0].ImdbId)
	if err != nil {
		s.Logger.Error("Sonarr Webhook", "error", err)
		return
	}
	if len(page.Results) == 0 {
		return
	}

	//get rootpath and qualityprofile properties for notion db
	movieQualityProp, err := s.N.GetNotionQualityProfileProp(series[0].QualityProfileId)
	if err != nil {
		s.Logger.Error("Sonarr Webhook", "Failed to fetch notion DB movie quality profile property value", err)
		return
	}
	rootPathProp, err := s.N.GetNotionRootPathProp(series[0].RootFolderPath)
	if err != nil {
		s.Logger.Error("Sonarr Webhook", "Failed to fetch notion DB rootfolder property value", err)
		return
	}
	switch seriesData.EventType {
	case "SeriesAdd":
		//check if series was imported manually (file already exists)
		if series[0].Statistics.PercentOfEpisodes == 100 {
			err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, true, "Downloaded", movieQualityProp, rootPathProp)
		} else {
			err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, true, "Queued", movieQualityProp, rootPathProp)
		}
		if err != nil {
			s.Logger.Error("Sonarr Webhook", "Failed to update download status in watchlist", err)
		}
	case "Grab":
		err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, true, "Downloading", movieQualityProp, rootPathProp)
		if err != nil {
			s.Logger.Error("Sonarr Webhook", "Failed to update download status in watchlist", err)
		}
	case "Download":
		// check if all episodes were downloaded or not
		if series[0].Statistics.PercentOfEpisodes == 100 {
			err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, true, "Downloaded", movieQualityProp, rootPathProp)
		} else {
			err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, true, "Downloading", movieQualityProp, rootPathProp)
		}
		if err != nil {
			s.Logger.Error("Sonarr Webhook", "Failed to update download status in watchlist", err)
		}
	case "SeriesDelete":
		if !seriesData.DeletedFiles {
			return
		}
		err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, false, "Not Downloaded", "", "")
		if err != nil {
			s.Logger.Error("Sonarr Webhook", "Failed to update download status in watchlist", err)
		}
	default:
		s.Logger.Error("Sonarr Webhook", "error", "EventType not valid in payload")
	}

}

func (s *Server) incorrectReqHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}
