package radarr

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

type RadarrClient struct {
	req                   *http.Request
	hostpath              string
	DefaultRootPath       string
	DefaultQualityProfile int
	DefaultMonitorProfile string
}

func (r *RadarrClient) performReq(method string, endpoint string, data []byte) (*http.Response, []byte, error) {
	r.req.Method = method
	r.req.URL, _ = url.Parse(r.hostpath + "/api/v3" + endpoint)
	if method == "POST" {
		r.req.Body = io.NopCloser(bytes.NewBuffer(data))
		r.req.ContentLength = int64(len(data))
	} else {
		r.req.Body = nil
		r.req.ContentLength = 0
	}
	resp, err := http.DefaultClient.Do(r.req)
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

// Fetches the rootfolder path set in Radarr
func (r *RadarrClient) GetRootFolder() (GetRootFolderResponse, error) {
	_, body, err := r.performReq("GET", "/rootfolder", nil)
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
func (r *RadarrClient) GetQualityProfiles() (QualityProfileResponse, error) {
	_, body, err := r.performReq("GET", "/qualityprofile", nil)
	if err != nil {
		return nil, err
	}
	var qp QualityProfileResponse
	err = util.ParseJson(body, &qp)
	if err != nil {
		return nil, err
	}
	return qp, nil

}

type LookupMovieByImdbidResponse struct {
	Tmdbid int `json:"tmdbId"`
}

// lookup movie via Radarr to get tmdbid
func (r *RadarrClient) LookupMovieByImdbid(imdbId string) (LookupMovieByImdbidResponse, error) {

	_, body, err := r.performReq("GET", fmt.Sprintf("/movie/lookup/imdb?imdbId=%s", imdbId), nil)
	if err != nil {
		return LookupMovieByImdbidResponse{}, err
	}
	var lMBIR LookupMovieByImdbidResponse
	err = util.ParseJson(body, &lMBIR)
	if err != nil {
		return LookupMovieByImdbidResponse{}, err
	}
	return lMBIR, nil
}

// Add the movie to Radarr
//
// monitor : "MovieOnly" | "MovieandCollection" | "None"
func (r *RadarrClient) AddMovie(title string, qualityProfileId int, tmdbId int, rootFolderPath string, monitored bool, searchForMovie bool, monitorProfile string) error {
	type addMoviePayload struct {
		Title            string `json:"title"`
		QualityProfileId int    `json:"qualityProfileId"`
		TmdbId           int    `json:"tmdbId"`
		RootFolderPath   string `json:"rootFolderPath"`
		Monitored        bool   `json:"monitored"`
		AddOptions       struct {
			SearchForMovie bool `json:"searchForMovie"`
			MonitorTypes   string
		} `json:"addOptions"`
	}
	payload := addMoviePayload{Title: title, QualityProfileId: qualityProfileId, TmdbId: tmdbId, RootFolderPath: rootFolderPath, Monitored: monitored, AddOptions: struct {
		SearchForMovie bool `json:"searchForMovie"`
		MonitorTypes   string
	}{SearchForMovie: searchForMovie, MonitorTypes: monitorProfile}}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, _, err = r.performReq("POST", "/movie", data)
	if err != nil {
		return err
	}
	return nil
}

// getMovie response struct
type GetMovieResponse []struct {
	HasFile          bool   `json:"hasFile"`
	QualityProfileId int    `json:"qualityProfileId"`
	Monitored        bool   `json:"monitored"`
	RootFolderPath   string `json:"rootFolderPath"`
	ImdbId           string `json:"imdbId"`
	TmdbId           int    `json:"tmdbId"`
	MovieID          int    `json:"id"`
}

// Fetch movie details in Radarr
func (r *RadarrClient) GetMovie(tmdbId int) (GetMovieResponse, error) {
	var query string
	if tmdbId == -1 {
		query = "/movie"
	} else {
		query = fmt.Sprintf("/movie?tmdbId=%d", tmdbId)
	}
	_, body, err := r.performReq("GET", query, nil)
	if err != nil {
		return nil, err
	}
	var gMR GetMovieResponse
	err = util.ParseJson(body, &gMR)
	if err != nil {
		return nil, err
	}
	return gMR, nil
}

// Fetch movie download status
func (r *RadarrClient) GetQueueDetails(movieID int) (util.GetQueueDetailsResponse, error) {
	_, body, err := r.performReq("GET", fmt.Sprintf("/queue?id=%d", movieID), nil)
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

// Sets the default profiles and fetches the quality, rootpath profiles from radarr
func (r *RadarrClient) RadarrDefaults(radarrDefaultRootPath string, radarrDefaultQualityProfile string, radarrDefaultMonitorProfile string, rpid map[string]string, qpid map[string]int) error {
	//set default monitor
	if radarrDefaultMonitorProfile == "" {
		r.DefaultMonitorProfile = "MovieOnly"
	} else {
		r.DefaultMonitorProfile = radarrDefaultMonitorProfile
	}
	// Root path
	radarrRootPaths, err := r.GetRootFolder()
	if len(radarrRootPaths) == 0 || err != nil {
		return errors.Join(errors.New("failed to fetch radarr root paths from radarr"), err)
	}

	for _, r := range radarrRootPaths {
		rpid["Movie: "+r.Path] = r.Path
	}
	if radarrDefaultRootPath == "" {
		r.DefaultRootPath = radarrRootPaths[0].Path
	} else {
		r.DefaultRootPath = radarrDefaultRootPath
	}
	// Quality Profiles
	radarrQualityProfiles, err := r.GetQualityProfiles()
	if len(radarrQualityProfiles) == 0 || err != nil {
		return errors.Join(errors.New("failed to fetch radarr quality profiles from radarr"), err)
	}

	for _, v := range radarrQualityProfiles {
		qpid["Movie: "+v.Name] = v.Id
	}
	if radarrDefaultQualityProfile == "" {
		r.DefaultQualityProfile = radarrQualityProfiles[0].Id
	} else {
		//check if user passed quality profile is valid or not
		profileId, exists := qpid["Movie: "+radarrDefaultQualityProfile]
		if !exists {
			return errors.New("wrong default quality profile passed")
		}
		r.DefaultQualityProfile = profileId
	}
	return nil
}

func InitRadarrClient(apikey string, hostpath string) *RadarrClient {
	r := &RadarrClient{hostpath: hostpath}
	r.req, _ = http.NewRequest("", "", nil)
	r.req.Header.Add("X-Api-Key", apikey)
	r.req.Header.Add("Content-Type", "application/json")
	return r
}
