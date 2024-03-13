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
	DefaultQualityProfile string
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
		return nil, nil, err
	}
	return resp, body, nil
}

// getRootFolder Response struct
type getRootFolderResponse []struct {
	Path string `json:"path"`
}

// Fetches the rootfolder path set in Sonarr
func (s *SonarrClient) GetRootFolder() (getRootFolderResponse, error) {
	_, body, err := s.performReq("GET", "/rootfolder", nil)
	if err != nil {
		return nil, err
	}
	var rf getRootFolderResponse
	err = util.ParseJson(body, &rf)
	if err != nil {
		return nil, err
	}
	return rf, nil

}

// getQualityProfile response struct
type qualityProfileResponse []struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// Fetches the quality profiles
func (s *SonarrClient) GetQualityProfiles() (qualityProfileResponse, error) {
	_, body, err := s.performReq("GET", "/qualityprofile", nil)
	if err != nil {
		return nil, err
	}
	var qPR qualityProfileResponse
	err = util.ParseJson(body, &qPR)
	if err != nil {
		return nil, err
	}
	return qPR, nil

}

type LookupSeriesByTvdbidResponse []struct {
	TvdbId int `json:"tvdbId"`
	// ImdbId string `json:"imdbId"`
}

// lookup series via Sonarr to get tvdbid
func (s *SonarrClient) LookupSeriesByTmdbid(tmdbId int) (LookupSeriesByTvdbidResponse, error) {

	_, body, err := s.performReq("GET", fmt.Sprintf("/series/lookup?term=tmdbId:%d", tmdbId), nil)
	if err != nil {
		return nil, err
	}
	var lSBTR LookupSeriesByTvdbidResponse
	err = util.ParseJson(body, &lSBTR)
	if err != nil {
		return nil, err
	}
	return lSBTR, nil
}

type addSeriesPayload struct {
	Title            string `json:"title"`
	QualityProfileId int    `json:"qualityProfileId"`
	TvdbId           int    `json:"tvdbId"`
	RootFolderPath   string `json:"rootFolderPath"`
	Monitored        bool   `json:"monitored"`
	SeasonFolder     bool   `json:"seasonFolder"`
	AddOptions       struct {
		SearchForMissingEpisodes bool   `json:"searchForMissingEpisodes"`
		Monitor                  string `json:"monitor"`
	} `json:"addOptions"`
}

// Add the series to Sonarr
//
// monitor : "AllEpisodes" | "FutureEpisodes" | "MissingEpisodes" | "ExistingEpisodes" | "RecentEpisodes" | "PilotEpisode" | "FirstSeason" | "LastSeason" | "MonitorSpecials" | "UnmonitorSpecials" | "None"
func (s *SonarrClient) AddSeries(title string, qualityProfileId int, TvdbId int, rootFolderPath string, monitored bool, seasonFolder bool, SearchForMissingEpisodes bool, monitor string) error {
	payload := addSeriesPayload{Title: title, QualityProfileId: qualityProfileId, TvdbId: TvdbId, RootFolderPath: rootFolderPath, Monitored: monitored, SeasonFolder: seasonFolder, AddOptions: struct {
		SearchForMissingEpisodes bool   `json:"searchForMissingEpisodes"`
		Monitor                  string `json:"monitor"`
	}{SearchForMissingEpisodes: SearchForMissingEpisodes, Monitor: monitor}}

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
type getSeriesResponse []struct {
	// HasFile          bool   `json:"hasFile"`
	QualityProfileId int    `json:"qualityProfileId"`
	Monitored        bool   `json:"monitored"`
	RootFolderPath   string `json:"rootFolderPath"`
	Statistics       struct {
		PercentOfEpisodes int `json:"percentOfEpisodes"`
	} `json:"statistics"`
}

// Fetch series details in Sonarr
func (s *SonarrClient) GetSeries(tvdbId int) (getSeriesResponse, error) {
	_, body, err := s.performReq("GET", fmt.Sprintf("/series?tvdbId=%d", tvdbId), nil)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(body))
	var gSR getSeriesResponse
	err = util.ParseJson(body, &gSR)
	if err != nil {
		return nil, err
	}
	return gSR, nil
}

func (s *SonarrClient) GetSonarrDefaults(sonarrDefaultRootPath string, sonarrDefaultQualityProfile string, rpid map[string]string, qpid map[string]int) error {
	// Root path
	sonarrRootPaths, err := s.GetRootFolder()
	if len(sonarrRootPaths) == 0 || err != nil {
		return errors.New("SONARR ROOT PATH ERROR")
	}

	for _, s := range sonarrRootPaths {
		rpid["TV Series: "+s.Path] = s.Path
	}
	if sonarrDefaultRootPath == "" {
		s.DefaultRootPath = "TV Series: " + sonarrRootPaths[0].Path
	} else {
		s.DefaultRootPath = "TV Series: " + sonarrDefaultRootPath
	}
	// Quality Profiles
	sonarrQualityProfiles, err := s.GetQualityProfiles()
	if len(sonarrQualityProfiles) == 0 || err != nil {
		return errors.New("SONARR QUALITY PATH ERROR")
	}

	for _, v := range sonarrQualityProfiles {
		qpid["TV Series: "+v.Name] = v.Id
	}
	if sonarrDefaultQualityProfile == "" {
		s.DefaultQualityProfile = "TV Series: " + sonarrQualityProfiles[0].Name
	} else {
		s.DefaultQualityProfile = "TV Series: " + sonarrDefaultQualityProfile
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
