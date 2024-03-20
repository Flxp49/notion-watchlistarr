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
func ExistingTitleErrorHandle(err error) (bool, error) {
	type errorMessage []struct {
		ErrorCode string `json:"errorCode"`
	}
	re, ok := err.(*RequestError)
	if !ok {
		return false, err
	}
	if re.StatusCode == 400 {
		var eM errorMessage
		err := ParseJson([]byte(re.Error()), &eM)
		if err != nil {
			return false, err
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