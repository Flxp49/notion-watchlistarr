package notion

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type NotionClient struct {
	user   string
	secret string
	dbid   string
	req    *http.Request
}

func parseJson(body []byte, target interface{}) error {
	return json.Unmarshal(body, target)
}

func (n *NotionClient) performReq(method string, endpoint string, data []byte) (*http.Response, []byte, error) {
	n.req.Method = method
	n.req.URL, _ = url.Parse("https://api.notion.com" + "/" + endpoint)
	if method == "POST" || method == "PATCH" {
		n.req.Body = io.NopCloser(bytes.NewBuffer(data))
		n.req.ContentLength = int64(len(data))
	} else {
		n.req.Body = nil
		n.req.ContentLength = 0
	}
	resp, err := http.DefaultClient.Do(n.req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != http.StatusOK {
		if err == nil {
			err = errors.New(string(body))
		}
		return nil, nil, err
	}
	return resp, body, nil
}

// Payload struct for UpdateDownloadStatus
type updateDownloadStatus struct {
	Properties struct {
		DStatus struct {
			Select struct {
				Name  string `json:"name"`
				Color string `json:"color"`
			} `json:"select"`
		} `json:"Download Status"`
		// Dprogress struct {
		// 	Number float64 `json:"number"`
		// } `json:"Download Progress"`
	} `json:"properties"`
}

// updateDownloadStatus function updates the "Download Status" prop
//
// id - page id to update
//
// status - "Not started" or "In progress" or "Done"
func (n *NotionClient) UpdateDownloadStatus(id string, status string) error {
	UpdateDownloadStatus := updateDownloadStatus{struct {
		DStatus struct {
			Select struct {
				Name  string `json:"name"`
				Color string `json:"color"`
			} `json:"select"`
		} `json:"Download Status"`
	}{DStatus: struct {
		Select struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		} `json:"select"`
	}{Select: struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}{Name: status, Color: "grey"}}}}

	data, err := json.Marshal(UpdateDownloadStatus)
	if err != nil {
		return err
	}
	_, _, err = n.performReq("PATCH", fmt.Sprintf("v1/pages/%s", id), data)
	if err != nil {
		return err
	}
	return nil
}

// QueryDB Response struct
type queryDB struct {
	Results []struct {
		// Pgid       string `json:"id"`
		Properties struct {
			Tmdbid struct {
				Number int `json:"number"`
			} `json:"ID"`
			Name struct {
				Title []struct {
					Plain_text string `json:"plain_text"`
				} `json:"title"`
			}
			Type struct {
				Select struct {
					Name string `json:"name"`
				} `json:"select"`
			} `json:"type"`
		} `json:"properties"`
	} `json:"results"`
}

// QueryDB payload struct
type queryDBPayload struct {
	Filter struct {
		And []struct {
			Property *string `json:"property,omitempty"`
			Checkbox *struct {
				Equals bool `json:"equals"`
			} `json:"checkbox,omitempty"`
			Or []struct {
				Property string `json:"property"`
				Select   struct {
					Equals   *string `json:"equals,omitempty"`
					Is_empty *bool   `json:"is_empty,omitempty"`
				} `json:"select"`
			} `json:"or,omitempty"`
		} `json:"and"`
	} `json:"filter"`
}

// Query DB for titles to Downlaod
//
// mtype : Movie || TV Series
func (n *NotionClient) QueryDB() (queryDB, error) {
	d := "Download"
	nd := "Not Downloaded"
	ndb := true
	payload := queryDBPayload{
		Filter: struct {
			And []struct {
				Property *string "json:\"property,omitempty\""
				Checkbox *struct {
					Equals bool "json:\"equals\""
				} "json:\"checkbox,omitempty\""
				Or []struct {
					Property string "json:\"property\""
					Select   struct {
						Equals   *string "json:\"equals,omitempty\""
						Is_empty *bool   "json:\"is_empty,omitempty\""
					} "json:\"select\""
				} "json:\"or,omitempty\""
			} "json:\"and\""
		}{And: []struct {
			Property *string "json:\"property,omitempty\""
			Checkbox *struct {
				Equals bool "json:\"equals\""
			} "json:\"checkbox,omitempty\""
			Or []struct {
				Property string "json:\"property\""
				Select   struct {
					Equals   *string "json:\"equals,omitempty\""
					Is_empty *bool   "json:\"is_empty,omitempty\""
				} "json:\"select\""
			} "json:\"or,omitempty\""
		}{{Property: &d, Checkbox: &struct {
			Equals bool "json:\"equals\""
		}{Equals: true}}, {Or: []struct {
			Property string "json:\"property\""
			Select   struct {
				Equals   *string "json:\"equals,omitempty\""
				Is_empty *bool   "json:\"is_empty,omitempty\""
			} "json:\"select\""
		}{{Property: "Download Status", Select: struct {
			Equals   *string "json:\"equals,omitempty\""
			Is_empty *bool   "json:\"is_empty,omitempty\""
		}{Equals: &nd}}, {Property: "Download Status", Select: struct {
			Equals   *string "json:\"equals,omitempty\""
			Is_empty *bool   "json:\"is_empty,omitempty\""
		}{Is_empty: &ndb}}}}}},
	}
	data, _ := json.Marshal(payload)
	_, body, err := n.performReq("POST", fmt.Sprintf("v1/databases/%s/query", n.dbid), data)
	if err != nil {
		return queryDB{}, err
	}
	var qDB queryDB
	err = parseJson(body, &qDB)
	if err != nil {
		return queryDB{}, err
	}
	return qDB, nil
}

// addQualityProfiles() updates the DB with the Radarr quality profiles
func (n *NotionClient) addQualityProfiles() error {
	
	return nil
}

// constructor
func InitNotionClient(username string, secret string, dbid string) *NotionClient {
	n := &NotionClient{user: username, secret: secret, dbid: dbid}
	n.req, _ = http.NewRequest("", "", nil)
	n.req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", n.secret))
	n.req.Header.Add("Notion-Version", "2022-06-28")
	n.req.Header.Add("Content-Type", "application/json")
	return n
}
