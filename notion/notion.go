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
	if err != nil || (resp.StatusCode < 200 || resp.StatusCode >= 300) {
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
		QualityProfile *struct {
			Select struct {
				Name string `json:"name"`
			} `json:"select"`
		} `json:"Quality Profile,omitempty"`
		RootFolder *struct {
			Select struct {
				Name string `json:"name"`
			} `json:"select"`
		} `json:"Root Folder,omitempty"`
	} `json:"properties"`
}

type statusMap struct {
	name  string
	color string
}

var sMap = map[string]statusMap{
	"Error":       {name: "🔴 Error", color: "red"},
	"Downloaded":  {name: "🟢 Downloading", color: "green"},
	"Downloading": {name: "🔵 Downloaded", color: "blue"},
	"Queued":      {name: "🟡 Queued", color: "yellow"},
}

// updateDownloadStatus function updates the "Download Status" prop
//
// id - page id to update
//
// status - "Queued" , "Downloading" , "Downloaded" or "Error"
func (n *NotionClient) UpdateDownloadStatus(id string, status string, qualityProfile string, rootPath string) error {
	payload := updateDownloadStatus{}
	if status != "Error" {
		payload.Properties = struct {
			DStatus struct {
				Select struct {
					Name  string "json:\"name\""
					Color string "json:\"color\""
				} "json:\"select\""
			} "json:\"Download Status\""
			QualityProfile *struct {
				Select struct {
					Name string "json:\"name\""
				} "json:\"select\""
			} "json:\"Quality Profile,omitempty\""
			RootFolder *struct {
				Select struct {
					Name string "json:\"name\""
				} "json:\"select\""
			} "json:\"Root Folder,omitempty\""
		}{QualityProfile: &struct {
			Select struct {
				Name string "json:\"name\""
			} "json:\"select\""
		}{Select: struct {
			Name string "json:\"name\""
		}{Name: qualityProfile}}, RootFolder: &struct {
			Select struct {
				Name string "json:\"name\""
			} "json:\"select\""
		}{struct {
			Name string "json:\"name\""
		}{Name: rootPath}}}
	}
	payload.Properties.DStatus.Select.Name = sMap[status].name
	payload.Properties.DStatus.Select.Color = sMap[status].color
	data, err := json.Marshal(payload)
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
		Pgid       string `json:"id"`
		Properties struct {
			Tmdbid struct {
				Number int `json:"number"`
			} `json:"ID"`
			Name struct {
				Title []struct {
					Plain_text string `json:"plain_text"`
				} `json:"title"`
			}
			QualityProfile struct {
				Select struct {
					Name string `json:"name"`
				} `json:"select"`
			} `json:"Quality Profile,omitempty"`
			RootFolder struct {
				Select struct {
					Name string `json:"name"`
				} `json:"select,omitempty"`
			} `json:"Root Folder"`
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
			Select *struct {
				Equals string `json:"equals"`
			} `json:"select,omitempty"`
			Or []struct {
				Property string `json:"property"`
				Select   struct {
					Equals   *string `json:"equals,omitempty"`
					Is_empty *bool   `json:"is_empty,omitempty"`
				} `json:"select"`
			} `json:"or,omitempty"`
		} `json:"and"`
	} `json:"filter"`
	Page_size int `json:"page_size"`
}

// Query DB for titles to Download
//
// mtype : Movie || TV Series
func (n *NotionClient) QueryDB(mtype string) (queryDB, error) {
	d := "Download"
	nd := "Not Downloaded"
	ndb := true
	nt := "Type"
	payload := queryDBPayload{Filter: struct {
		And []struct {
			Property *string "json:\"property,omitempty\""
			Checkbox *struct {
				Equals bool "json:\"equals\""
			} "json:\"checkbox,omitempty\""
			Select *struct {
				Equals string "json:\"equals\""
			} "json:\"select,omitempty\""
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
		Select *struct {
			Equals string "json:\"equals\""
		} "json:\"select,omitempty\""
		Or []struct {
			Property string "json:\"property\""
			Select   struct {
				Equals   *string "json:\"equals,omitempty\""
				Is_empty *bool   "json:\"is_empty,omitempty\""
			} "json:\"select\""
		} "json:\"or,omitempty\""
	}{{Property: &d, Checkbox: &struct {
		Equals bool "json:\"equals\""
	}{Equals: true}}, {Property: &nt, Select: &struct {
		Equals string "json:\"equals\""
	}{Equals: mtype}}, {Or: []struct {
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
	}{Is_empty: &ndb}}}}}}, Page_size: 5}

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

type addDBPropertiesPayload struct {
	Properties struct {
		QualityProfile struct {
			Type   string `json:"type"`
			Select struct {
				Options []struct {
					Name string `json:"name"`
				} `json:"options"`
			} `json:"select"`
		} `json:"Quality Profile,omitempty"`
		Download struct {
			Type     string   `json:"type"`
			Checkbox struct{} `json:"checkbox"`
		} `json:"Download,omitempty"`
		DownloadStatus struct {
			Type   string `json:"type"`
			Select struct {
			} `json:"select"`
		} `json:"Download Status,omitempty"`
		RootFolder struct {
			Type   string `json:"type"`
			Select struct {
				Options []struct {
					Name string `json:"name"`
				} `json:"options"`
			} `json:"select"`
		} `json:"Root Folder,omitempty"`
	} `json:"properties"`
}

// addQualityProfiles() adds the properties ( Download, Download Status, Quality Profile ) to the DB.
//
// profiles : Radarr/Sonarr quality profiles to add
func (n *NotionClient) AddDBProperties(profiles []string, rootpaths []string) error {
	payload := addDBPropertiesPayload{}
	payload.Properties.QualityProfile.Type = "select"
	for _, u := range profiles {
		payload.Properties.QualityProfile.Select.Options = append(payload.Properties.QualityProfile.Select.Options, struct {
			Name string `json:"name"`
		}{Name: u})
	}
	payload.Properties.RootFolder.Type = "select"
	for _, v := range rootpaths {
		payload.Properties.RootFolder.Select.Options = append(payload.Properties.RootFolder.Select.Options, struct {
			Name string `json:"name"`
		}{Name: v})
	}
	payload.Properties.Download.Type = "checkbox"
	payload.Properties.DownloadStatus.Type = "select"
	data, _ := json.Marshal(payload)
	_, _, err := n.performReq("PATCH", fmt.Sprintf("v1/databases/%s/", n.dbid), data)
	if err != nil {
		return err
	}
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
