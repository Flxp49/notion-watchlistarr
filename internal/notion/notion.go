package notion

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/flxp49/notion-watchlistarr/internal/constant"
	"github.com/flxp49/notion-watchlistarr/internal/util"
)

//	type monitorMap struct {
//		name  string
//		color string
//	}

var MonitorProfiles = map[string]string{
	constant.NotionOptionAllEpisodes:       constant.AllEpisodes,
	constant.NotionOptionFutureEpisodes:    constant.FutureEpisodes,
	constant.NotionOptionMissingEpisodes:   constant.MissingEpisodes,
	constant.NotionOptionExistingEpisodes:  constant.ExistingEpisodes,
	constant.NotionOptionRecentpisodes:     constant.RecentEpisodes,
	constant.NotionOptionPilotEpisode:      constant.PilotEpisode,
	constant.NotionOptionFirstSeason:       constant.FirstSeason,
	constant.NotionOptionLastSeason:        constant.LastSeason,
	constant.NotionOptionMonitorSpecials:   constant.MonitorSpecials,
	constant.NotionOptionUnmonitorSpecials: constant.UnmonitorSpecials,
	constant.NotionOptionMovieOnly:         constant.MovieOnly,
	constant.NotionOptionCollection:        constant.MovieAndCollection,
}

type NotionClient struct {
	secret string
	dbid   string
	req    *http.Request
	Rpid   map[string]string
	Qpid   map[string]int
}

