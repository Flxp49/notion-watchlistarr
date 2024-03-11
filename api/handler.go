package api

import (
	"io"
	"log"
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

func (s *Server) radarrHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	r.Body.Close()
	var movieData MovieInfo
	err := util.ParseJson(body, &movieData)
	if err != nil {
		w.WriteHeader(400)
		s.Logger.Error("Radarr Webhook Error", "body", body, "error", err)
		return
	}
	w.WriteHeader(200)
	s.Logger.Info("Radarr post request", "data", movieData)
	// Check if title exists in the watchlist
	page, err := s.N.QueryDBTmdb(movieData.Movie.TmdbId)
	if err != nil {
		s.Logger.Error("Radarr Webhook Error", "error", err)
		return
	}
	if len(page.Results) == 0 {
		return
	}

	//get movie file details
	movie, err := s.R.GetMovie(movieData.Movie.TmdbId)
	if err != nil {
		s.Logger.Error("Radarr Webhook Error", "error", err)
		return
	}
	switch movieData.EventType {
	case "MovieAdded":
		//get rootpath and qualityprofile properties for notion db
		movieQualityProp, err := s.N.GetNotionQualityProfileProp(movie[0].QualityProfileId)
		if err != nil {
			s.Logger.Error("Radarr Webhook Error: Failed to fetch notion DB movie quality profile property value", "error", err)
			return
		}
		rootPathProp, err := s.N.GetNotionRootPathProp(movie[0].RootFolderPath)
		if err != nil {
			s.Logger.Error("Radarr Webhook Error: Failed to fetch notion DB rootfolder property value", "error", err)
			return
		}
		if movie[0].HasFile {
			err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, true, "Downloaded", movieQualityProp, rootPathProp)
		} else {
			err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, true, "Queued", movieQualityProp, rootPathProp)
		}
		if err != nil {
			s.Logger.Error("Radarr Webhook Error: Failed to update download status in watchlist", "error", err)
		}

	case "Grab":
		movie, err := s.R.GetMovie(movieData.Movie.TmdbId)
		if err != nil {
			s.Logger.Error("Radarr Webhook Error", "error", err)
			return
		}
		movieQualityProp, err := s.N.GetNotionQualityProfileProp(movie[0].QualityProfileId)
		if err != nil {
			s.Logger.Error("Radarr Webhook Error: Failed to fetch notion DB movie quality profile property value", "error", err)
			return
		}
		rootPathProp, err := s.N.GetNotionRootPathProp(movie[0].RootFolderPath)
		if err != nil {
			s.Logger.Error("Radarr Webhook Error: Failed to fetch notion DB rootfolder property value", "error", err)
			return
		}
		err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, true, "Downloading", movieQualityProp, rootPathProp)
		if err != nil {
			s.Logger.Error("Radarr Webhook Error: Failed to update download status in watchlist", "error", err)
		}
	case "Download":
		movieQualityProp, err := s.N.GetNotionQualityProfileProp(movie[0].QualityProfileId)
		if err != nil {
			s.Logger.Error("Radarr Webhook Error: Failed to fetch notion DB movie quality profile property value", "error", err)
			return
		}
		rootPathProp, err := s.N.GetNotionRootPathProp(movie[0].RootFolderPath)
		if err != nil {
			s.Logger.Error("Radarr Webhook Error: Failed to fetch notion DB rootfolder property value", "error", err)
			return
		}
		err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, true, "Downloaded", movieQualityProp, rootPathProp)
		if err != nil {
			s.Logger.Error("Radarr Webhook Error: Failed to update download status in watchlist", "error", err)
		}
	case "MovieDelete":
		if !movieData.DeletedFiles {
			return
		}
		err = s.N.UpdateDownloadStatus(page.Results[0].Pgid, false, "Not Downloaded", "", "")
		if err != nil {
			s.Logger.Error("Radarr Webhook Error: Failed to update download status in watchlist", "error", err)
		}
	default:
		s.Logger.Error("Radarr Webhook Error", "error", "EventType not valid in payload")
	}
}

func (s *Server) sonarrHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	r.Body.Close()
	log.Println(string(body))
}

func (s *Server) incorrectReqHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(405)
}
