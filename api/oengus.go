package api

type MarathonStats struct {
	SubmissionCount int    `json:"submissionCount"`
	RunnerCount     int    `json:"runnerCount"`
	TotalLength     string `json:"totalLength"`
	AverageEstimate string `json:"averageEstimate"`
}

type MarathonDiscordSettings struct {
	GuildId                string `json:"guild_id"`
	RunnerRoleId           string `json:"runner_role_id"`
	DonationChannel        string `json:"donation_channel"`
	SubmissionChannel      string `json:"submission_channel"`
	SubmissionAuditChannel string `json:"submission_audit_channel"`
}

func getMarathonStats(marathonId string) MarathonStats {
	return MarathonStats{} // TODO
}
