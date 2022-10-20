package api

type MarathonStats struct {
	SubmissionCount int    `json:"submissionCount"`
	RunnerCount     int    `json:"runnerCount"`
	TotalLength     string `json:"totalLength"`
	AverageEstimate string `json:"averageEstimate"`
}

type MarathonSettings struct {
	Discord MarathonDiscordSettings `json:"discord_settings"`
}

type MarathonDiscordSettings struct {
	GuildId                string `json:"guild_id"`
	RunnerRoleId           string `json:"runner_role_id"`
	DonationChannel        string `json:"donation_channel"`
	SubmissionChannel      string `json:"submission_channel"`
	SubmissionAuditChannel string `json:"submission_audit_channel"`
}

type SelectionResponse map[string]struct {
	Id         int    `json:"id"`
	CategoryId int    `json:"categoryId"`
	Status     string `json:"status"`
}

type SubmissionResponse struct {
	Content     []Submission `json:"content"`
	TotalPages  int          `json:"totalPages"`
	CurrentPage int          `json:"currentPage"`
	First       bool         `json:"first"`
	Last        bool         `json:"last"`
	Empty       bool         `json:"empty"`
}

type Submission struct {
	Id    int    `json:"id"`
	User  User   `json:"user"`
	Games []Game `json:"games"`
}

type Game struct {
	Id         int `json:"id"`
	Categories []struct {
		Id     int    `json:"id"`
		Status string `json:"status"`
	} `json:"categories"`
}

// NOTE: we're only storing the most important bits
type User struct {
	Id          int              `json:"id"`
	Username    string           `json:"username"`
	Connections []UserConnection `json:"connections"`
}

type UserConnection struct {
	Id       int    `json:"id"`
	Platform string `json:"platform"`
	Username string `json:"username"`
}