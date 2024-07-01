package sonarr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/flxp49/notion-watchlistarr/internal/constant"
	"github.com/flxp49/notion-watchlistarr/internal/util"
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
	if method == http.MethodGet {
		s.req.Body = nil
		s.req.ContentLength = 0
	} else {
		s.req.Body = io.NopCloser(bytes.NewBuffer(data))
		s.req.ContentLength = int64(len(data))
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
	_, body, err := s.performReq(http.MethodGet, "/rootfolder", nil)
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
	_, body, err := s.performReq(http.MethodGet, "/qualityprofile", nil)
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

type LookupSeriesResponse struct {
	Title     string `json:"title"`
	SortTitle string `json:"sortTitle"`
	Status    string `json:"status"`
	Ended     bool   `json:"ended"`
	Overview  string `json:"overview"`
	Network   string `json:"network"`
	AirTime   string `json:"airTime"`
	Images    []struct {
		CoverType string `json:"coverType"`
		URL       string `json:"url"`
		RemoteURL string `json:"remoteUrl"`
	} `json:"images"`
	OriginalLanguage struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"originalLanguage"`
	RemotePoster string `json:"remotePoster"`
	Seasons      []struct {
		SeasonNumber int  `json:"seasonNumber"`
		Monitored    bool `json:"monitored"`
	} `json:"seasons"`
	Year              int           `json:"year"`
	QualityProfileID  int           `json:"qualityProfileId"`
	SeasonFolder      bool          `json:"seasonFolder"`
	Monitored         bool          `json:"monitored"`
	MonitorNewItems   string        `json:"monitorNewItems"`
	UseSceneNumbering bool          `json:"useSceneNumbering"`
	Runtime           int           `json:"runtime"`
	TvdbID            int           `json:"tvdbId"`
	TvRageID          int           `json:"tvRageId"`
	TvMazeID          int           `json:"tvMazeId"`
	FirstAired        time.Time     `json:"firstAired"`
	LastAired         time.Time     `json:"lastAired"`
	SeriesType        string        `json:"seriesType"`
	CleanTitle        string        `json:"cleanTitle"`
	ImdbID            string        `json:"imdbId"`
	TitleSlug         string        `json:"titleSlug"`
	Folder            string        `json:"folder"`
	Certification     string        `json:"certification"`
	Genres            []string      `json:"genres"`
	Tags              []interface{} `json:"tags"`
	Added             time.Time     `json:"added"`
	Ratings           struct {
		Votes int     `json:"votes"`
		Value float32 `json:"value"`
	} `json:"ratings"`
	Statistics struct {
		SeasonCount       int     `json:"seasonCount"`
		EpisodeFileCount  int     `json:"episodeFileCount"`
		EpisodeCount      int     `json:"episodeCount"`
		TotalEpisodeCount int     `json:"totalEpisodeCount"`
		SizeOnDisk        float32 `json:"sizeOnDisk"`
		PercentOfEpisodes float32 `json:"percentOfEpisodes"`
	} `json:"statistics"`
	LanguageProfileID int `json:"languageProfileId"`
}

// lookup series by imdbid or tvdbid via Sonarr to get series data
//
// idType : "imdb" || "tvdb"
func (s *SonarrClient) LookupSeries(idType string, id string) (LookupSeriesResponse, error) {
	if !(idType == "imdb" || idType == "tvdb") {
		return LookupSeriesResponse{}, errors.New(`idType to be either "imdb" || "tvdb"`)
	}
	_, body, err := s.performReq(http.MethodGet, fmt.Sprintf("/series/lookup?term=%s:%s", idType, id), nil)
	if err != nil {
		return LookupSeriesResponse{}, err
	}
	var lSBIR []LookupSeriesResponse
	err = util.ParseJson(body, &lSBIR)
	if err != nil || len(lSBIR) == 0 {
		if err == nil {
			err = errors.New("no title found via lookup")
		}
		return LookupSeriesResponse{}, err
	}
	return lSBIR[0], nil
}

// Add the series to Sonarr
//
// monitor : "All" | "Future" | "Missing" | "Existing" | "Recent" | "Pilot" | "FirstSeason" | "LastSeason" | "MonitorSpecials" | "UnmonitorSpecials"
func (s *SonarrClient) AddSeries(seriesLookupData LookupSeriesResponse, qualityProfileId int, rootFolderPath string, monitored bool, seasonFolder bool, SearchForMissingEpisodes bool, monitorProfile string) error {
	type addSeriesPayload struct {
		LookupSeriesResponse
		RootFolderPath string `json:"rootFolderPath"`
		AddOptions     struct {
			SearchForMissingEpisodes bool   `json:"searchForMissingEpisodes"`
			MonitorProfile           string `json:"monitor"`
		} `json:"addOptions"`
	}
	payload := addSeriesPayload{LookupSeriesResponse: seriesLookupData, AddOptions: struct {
		SearchForMissingEpisodes bool   `json:"searchForMissingEpisodes"`
		MonitorProfile           string `json:"monitor"`
	}{SearchForMissingEpisodes: SearchForMissingEpisodes, MonitorProfile: monitorProfile}}
	payload.QualityProfileID = qualityProfileId
	payload.Monitored = monitored
	payload.RootFolderPath = rootFolderPath
	payload.SeasonFolder = seasonFolder
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, _, err = s.performReq(http.MethodPost, "/series", data)
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
	Path             string `json:"path"`
	RootFolderPath   string `json:"rootFolderPath"`
	Statistics       struct {
		PercentOfEpisodes float32 `json:"percentOfEpisodes"`
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
	_, body, err := s.performReq(http.MethodGet, query, nil)
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
	_, body, err := s.performReq(http.MethodGet, fmt.Sprintf("/queue?id=%d", seriesId), nil)
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
		s.DefaultMonitorProfile = constant.AllEpisodes
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
		//check if user passed root path is valid or not
		for _, path := range sonarrRootPaths {
			if util.CheckSamePath(path.Path, sonarrDefaultRootPath) {
				s.DefaultRootPath = sonarrDefaultRootPath
				break
			}
		}
		if s.DefaultRootPath == "" {
			return errors.New("invalid sonarr default root path passed")
		}
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
	s := &SonarrClient{hostpath: hostpath, DefaultRootPath: "", DefaultMonitorProfile: ""}
	s.req, _ = http.NewRequest("", "", nil)
	s.req.Header.Add("X-Api-Key", apikey)
	s.req.Header.Add("Content-Type", "application/json")
	return s
}
