package app

import (
	"errors"

	"github.com/flxp49/notion-watchlistarr/internal/constant"
	"github.com/flxp49/notion-watchlistarr/internal/notion"
	"github.com/flxp49/notion-watchlistarr/internal/radarr"
)

type RadarrMedia struct {
	N *notion.NotionClient
	R *radarr.RadarrClient
}

func NewRadarrMedia(N *notion.NotionClient, R *radarr.RadarrClient) *RadarrMedia {
	return &RadarrMedia{N: N, R: R}
}

func (radarrMedia RadarrMedia) PollTitles() (notion.QueryDBResponse, error) {
	results, err := radarrMedia.N.QueryDB(constant.MediaTypeMovie)
	if err != nil {
		return notion.QueryDBResponse{}, err
	}
	return results, nil
}

func (radarrMedia RadarrMedia) FetchRadarrLibrary() ([]radarr.GetMovieResponse, error) {
	radarrMovies, err := radarrMedia.R.GetMovie(-1)
	if err != nil {
		return []radarr.GetMovieResponse{}, err
	}
	return radarrMovies, nil
}

func (radarrMedia RadarrMedia) QueryTitle(radarrMovie radarr.GetMovieResponse) (notion.QueryDBIdResponse, error) {
	watchlistMovie, err := radarrMedia.N.QueryDBImdb(radarrMovie.ImdbID)
	if err != nil {
		return notion.QueryDBIdResponse{}, err
	}
	return watchlistMovie, nil
}

func (radarrMedia RadarrMedia) ProcessTitles(notionPage notion.Result) (radarr.MovieLookupResponse, []radarr.GetMovieResponse, error) {
	movieLookupInfo, err := radarrMedia.R.LookupMovie(notionPage.Properties.Imdbid.Rich_text[0].Plain_text)
	if err != nil {
		return radarr.MovieLookupResponse{}, nil, err
	}
	//check if movie exists or not
	LibraryData, err := radarrMedia.R.GetMovie(movieLookupInfo.TmdbID)
	if err != nil {
		return radarr.MovieLookupResponse{}, nil, err
	}
	return movieLookupInfo, LibraryData, nil
}

func (radarrMedia RadarrMedia) AddTitle(LookupData radarr.MovieLookupResponse, notionPage notion.Result) error {
	// set monitor property
	if notionPage.Properties.MonitorProfile.Select.Name == "" {
		monitorProfile, err := radarrMedia.N.GetNotionMonitorProp(radarrMedia.R.DefaultMonitorProfile, constant.MediaTypeMovie)
		if err != nil {
			return errors.Join(errors.New("failed to get monitor profile notion property"), err)
		}
		notionPage.Properties.MonitorProfile.Select.Name = monitorProfile
	}
	//get rootpath and qualityprofile properties for notion db
	qualityProp, rootPathProp, err := radarrMedia.N.GetNotionQualityAndRootProps(radarrMedia.R.DefaultQualityProfile, radarrMedia.R.DefaultRootPath, constant.MediaTypeMovie)
	if err != nil {
		return errors.Join(errors.New("failed to get quality and root path profile notion property"), err)
	}
	// set root folder property
	if notionPage.Properties.RootFolder.Select.Name == "" {
		notionPage.Properties.RootFolder.Select.Name = rootPathProp
	}
	// set qualty profile property
	if notionPage.Properties.QualityProfile.Select.Name == "" {
		notionPage.Properties.QualityProfile.Select.Name = qualityProp
	}
	err = radarrMedia.R.AddMovie(LookupData, radarrMedia.N.Qpid[notionPage.Properties.QualityProfile.Select.Name], radarrMedia.N.Rpid[notionPage.Properties.RootFolder.Select.Name], true, true, notion.MonitorProfiles[notionPage.Properties.MonitorProfile.Select.Name])
	if err != nil {
		return errors.Join(errors.New("failed to add movie to radarr"), err)
	}
	return nil
}

