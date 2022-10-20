package api

import "net/http"

var httpClient = http.DefaultClient

//var baseUrl = "https://oengus.io/api/v1"
var baseUrl = "https://oengus.dev/api/v1"

func GetMarathonStats(marathonId string) MarathonStats {
	return MarathonStats{} // TODO
}

// 1. load all submissions
// 2. load statuses
// TODO: make sure user gets a warning when selection is not published

func FetchAcceptedRunners(marathonId string) {}

func _fetchStatuses(marathonId string, statuses string) {
	//
}

func _fetchNextSubmissionPage(marathonId string, page int) {
	//
}
