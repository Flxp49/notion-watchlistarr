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
	DefaultQualityProfile string
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
		return nil, nil, err
	}
	return resp, body, nil
}

// getRootFolder Response struct
type getRootFolderResponse []struct {
	Path string `json:"path"`
}

// Fetches the rootfolder path set in Radarr
func (r *RadarrClient) GetRootFolder() (getRootFolderResponse, error) {
	_, body, err := r.performReq("GET", "/rootfolder", nil)
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
func (r *RadarrClient) GetQualityProfiles() (qualityProfileResponse, error) {
	_, body, err := r.performReq("GET", "/qualityprofile", nil)
	if err != nil {
		return nil, err
	}
	var qp qualityProfileResponse
	err = util.ParseJson(body, &qp)
	if err != nil {
		return nil, err
	}
	return qp, nil

}

type addMoviePayload struct {
	Title            string `json:"title"`
	QualityProfileId int    `json:"qualityProfileId"`
	TmdbId           int    `json:"tmdbId"`
	RootFolderPath   string `json:"rootFolderPath"`
	Monitored        bool   `json:"monitored"`
	AddOptions       struct {
		SearchForMovie bool   `json:"searchForMovie"`
		Monitor        string `json:"monitor"`
	} `json:"addOptions"`
}

// Add the movie to Radarr
func (r *RadarrClient) AddMovie(title string, qualityProfileId int, tmdbId int, rootFolderPath string, monitored bool, searchForMovie bool) error {
	payload := addMoviePayload{Title: title, QualityProfileId: qualityProfileId, TmdbId: tmdbId, RootFolderPath: rootFolderPath, Monitored: monitored, AddOptions: struct {
		SearchForMovie bool   `json:"searchForMovie"`
		Monitor        string `json:"monitor"`
	}{}}

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
type getMovieResponse []struct {
	HasFile          bool   `json:"hasFile"`
	QualityProfileId int    `json:"qualityProfileId"`
	Monitored        bool   `json:"monitored"`
	RootFolderPath   string `json:"rootFolderPath"`
}

// Fetch movie details in Radarr
//
// id: tmdbid
func (r *RadarrClient) GetMovie(id int) (getMovieResponse, error) {
	_, body, err := r.performReq("GET", fmt.Sprintf("/movie?tmdbId=%d", id), nil)
	if err != nil {
		return nil, err
	}
	var gMR getMovieResponse
	err = util.ParseJson(body, &gMR)
	if err != nil {
		return nil, err
	}
	return gMR, nil
}

func (r *RadarrClient) GetRadarrDefaults(radarrDefaultRootPath string, radarrDefaultQualityProfile string, rpid map[string]string, qpid map[string]int) error {
	// Root path
	radarrRootPaths, err := r.GetRootFolder()
	if len(radarrRootPaths) == 0 || err != nil {
		return errors.New("RADARR ROOT PATH ERROR")
	}

	for _, r := range radarrRootPaths {
		rpid["Movie: "+r.Path] = r.Path
	}
	if radarrDefaultRootPath == "" {
		r.DefaultRootPath = "Movie: " + radarrRootPaths[0].Path
	} else {
		r.DefaultRootPath = "Movie: " + radarrDefaultRootPath
	}
	// Quality Profiles
	radarrQualityProfiles, err := r.GetQualityProfiles()
	if len(radarrQualityProfiles) == 0 || err != nil {
		return errors.New("RADARR QUALITY PATH ERROR")
	}

	for _, v := range radarrQualityProfiles {
		qpid["Movie: "+v.Name] = v.Id
	}
	if radarrDefaultQualityProfile == "" {
		r.DefaultQualityProfile = "Movie: " + radarrQualityProfiles[0].Name
	} else {
		r.DefaultQualityProfile = "Movie: " + radarrDefaultQualityProfile
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
