package radarr

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
)

type RadarrClient struct {
	req                   *http.Request
	hostpath              string
	DefaultRootPath       string
	DefaultQualityProfile string
}

func parseJson(body []byte, target interface{}) error {
	return json.Unmarshal(body, target)
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
type getRootFolder []struct {
	Path string `json:"path"`
}

// Fetches the rootfolder path set in Radarr
func (r *RadarrClient) GetRootFolder() (getRootFolder, error) {
	_, body, err := r.performReq("GET", "/rootfolder", nil)
	if err != nil {
		return nil, err
	}
	var rf getRootFolder
	err = parseJson(body, &rf)
	if err != nil {
		return nil, err
	}
	return rf, nil

}

// getQualityProfile Response struct
type qualityProfile []struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// Fetches the quality profiles
func (r *RadarrClient) GetQualityProfiles() (qualityProfile, error) {
	_, body, err := r.performReq("GET", "/qualityprofile", nil)
	if err != nil {
		return nil, err
	}
	var qp qualityProfile
	err = parseJson(body, &qp)
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

func InitRadarrClient(apikey string, hostpath string) *RadarrClient {
	r := &RadarrClient{hostpath: hostpath}
	r.req, _ = http.NewRequest("", "", nil)
	r.req.Header.Add("X-Api-Key", apikey)
	r.req.Header.Add("Content-Type", "application/json")
	return r
}
