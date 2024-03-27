package util

import (
	"encoding/json"
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
