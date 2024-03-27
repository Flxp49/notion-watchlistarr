package sonarr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/flxp49/notion-watchlist-radarr-sonarr/util"
)

type SonarrClient struct {
	req                   *http.Request
	hostpath              string
	DefaultRootPath       string
	DefaultQualityProfile int
	DefaultMonitorProfile string
}

func (s *SonarrClient) performReq(method string, endpoint string, data []byte) (*http.Response, []byte, error) {
	s.req.Method = method
	s.req.URL, _ = url.Parse(s.hostpath + "/api/v3" + endpoint)
	if method == "POST" {
		s.req.Body = io.NopCloser(bytes.NewBuffer(data))
		s.req.ContentLength = int64(len(data))
	} else {
		s.req.Body = nil
		s.req.ContentLength = 0
	}
	resp, err := http.DefaultClient.Do(s.req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil || (resp.StatusCode < 200 || resp.StatusCode >= 300) {
		if err == nil {
			err = errors.New(string(body))
		}
		return nil, nil, &util.RequestError{StatusCode: resp.StatusCode, Err: err}
	}
	return resp, body, nil
}

// getRootFolder Response struct
type GetRootFolderResponse []struct {
	Path string `json:"path"`
}

// Fetches the rootfolder path set in Sonarr
func (s *SonarrClient) GetRootFolder() (GetRootFolderResponse, error) {
	_, body, err := s.performReq("GET", "/rootfolder", nil)
	if err != nil {
		return nil, err
	}
	var rf GetRootFolderResponse
	err = util.ParseJson(body, &rf)
	if err != nil {
		return nil, err
	}
	return rf, nil

}

// getQualityProfile response struct
type QualityProfileResponse []struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// Fetches the quality profiles
func (s *SonarrClient) GetQualityProfiles() (QualityProfileResponse, error) {
	_, body, err := s.performReq("GET", "/qualityprofile", nil)
	if err != nil {
		return nil, err
	}
	var qPR QualityProfileResponse
	err = util.ParseJson(body, &qPR)
	if err != nil {
		return nil, err
	}
	return qPR, nil

}

type LookupSeriesByImdbidResponse []struct {
	TvdbId int `json:"tvdbId"`
	// ImdbId string `json:"imdbId"`
}

// lookup series via Sonarr to get tvdbid
func (s *SonarrClient) LookupSeriesByImdbid(imdbId string) (LookupSeriesByImdbidResponse, error) {

	_, body, err := s.performReq("GET", fmt.Sprintf("/series/lookup?term=imdbId:%s", imdbId), nil)
	if err != nil {
		return nil, err
	}
	var lSBIR LookupSeriesByImdbidResponse
	err = util.ParseJson(body, &lSBIR)
	if err != nil {
		return nil, err
	}
	return lSBIR, nil
}

// Add the series to Sonarr
//
// monitor : "AllEpisodes" | "FutureEpisodes" | "MissingEpisodes" | "ExistingEpisodes" | "RecentEpisodes" | "PilotEpisode" | "FirstSeason" | "LastSeason" | "MonitorSpecials" | "UnmonitorSpecials" | "None"
func (s *SonarrClient) AddSeries(title string, qualityProfileId int, TvdbId int, rootFolderPath string, monitored bool, seasonFolder bool, SearchForMissingEpisodes bool, monitorProfile string) error {
	type addSeriesPayload struct {
		Title            string `json:"title"`
		QualityProfileId int    `json:"qualityProfileId"`
		TvdbId           int    `json:"tvdbId"`
		RootFolderPath   string `json:"rootFolderPath"`
		Monitored        bool   `json:"monitored"`
		SeasonFolder     bool   `json:"seasonFolder"`
		AddOptions       struct {
			SearchForMissingEpisodes bool `json:"searchForMissingEpisodes"`
			MonitorTypes             string
		} `json:"addOptions"`
	}
	payload := addSeriesPayload{Title: title, QualityProfileId: qualityProfileId, TvdbId: TvdbId, RootFolderPath: rootFolderPath, Monitored: monitored, SeasonFolder: seasonFolder, AddOptions: struct {
		SearchForMissingEpisodes bool `json:"searchForMissingEpisodes"`
		MonitorTypes             string
	}{SearchForMissingEpisodes: SearchForMissingEpisodes, MonitorTypes: monitorProfile}}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, _, err = s.performReq("POST", "/series", data)
	if err != nil {
		return err
	}
	return nil
}

// getMovie response struct
type GetSeriesResponse []struct {
	ImdbId           string `json:"imdbId"`
	SeriesID         int    `json:"id"`
	QualityProfileId int    `json:"qualityProfileId"`
	Monitored        bool   `json:"monitored"`
	RootFolderPath   string `json:"rootFolderPath"`
	Statistics       struct {
		PercentOfEpisodes int `json:"percentOfEpisodes"`
	} `json:"statistics"`
}

// Fetch series details in Sonarr
func (s *SonarrClient) GetSeries(tvdbId int) (GetSeriesResponse, error) {
	var query string
	if tvdbId == -1 {
		query = "/series"
	} else {
		query = fmt.Sprintf("/series?tvdbId=%d", tvdbId)
	}
	_, body, err := s.performReq("GET", query, nil)
	if err != nil {
		return nil, err
	}
	var gSR GetSeriesResponse
	err = util.ParseJson(body, &gSR)
	if err != nil {
		return nil, err
	}
	return gSR, nil
}

// Fetch movie download status
func (s *SonarrClient) GetQueueDetails(seriesId int) (util.GetQueueDetailsResponse, error) {
	_, body, err := s.performReq("GET", fmt.Sprintf("/queue?id=%d", seriesId), nil)
	if err != nil {
		return util.GetQueueDetailsResponse{}, err
	}
	var gDSR util.GetQueueDetailsResponse
	err = util.ParseJson(body, &gDSR)
	if err != nil {
		return util.GetQueueDetailsResponse{}, err
	}
	return gDSR, nil

}

// Sets the default profiles and fetches the quality, rootpath profiles from sonarr
func (s *SonarrClient) SonarrDefaults(sonarrDefaultRootPath string, sonarrDefaultQualityProfile string, sonarrDefaultMonitorProfile string, rpid map[string]string, qpid map[string]int) error {
	//set default monitor
	if sonarrDefaultMonitorProfile == "" {
		s.DefaultMonitorProfile = "AllEpisodes"
	} else {
		s.DefaultMonitorProfile = sonarrDefaultMonitorProfile
	}
	// Root path
	sonarrRootPaths, err := s.GetRootFolder()
	if len(sonarrRootPaths) == 0 || err != nil {
		return errors.Join(errors.New("failed to fetch sonarr root paths from sonarr"), err)
	}

	for _, s := range sonarrRootPaths {
		rpid["TV Series: "+s.Path] = s.Path
	}
	if sonarrDefaultRootPath == "" {
		s.DefaultRootPath = sonarrRootPaths[0].Path
	} else {
		s.DefaultRootPath = sonarrDefaultRootPath
	}
	// Quality Profiles
	sonarrQualityProfiles, err := s.GetQualityProfiles()
	if len(sonarrQualityProfiles) == 0 || err != nil {
		return errors.Join(errors.New("failed to fetch sonarr quality profiles from sonarr"), err)
	}

	for _, v := range sonarrQualityProfiles {
		qpid["TV Series: "+v.Name] = v.Id
	}
	if sonarrDefaultQualityProfile == "" {
		s.DefaultQualityProfile = sonarrQualityProfiles[0].Id
	} else {
		//check if user passed quality profile is valid or not
		profileId, exists := qpid["TV Series: "+sonarrDefaultQualityProfile]
		if !exists {
			return errors.New("wrong default quality profile passed")
		}
		s.DefaultQualityProfile = profileId
	}
	return nil
}

func InitSonarrClient(apikey string, hostpath string) *SonarrClient {
	s := &SonarrClient{hostpath: hostpath}
	s.req, _ = http.NewRequest("", "", nil)
	s.req.Header.Add("X-Api-Key", apikey)
	s.req.Header.Add("Content-Type", "application/json")
	return s
}
