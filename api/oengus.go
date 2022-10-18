package api

type MarathonStats struct {
    SubmissionCount int    `json:"submissionCount"`
    RunnerCount     int    `json:"runnerCount"`
    TotalLength     string `json:"totalLength"`
    AverageEstimate string `json:"averageEstimate"`
}

func getMarathonStats(marathonId string) MarathonStats {
    return MarathonStats{} // TODO
}
