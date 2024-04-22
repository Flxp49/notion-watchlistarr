package util

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type RequestError struct {
	StatusCode int
	Err        error
}

func (e *RequestError) Error() string {
	return e.Err.Error()
}

// this checks if title exists in radarr/sonarr
func ExistingTitleErrorHandle(addErr error) (bool, error) {
	type errorMessage []struct {
		ErrorCode string `json:"errorCode"`
	}
	re, ok := addErr.(*RequestError)
	if !ok {
		return false, addErr
	}
	if re.StatusCode == 400 {
		var eM errorMessage
		err := ParseJson([]byte(re.Error()), &eM)
		if err != nil {
			return false, addErr
		}
		if eM[0].ErrorCode == "MovieExistsValidator" || eM[0].ErrorCode == "SeriesExistsValidator" {
			return true, nil
		}
	}
	return false, nil
}

func ParseJson(body []byte, target interface{}) error {
	return json.Unmarshal(body, target)
}

type GetQueueDetailsResponse struct {
	TotalRecords int `json:"totalRecords"`
	Records      []struct {
		Status               string `json:"status"`
		TrackedDownloadState string `json:"trackedDownloadStatus"`
		ErrorMessage         string `json:"errorMessage"`
	} `json:"records"`
}

// this takes the GetQueueDetailsResponse and returns the
func GetDownloadStatus(download GetQueueDetailsResponse) (string, error) {
	if download.TotalRecords == 0 {
		return "Not Downloaded", nil
	} else {
		// check if error message present, if present: return error as status
		// else return downloading
		if download.Records[0].ErrorMessage != "" {
			return "Error", nil
		} else {
			return "Downloading", nil
		}
	}
}

func CheckSamePath(p1 string, p2 string) bool {
	f1 := strings.ReplaceAll(p1, "/", "")
	f1 = strings.ReplaceAll(f1, "\\", "")
	f1 = strings.ToLower(f1)
	f2 := strings.ReplaceAll(p2, "/", "")
	f2 = strings.ReplaceAll(f2, "\\", "")
	f2 = strings.ToLower(f2)
	return f1 == f2
}

type GetSeriesByRemoteIDResponse struct {
	XMLName xml.Name `xml:"Data"`
	Series  struct {
		Text     string `xml:",chardata"`
		Seriesid string `xml:"seriesid"`
		IMDBID   string `xml:"IMDB_ID"`
	} `xml:"Series"`
}

// To be used when sonarr.LookupSeriesByImdbid returns no match
//
// Returns tvdbid
func GetSeriesByRemoteID(imdbid string) (int, error) {
	resp, err := http.Get(fmt.Sprintf("https://thetvdb.com/api/GetSeriesByRemoteID.php?imdbid=%s", imdbid))
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != 200 {
		if err == nil {
			err = errors.New(string(body))
		}
		return -1, err
	}
	var GSBRIR GetSeriesByRemoteIDResponse
	err = xml.Unmarshal(body, &GSBRIR)
	if err != nil {
		return -1, err
	}
	id, err := strconv.Atoi(GSBRIR.Series.Seriesid)
	if err != nil {
		return -1, err
	}
	return id, nil
}
