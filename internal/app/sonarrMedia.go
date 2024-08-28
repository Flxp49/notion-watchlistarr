package app

import (
	"errors"

	"github.com/flxp49/notion-watchlistarr/internal/notion"
	"github.com/flxp49/notion-watchlistarr/internal/sonarr"
	"github.com/flxp49/notion-watchlistarr/internal/util"
)

type SonarrMedia struct {
	N *notion.NotionClient
	S *sonarr.SonarrClient
}

func NewSonarrMedia(N *notion.NotionClient, S *sonarr.SonarrClient) *SonarrMedia {
	return &SonarrMedia{N: N, S: S}
}

func (sonarrMedia SonarrMedia) PollTitles() (notion.QueryDBResponse, error) {
	results, err := sonarrMedia.N.QueryDB("TV Series")
	if err != nil {
		return notion.QueryDBResponse{}, err
	}
	return results, nil
}

func (sonarrMedia SonarrMedia) FetchSonarrLibrary() ([]sonarr.GetSeriesResponse, error) {
	sonarrSeries, err := sonarrMedia.S.GetSeries(-1)
	if err != nil {
		return []sonarr.GetSeriesResponse{}, err
	}
	return sonarrSeries, nil
}

func (sonarrMedia SonarrMedia) QueryTitle(sonarrSeries sonarr.GetSeriesResponse) (notion.QueryDBIdResponse, error) {
	watchlistSeries, err := sonarrMedia.N.QueryDBImdb(sonarrSeries.ImdbID)
	if err != nil {
		return notion.QueryDBIdResponse{}, err
	}
	return watchlistSeries, nil
}
func (sonarrMedia SonarrMedia) ProcessTitles(notionPage notion.Result) (sonarr.LookupSeriesResponse, []sonarr.GetSeriesResponse, error) {
	seriesLookupInfo, err := sonarrMedia.S.LookupSeries("imdb", notionPage.Properties.Imdbid.Rich_text[0].Plain_text)
	if err != nil {
		// use secondary search
		fallbackTvdb, err := util.GetSeriesByRemoteID(notionPage.Properties.Imdbid.Rich_text[0].Plain_text)
		if err != nil {
			return sonarr.LookupSeriesResponse{}, nil, err
		}
		seriesLookupInfo, err = sonarrMedia.S.LookupSeries("tvdb", fallbackTvdb.Series.Seriesid)
		if err != nil {
			return sonarr.LookupSeriesResponse{}, nil, err
		}
	}
	//check if series exists or not
	LibraryData, err := sonarrMedia.S.GetSeries(seriesLookupInfo.TvdbID)
	if err != nil {
		return sonarr.LookupSeriesResponse{}, nil, err
	}
	return seriesLookupInfo, LibraryData, nil
}

func (sonarrMedia SonarrMedia) AddTitle(LookupData sonarr.LookupSeriesResponse, notionPage notion.Result) error {
	// set monitor property
	if notionPage.Properties.MonitorProfile.Select.Name == "" {
		monitorProfile, err := sonarrMedia.N.GetNotionMonitorProp(sonarrMedia.S.DefaultMonitorProfile, "TV Series")
		if err != nil {
			return errors.Join(errors.New("failed to get monitor profile notion property"), err)
		}
		notionPage.Properties.MonitorProfile.Select.Name = monitorProfile
	}
	//get rootpath and qualityprofile properties for notion db
	qualityProp, rootPathProp, err := sonarrMedia.N.GetNotionQualityAndRootProps(sonarrMedia.S.DefaultQualityProfile, sonarrMedia.S.DefaultRootPath, "TV Series")
	if err != nil {
		return errors.Join(errors.New("failed to get quality and root path profile notion property"), err)
	}
	// set root folder property
	if notionPage.Properties.RootFolder.Select.Name == "" {
		notionPage.Properties.RootFolder.Select.Name = rootPathProp
	}
	// set quality profile property
	if notionPage.Properties.QualityProfile.Select.Name == "" {
		notionPage.Properties.QualityProfile.Select.Name = qualityProp
	}
	err = sonarrMedia.S.AddSeries(LookupData, sonarrMedia.N.Qpid[notionPage.Properties.QualityProfile.Select.Name], sonarrMedia.N.Rpid[notionPage.Properties.RootFolder.Select.Name], true, true, true, notion.MonitorProfiles[notionPage.Properties.MonitorProfile.Select.Name])
	if err != nil {
		return errors.Join(errors.New("failed to add movie to sonarr"), err)
	}
	return nil
}

func (sonarrMedia SonarrMedia) HandleExistingTitle(LibraryData []sonarr.GetSeriesResponse, notionPage notion.Result) error {
	qualityProp, rootPathProp, err := sonarrMedia.N.GetNotionQualityAndRootProps(LibraryData[0].QualityProfileID, LibraryData[0].RootFolderPath, "TV Series")
	if err != nil {
		return err
	}
	if LibraryData[0].Statistics.PercentOfEpisodes == 100 {
		sonarrMedia.N.UpdateDownloadStatus("series", notionPage.Pgid, false, "Downloaded", qualityProp, rootPathProp, "")
		return nil
	}
	//check for download queue
	queueStatus, err := sonarrMedia.S.GetQueueDetails(LibraryData[0].ID)
	if err != nil {
		return errors.Join(errors.New("failed to get queue details in sonarr"), err)
	}
	if queueStatus {
		sonarrMedia.N.UpdateDownloadStatus("series", notionPage.Pgid, false, "Downloading", qualityProp, rootPathProp, "")
		return nil
	}

	// trigger search for series
	err = sonarrMedia.S.SeriesSearchCommand(LibraryData[0].ID)
	if err != nil {
		return errors.Join(errors.New("failed to trigger series search command in sonarr"), err)
	}
	sonarrMedia.N.UpdateDownloadStatus("series", notionPage.Pgid, false, "Queued", qualityProp, rootPathProp, "")
	return nil
}

func (sonarrMedia SonarrMedia) ProcessLibraryTitle(watchlistSeries notion.QueryDBIdResponse, sonarrSeries sonarr.GetSeriesResponse) error {
	//get rootpath and qualityprofile properties for notion db
	qualityProp, rootPathProp, err := sonarrMedia.N.GetNotionQualityAndRootProps(sonarrSeries.QualityProfileID, sonarrSeries.RootFolderPath, "TV Series")
	if err != nil {
		return errors.Join(errors.New("failed to get quality and root path profile notion property"), err)
	}
	if sonarrSeries.Statistics.PercentOfEpisodes == 100 {
		sonarrMedia.N.UpdateDownloadStatus("series", watchlistSeries.Results[0].Pgid, false, "Downloaded", qualityProp, rootPathProp, "")
		return nil
	}
	//check for queue status
	queueStatus, err := sonarrMedia.S.GetQueueDetails(sonarrSeries.ID)
	if err != nil {
		return errors.Join(errors.New("failed to get queue details in radarr"), err)
	}
	if queueStatus {
		sonarrMedia.N.UpdateDownloadStatus("series", watchlistSeries.Results[0].Pgid, false, "Downloading", qualityProp, rootPathProp, "")
		return nil
	}
	sonarrMedia.N.UpdateDownloadStatus("series", watchlistSeries.Results[0].Pgid, false, "Not Downloaded", qualityProp, rootPathProp, "")

	return nil
}
