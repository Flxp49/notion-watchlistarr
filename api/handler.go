package api

import (
	"io"
	"net/http"

	"github.com/flxp49/notion-watchlist-radarr-sonarr/util"
)

type MovieInfo struct {
	Movie struct {
		Id     int    `json:"id"`
		ImdbId string `json:"imdbId"`
		TmdbId int    `json:"tmdbId"`
	} `json:"movie"`
	EventType    string `json:"eventType"`    //allowed values MovieAdded|Grab|Download|MovieDelete
	DeletedFiles bool   `json:"deletedFiles"` //only present when EventType: "MovieDelete"
}

func (s *Server) radarrHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	r.Body.Close()
	var movieData MovieInfo
	err := util.ParseJson(body, &movieData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.Logger.Error("RadarrWebhook", "body", body, "error", err)
		return
	}
	s.Logger.Info("RadarrWebhook", "data", movieData)
	// Check if title exists in the watchlist
	page, err := s.N.QueryDBImdb(movieData.Movie.ImdbId)
	if err != nil {
		s.Logger.Error("RadarrWebhook", "error", err)
		return
	}

	if len(page.Results) == 0 {
		return
	}

	if movieData.EventType == "MovieDelete" {
		if !movieData.DeletedFiles {
			return
		}
		err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, false, "Not Downloaded", "", "")
		if err != nil {
			s.Logger.Error("RadarrWebhook", "Failed to update download status in watchlist", err)
		}
		return
	}
	if movieData.EventType == "Test" {
		return
	}
	//get movie file details
	movie, err := s.R.GetMovie(movieData.Movie.TmdbId)
	if err != nil {
		s.Logger.Error("RadarrWebhook", "error", err)
		return
	}
	//get rootpath and qualityprofile properties for notion db
	movieQualityProp, rootPathProp, err := s.N.GetNotionQualityAndRootProps(movie[0].QualityProfileId, movie[0].RootFolderPath, "Movie")
	if err != nil {
		s.Logger.Error("RadarrWebhook", "Failed to fetch notion DB property", err)
		return
	}
	switch movieData.EventType {
	case "MovieAdded":
		//check if movie was imported manually (file already exists)
		if movie[0].HasFile {
			err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, false, "Downloaded", movieQualityProp, rootPathProp)
		} else {
			err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, false, "Queued", movieQualityProp, rootPathProp)
		}
		if err != nil {
			s.Logger.Error("RadarrWebhook", "Failed to update download status in watchlist", err)
		}

	case "Grab":
		err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, false, "Downloading", movieQualityProp, rootPathProp)
		if err != nil {
			s.Logger.Error("RadarrWebhook", "Failed to update download status in watchlist", err)
		}
	case "Download":
		err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, false, "Downloaded", movieQualityProp, rootPathProp)
		if err != nil {
			s.Logger.Error("RadarrWebhook", "Failed to update download status in watchlist", err)
		}
	default:
		s.Logger.Error("RadarrWebhook", "error", "EventType not valid in payload")
	}

}

type SeriesInfo struct {
	Series struct {
		Id     int    `json:"id"`
		ImdbId string `json:"imdbId"`
		TvdbId int    `json:"tvdbId"`
	} `json:"series"`
	EventType    string `json:"eventType"`    //allowed values SeriesAdd|Grab|Download|SeriesDelete
	DeletedFiles bool   `json:"deletedFiles"` //only present when EventType: "SeriesDelete"
}

func (s *Server) sonarrHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	r.Body.Close()
	var seriesData SeriesInfo
	err := util.ParseJson(body, &seriesData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.Logger.Error("SonarrWebhook", "body", body, "error", err)
		return
	}
	s.Logger.Info("SonarrWebhook", "data", seriesData)
	// check if title exists in watchlist db
	page, err := s.N.QueryDBImdb(seriesData.Series.ImdbId)
	if err != nil {
		s.Logger.Error("SonarrWebhook", "error", err)
		return
	}
	if len(page.Results) == 0 {
		return
	}

	if seriesData.EventType == "SeriesDelete" {
		if !seriesData.DeletedFiles {
			return
		}
		err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, false, "Not Downloaded", "", "")
		if err != nil {
			s.Logger.Error("SonarrWebhook", "Failed to update download status in watchlist", err)
		}
		return
	}
	if seriesData.EventType == "Test" {
		return
	}

	series, err := s.S.GetSeries(seriesData.Series.TvdbId)
	if err != nil {
		s.Logger.Error("SonarrWebhook", "body", body, "error", err)
		return
	}
	//get rootpath and qualityprofile properties for notion db
	qualityProp, rootPathProp, err := s.N.GetNotionQualityAndRootProps(series[0].QualityProfileId, series[0].RootFolderPath, "TV Series")
	if err != nil {
		s.Logger.Error("SonarrWebhook", "Failed to fetch notion DB property", err)
		return
	}
	switch seriesData.EventType {
	case "SeriesAdd":
		//check if series was imported manually (file already exists)
		if series[0].Statistics.PercentOfEpisodes == 100 {
			err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, false, "Downloaded", qualityProp, rootPathProp)
		} else {
			err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, false, "Queued", qualityProp, rootPathProp)
		}
		if err != nil {
			s.Logger.Error("SonarrWebhook", "Failed to update download status in watchlist", err)
		}
	case "Grab":
		err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, false, "Downloading", qualityProp, rootPathProp)
		if err != nil {
			s.Logger.Error("SonarrWebhook", "Failed to update download status in watchlist", err)
		}
	case "Download":
		// check if all episodes were downloaded or not
		if series[0].Statistics.PercentOfEpisodes == 100 {
			err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, false, "Downloaded", qualityProp, rootPathProp)
		} else {
			err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, false, "Downloading", qualityProp, rootPathProp)
		}
		if err != nil {
			s.Logger.Error("SonarrWebhook", "Failed to update download status in watchlist", err)
		}
	default:
		s.Logger.Error("SonarrWebhook", "error", "EventType not valid in payload")
	}

}

func (s *Server) incorrectReqHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}
