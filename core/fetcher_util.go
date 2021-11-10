package core

import (
	"errors"
	"net/http"
)

func ProcessFetchResult(client *http.Client, req *http.Request, res *http.Response, err error,
	fulfilResult func(result *StatusFetchResult) error) *StatusFetchResult {
	result := &StatusFetchResult{
		Client:   client,
		Request:  req,
		Response: res,
	}

	if err != nil {
		if res != nil {
			result.StatusCode = res.StatusCode
		}

		result.Error = err
		return result
	}

	result.StatusCode = res.StatusCode
	if res.StatusCode == http.StatusTooManyRequests {
		err := errors.New("rate limited")
		result.Error = err
		return result
	}

	if res.StatusCode >= 300 {
		err := errors.New("non success status code")
		result.Error = err
		return result
	}

	err = fulfilResult(result)
	if err != nil {
		result.Error = err
	}

	return result
}