func (radarrMedia RadarrMedia) HandleExistingTitle(LibraryData []radarr.GetMovieResponse, notionPage notion.Result) error {
	//get rootpath and qualityprofile properties for notion db
	qualityProp, rootPathProp, err := radarrMedia.N.GetNotionQualityAndRootProps(LibraryData[0].QualityProfileID, LibraryData[0].RootFolderPath, constant.MediaTypeMovie)
	if err != nil {
		return err
	}
	monitoredProfile, err := radarrMedia.getMovieMonitorProfile(LibraryData[0].Collection.TmdbID)
	if err != nil {
		return err
	}
	monitoredProfileNotionProp, _ := radarrMedia.N.GetNotionMonitorProp(monitoredProfile, constant.MediaTypeMovie)
	if LibraryData[0].HasFile {
		radarrMedia.N.UpdateDownloadStatus(constant.MediaTypeMovie, notionPage.Pgid, false, constant.MediaStatusDownloaded, qualityProp, rootPathProp, monitoredProfileNotionProp)
		return nil
	}
	//check for queue status
	queueStatus, err := radarrMedia.R.GetQueueDetails(LibraryData[0].ID)
	if err != nil {
		return errors.Join(errors.New("failed to get queue details in radarr"), err)
	}
	if queueStatus {
		radarrMedia.N.UpdateDownloadStatus(constant.MediaTypeMovie, notionPage.Pgid, false, constant.MediaStatusDownloading, qualityProp, rootPathProp, monitoredProfileNotionProp)
		return nil
	}
	//trigger movie search in Radarr
	err = radarrMedia.R.MovieSearchCommand(LibraryData[0].ID)
	if err != nil {
		return errors.Join(errors.New("failed to trigger movie search command in radarr"), err)
	}
	radarrMedia.N.UpdateDownloadStatus(constant.MediaTypeMovie, notionPage.Pgid, false, constant.MediaStatusQueued, qualityProp, rootPathProp, monitoredProfileNotionProp)
	return nil
}

func (radarrMedia RadarrMedia) ProcessLibraryTitle(watchlistMovie notion.QueryDBIdResponse, radarrMovie radarr.GetMovieResponse) error {
	monitoredProfile, err := radarrMedia.getMovieMonitorProfile(radarrMovie.Collection.TmdbID)
	if err != nil {
		return err
	}
	monitoredProfileNotionProp, _ := radarrMedia.N.GetNotionMonitorProp(monitoredProfile, constant.MediaTypeMovie)
	//get rootpath and qualityprofile properties for notion db
	qualityProp, rootPathProp, err := radarrMedia.N.GetNotionQualityAndRootProps(radarrMovie.QualityProfileID, radarrMovie.RootFolderPath, constant.MediaTypeMovie)
	if err != nil {
		return errors.Join(errors.New("failed to get quality and root path profile notion property"), err)
	}
	if radarrMovie.HasFile {
		radarrMedia.N.UpdateDownloadStatus(constant.MediaTypeMovie, watchlistMovie.Results[0].Pgid, false, constant.MediaStatusDownloaded, qualityProp, rootPathProp, monitoredProfileNotionProp)
		return nil
	}
	//check for queue status
	queueStatus, err := radarrMedia.R.GetQueueDetails(radarrMovie.ID)
	if err != nil {
		return errors.Join(errors.New("failed to get queue details in radarr"), err)
	}
	if queueStatus {
		radarrMedia.N.UpdateDownloadStatus(constant.MediaTypeMovie, watchlistMovie.Results[0].Pgid, false, constant.MediaStatusDownloading, qualityProp, rootPathProp, monitoredProfileNotionProp)
		return nil
	}
	radarrMedia.N.UpdateDownloadStatus(constant.MediaTypeMovie, watchlistMovie.Results[0].Pgid, false, constant.MediaStatusNotDownloaded, qualityProp, rootPathProp, monitoredProfileNotionProp)
	return nil
}

func (radarrMedia RadarrMedia) getMovieMonitorProfile(collectionTmdbid int) (string, error) {
	monitoredProfile := constant.MovieOnly
	if collectionTmdbid != 0 {
		collectionMonitored, err := radarrMedia.R.GetCollection(collectionTmdbid)
		if err != nil {
			return "", errors.Join(errors.New("failed to get movie collection"), err)
		}
		if collectionMonitored {
			monitoredProfile = constant.MovieAndCollection
		}
	}
	return monitoredProfile, nil
}
