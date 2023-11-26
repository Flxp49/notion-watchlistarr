package radarr

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type rdrr struct {
	req      *http.Request
	hostpath string
}

func parseJson(body []byte, target interface{}) error {
	return json.Unmarshal(body, target)
}

func (r *rdrr) performReq(method string, endpoint string, data []byte) (*http.Response, []byte, error) {
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
	if err != nil {
		return nil, nil, err
	}
	return resp, body, nil
}

// getRootFolder Response struct
type getRootFolder []struct {
	Id   int    `json:"id"`
	Path string `json:"path"`
}

// Fetches the rootfolder path set in Radarr
func (r *rdrr) getRootFolder() (getRootFolder, error) {
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
type getQualityProfile []struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// Fetches the quality profiles
func (r *rdrr) getQualityProfile() (getQualityProfile, error) {
	_, body, err := r.performReq("GET", "/qualityprofile", nil)
	if err != nil {
		return nil, err
	}
	var qp getQualityProfile
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
		SearchForMovie bool `json:"searchForMovie"`
	} `json:"addOptions"`
}

// Add the movie to Radarr
func (r *rdrr) addMovie(title string, qualityProfileId int, tmdbId int, rootFolderPath string, monitored bool, searchForMovie bool) error {
	payload := addMoviePayload{Title: title, QualityProfileId: qualityProfileId, TmdbId: tmdbId, RootFolderPath: rootFolderPath, Monitored: monitored, AddOptions: struct {
		SearchForMovie bool "json:\"searchForMovie\""
	}{SearchForMovie: searchForMovie}}
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

func Initrdrr(apikey string, hostpath string) (*rdrr, error) {
	r := &rdrr{hostpath: hostpath}
	r.req, _ = http.NewRequest("", "", nil)
	r.req.Header.Add("X-Api-Key", apikey)
	r.req.Header.Add("Content-Type", "application/json")
	return r, nil
}
