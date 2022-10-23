package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var httpClient = http.DefaultClient

//var baseUrl = "https://oengus.io/api/v1"
var baseUrl = "https://oengus.dev/api/v1"

func GetMarathonStats(marathonId string) (*MarathonStats, error) {
	url := fmt.Sprintf("%s/marathons/%s/stats", baseUrl, marathonId)
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	req.Header.Set("User-Agent", "oengus.io/bot")

	res, httpErr := httpClient.Do(req)
	if httpErr != nil {
		log.Println(httpErr)
		return nil, httpErr
	}

	if res.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("You are not authorised to do this action")
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, errors.New("Marathon not found!")
	}

	// defer calls are not executed until the function returns
	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Print(readErr)
		return nil, readErr
	}

	//log.Println(string(body))

	var response *MarathonStats

	jsonErr := json.Unmarshal(body, &response)
	if jsonErr != nil {
		log.Println(jsonErr)
		return response, jsonErr
	}

	return response, nil
}

// 1. load all submissions
// 2. load statuses
// TODO: make sure user gets a warning when selection is not published

func FetchAcceptedRunners(marathonId string) ([]Submission, error) {
	acceptedStatuses := map[string]byte{
		"VALIDATED": 0, "BONUS": 0, "BACKUP": 0,
	}

	keepGoing := true
	page := 1
	submissionList := make([]Submission, 0)

	for keepGoing {
		submissionResponse, err := _fetchSubmissionPage(marathonId, page)

		if err != nil {
			keepGoing = false
			return nil, err
		}

		keepGoing = !submissionResponse.Last && !submissionResponse.Empty
		page++

		if submissionResponse.Empty {
			break
		}

		for _, submission := range submissionResponse.Content {
			for _, game := range submission.Games {
				addSubmission := false

				for _, category := range game.Categories {
					fmt.Println(category.Id, category.Status)
					if _, ok := acceptedStatuses[category.Status]; ok {
						addSubmission = true
						goto SubmissionFound
					}
				}

			SubmissionFound:
				if addSubmission {
					submissionList = append(submissionList, submission)
				}
			}
		}
	}

	return submissionList, nil
}

func _fetchSelection(marathonId string, statuses string) (SelectionResponse, error) {
	url := fmt.Sprintf("%s/marathons/%s/selections?status=%s", baseUrl, marathonId, statuses)
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	req.Header.Set("User-Agent", "oengus.io/bot")

	res, httpErr := httpClient.Do(req)
	if httpErr != nil {
		log.Println(httpErr)
		return nil, httpErr
	}

	if res.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("not_published")
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, errors.New("not found?")
	}

	// defer calls are not executed until the function returns
	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Print(readErr)
		return nil, readErr
	}

	//log.Println(string(body))

	var response SelectionResponse

	jsonErr := json.Unmarshal(body, &response)
	if jsonErr != nil {
		log.Println(jsonErr)
		return response, jsonErr
	}

	return response, nil
}

func _fetchSubmissionPage(marathonId string, page int) (*SubmissionResponse, error) {
	url := fmt.Sprintf("%s/marathons/%s/submissions?page=%d", baseUrl, marathonId, page)
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	req.Header.Set("User-Agent", "oengus.io/bot")

	res, httpErr := httpClient.Do(req)
	if httpErr != nil {
		log.Println(httpErr)
		return nil, httpErr
	}

	// defer calls are not executed until the function returns
	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Print(readErr)
		return nil, readErr
	}

	//log.Println(string(body))

	var response *SubmissionResponse

	jsonErr := json.Unmarshal(body, &response)
	if jsonErr != nil {
		log.Println(jsonErr)
		return response, jsonErr
	}

	return response, nil
}