func (n *NotionClient) performNotionReq(method string, endpoint string, data []byte) (*http.Response, []byte, error) {
	n.req.Method = method
	n.req.URL, _ = url.Parse("https://api.notion.com" + "/" + endpoint)
	if method == http.MethodGet {
		n.req.Body = nil
		n.req.ContentLength = 0
	} else {
		n.req.Body = io.NopCloser(bytes.NewBuffer(data))
		n.req.ContentLength = int64(len(data))
	}
	resp, err := http.DefaultClient.Do(n.req)
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

type statusMap struct {
	name  string
	color string
}

var sMap = map[string]statusMap{
	"Error":          {name: "🔴 Error", color: "red"},
	"Not Downloaded": {name: "⚫ Not Downloaded", color: "gray"},
	"Downloading":    {name: "🟢 Downloading", color: "green"},
	"Downloaded":     {name: "🔵 Downloaded", color: "blue"},
	"Queued":         {name: "🟡 Queued", color: "yellow"},
}

// updateDownloadStatus function updates the "Download Status" prop
//
// id - page id to update
//
// download - true || false for checkbox property
//
// status - "Queued" , "Downloading" , "Downloaded" or "Error"
func (n *NotionClient) UpdateDownloadStatus(id string, download bool, status string, qualityProfile string, rootPath string) error {
	type updateDownloadStatus struct {
		Properties struct {
			Download struct {
				Checkbox bool `json:"checkbox"`
			}
			DStatus struct {
				Select struct {
					Name string `json:"name"`
					// Color string `json:"color"`
				} `json:"select"`
			} `json:"Download Status"`
			QualityProfile struct {
				Select *struct {
					Name string `json:"name"`
				} `json:"select"`
			} `json:"Quality Profile"`
			RootFolder struct {
				Select *struct {
					Name string `json:"name"`
				} `json:"select"`
			} `json:"Root Folder"`
		} `json:"properties"`
	}
	payload := updateDownloadStatus{}
	payload.Properties.DStatus.Select.Name = sMap[status].name
	payload.Properties.Download.Checkbox = download
	if status != "Error" && status != "Not Downloaded" {
		payload.Properties.QualityProfile = struct {
			Select *struct {
				Name string "json:\"name\""
			} "json:\"select\""
		}{Select: &struct {
			Name string "json:\"name\""
		}{Name: qualityProfile}}

		payload.Properties.RootFolder = struct {
			Select *struct {
				Name string "json:\"name\""
			} "json:\"select\""
		}{Select: &struct {
			Name string "json:\"name\""
		}{Name: rootPath}}
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, _, err = n.performNotionReq("PATCH", fmt.Sprintf("v1/pages/%s", id), data)
	if err != nil {
		return err
	}
	return nil
}

// QueryDB Response struct
type QueryDBResponse struct {
	Results []struct {
		Pgid       string `json:"id"`
		Properties struct {
			Imdbid struct {
				Rich_text []struct {
					Plain_text string `json:"plain_text"`
				} `json:"rich_text"`
			} `json:"IMDb ID"`
			QualityProfile struct {
				Select struct {
					Name string `json:"name"`
				} `json:"select"`
			} `json:"Quality Profile"`
			RootFolder struct {
				Select struct {
					Name string `json:"name"`
				} `json:"select"`
			} `json:"Root Folder"`
			MonitorProfile struct {
				Select struct {
					Name string `json:"name"`
				} `json:"select"`
			} `json:"Monitor"`
		} `json:"properties"`
	} `json:"results"`
}

// Query DB for titles to Download where download is checked
// mtype : Movie || TV Series
func (n *NotionClient) QueryDB(mtype string) (QueryDBResponse, error) {
	type queryDBPayload struct {
		Filter struct {
			And []struct {
				Property string `json:"property"`
				Checkbox *struct {
					Equals bool `json:"equals"`
				} `json:"checkbox,omitempty"`
				Select *struct {
					Equals string `json:"equals"`
				} `json:"select,omitempty"`
			} `json:"and"`
		} `json:"filter"`
		Page_size int `json:"page_size"`
	}
	payload := queryDBPayload{Filter: struct {
		And []struct {
			Property string "json:\"property\""
			Checkbox *struct {
				Equals bool "json:\"equals\""
			} "json:\"checkbox,omitempty\""
			Select *struct {
				Equals string "json:\"equals\""
			} "json:\"select,omitempty\""
		} `json:"and"`
	}{And: []struct {
		Property string "json:\"property\""
		Checkbox *struct {
			Equals bool "json:\"equals\""
		} "json:\"checkbox,omitempty\""
		Select *struct {
			Equals string "json:\"equals\""
		} "json:\"select,omitempty\""
	}{{Property: "Download", Checkbox: &struct {
		Equals bool "json:\"equals\""
	}{Equals: true}}, {Property: "Type", Select: &struct {
		Equals string "json:\"equals\""
	}{Equals: mtype}}}}, Page_size: 5}

	data, _ := json.Marshal(payload)
	_, body, err := n.performNotionReq("POST", fmt.Sprintf("v1/databases/%s/query", n.dbid), data)
	if err != nil {
		return QueryDBResponse{}, err
	}
	var qDB QueryDBResponse
	err = util.ParseJson(body, &qDB)
	if err != nil {
		return QueryDBResponse{}, err
	}
	return qDB, nil
}

// QueryDBTmdb Response struct
type QueryDBIdResponse struct {
	Results []struct {
		Pgid string `json:"id"`
		// Properties struct {
		// 	Tmdbid struct {
		// 		Number int `json:"number"`
		// 	} `json:"ID"`
		// 	Name struct {
		// 		Title []struct {
		// 			Plain_text string `json:"plain_text"`
		// 		} `json:"title"`
		// 	}
		// } `json:"properties"`
	} `json:"results"`
}

// Query DB for existing titles by TmdbID
//
// id : tmdbid
func (n *NotionClient) QueryDBTmdb(tmdbId int) (QueryDBIdResponse, error) {
	type queryDBTmdbPayload struct {
		Filter struct {
			Property string `json:"property"`
			Number   struct {
				Equals int `json:"equals"`
			} `json:"number"`
		} `json:"filter"`
		Page_size int `json:"page_size"`
	}
	payload := queryDBTmdbPayload{Filter: struct {
		Property string "json:\"property\""
		Number   struct {
			Equals int "json:\"equals\""
		} "json:\"number\""
	}{Number: struct {
		Equals int "json:\"equals\""
	}{Equals: tmdbId}}, Page_size: 5}
	data, _ := json.Marshal(payload)
	_, body, err := n.performNotionReq("POST", fmt.Sprintf("v1/databases/%s/query", n.dbid), data)
	if err != nil {
		return QueryDBIdResponse{}, err
	}
	var qDBT QueryDBIdResponse
	err = util.ParseJson(body, &qDBT)
	if err != nil {
		return QueryDBIdResponse{}, err
	}
	return qDBT, nil
}

// Query DB for existing titles by ImdbID
//
// id : ImdbID
func (n *NotionClient) QueryDBImdb(imdbId string) (QueryDBIdResponse, error) {
	type QueryDBImdbPayload struct {
		Filter struct {
			Property  string `json:"property"`
			Rich_text struct {
				Equals string `json:"equals"`
			} `json:"rich_text"`
		} `json:"filter"`
		Page_size int `json:"page_size"`
	}
	payload := QueryDBImdbPayload{Filter: struct {
		Property  string `json:"property"`
		Rich_text struct {
			Equals string `json:"equals"`
		} `json:"rich_text"`
	}{Property: "IMDb ID", Rich_text: struct {
		Equals string `json:"equals"`
	}{Equals: imdbId}}, Page_size: 5}
	data, _ := json.Marshal(payload)
	_, body, err := n.performNotionReq("POST", fmt.Sprintf("v1/databases/%s/query", n.dbid), data)
	if err != nil {
		return QueryDBIdResponse{}, err
	}
	var qDBI QueryDBIdResponse
	err = util.ParseJson(body, &qDBI)
	if err != nil {
		return QueryDBIdResponse{}, err
	}
	return qDBI, nil
}

// addQualityProfiles() adds the properties ( Download, Download Status, Quality Profile ) to the DB.
//
// profiles : Radarr/Sonarr quality profiles to add
func (n *NotionClient) AddDBProperties(qpid map[string]int, rpid map[string]string) error {
	n.Rpid = rpid
	n.Qpid = qpid
	type addDBPropertiesPayload struct {
		Properties struct {
			QualityProfile struct {
				Type   string `json:"type"`
				Select struct {
					Options []struct {
						Name string `json:"name"`
					} `json:"options"`
				} `json:"select"`
			} `json:"Quality Profile"`
			Download struct {
				Type     string   `json:"type"`
				Checkbox struct{} `json:"checkbox"`
			} `json:"Download"`
			DownloadStatus struct {
				Type   string `json:"type"`
				Select struct {
					Options []struct {
						Name  string `json:"name"`
						Color string `json:"color"`
					} `json:"options"`
				} `json:"select"`
			} `json:"Download Status"`
			RootFolder struct {
				Type   string `json:"type"`
				Select struct {
					Options []struct {
						Name string `json:"name"`
					} `json:"options"`
				} `json:"select"`
			} `json:"Root Folder"`
			Monitor struct {
				Type   string `json:"type"`
				Select struct {
					Options []struct {
						Name string `json:"name"`
					} `json:"options"`
				} `json:"select"`
			} `json:"Monitor"`
		} `json:"properties"`
	}
	payload := addDBPropertiesPayload{}
	payload.Properties.QualityProfile.Type = "select"
	for profile := range qpid {
		payload.Properties.QualityProfile.Select.Options = append(payload.Properties.QualityProfile.Select.Options, struct {
			Name string `json:"name"`
		}{Name: profile})
	}
	payload.Properties.RootFolder.Type = "select"
	for path := range rpid {
		payload.Properties.RootFolder.Select.Options = append(payload.Properties.RootFolder.Select.Options, struct {
			Name string `json:"name"`
		}{Name: path})
	}
	payload.Properties.Monitor.Type = "select"
	for m := range MonitorProfiles {
		payload.Properties.Monitor.Select.Options = append(payload.Properties.Monitor.Select.Options, struct {
			Name string `json:"name"`
		}{Name: m})
	}
	for _, val := range sMap {
		payload.Properties.DownloadStatus.Select.Options = append(payload.Properties.DownloadStatus.Select.Options, struct {
			Name  string "json:\"name\""
			Color string "json:\"color\""
		}{Name: val.name, Color: val.color})
	}
	payload.Properties.Download.Type = "checkbox"
	payload.Properties.DownloadStatus.Type = "select"
	data, _ := json.Marshal(payload)
	_, _, err := n.performNotionReq("PATCH", fmt.Sprintf("v1/databases/%s/", n.dbid), data)
	if err != nil {
		return err
	}
	return nil
}

func (n *NotionClient) GetNotionQualityAndRootProps(qualityProfile int, rootPath string, mtype string) (string, string, error) {
	qualityProfileProp := ""
	rootPathProp := ""
	for key, val := range n.Qpid {
		if val == qualityProfile && strings.Contains(key, mtype) {
			qualityProfileProp = key
			break
		}
	}
	if qualityProfileProp == "" {
		return "", "", errors.New("invalid qpid value passed")
	}
	for key, val := range n.Rpid {
		if util.CheckSamePath(val, rootPath) && strings.Contains(key, mtype) {
			rootPathProp = key
			break
		}
	}
	if rootPathProp == "" {
		return "", "", errors.New("invalid rpid value passed")
	}

	return qualityProfileProp, rootPathProp, nil
}

func (n *NotionClient) GetNotionMonitorProp(monitorProfile string, mtype string) (string, error) {
	for key, val := range MonitorProfiles {
		if val == monitorProfile && strings.Contains(key, mtype) {
			return key, nil
		}
	}
	return "", errors.New("invalid monitorProfile id value passed")
}

func InitNotionClient(secret string, dbid string) *NotionClient {
	n := &NotionClient{secret: secret, dbid: dbid}
	n.req, _ = http.NewRequest("", "", nil)
	n.req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", n.secret))
	n.req.Header.Add("Notion-Version", "2022-06-28")
	n.req.Header.Add("Content-Type", "application/json")
	return n
}
