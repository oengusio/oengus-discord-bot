package api

type GameDto struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Console string `json:"console"`
}

type ProfileDto struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
}

type SelectionDto struct {
	Id         int         `json:"id"`
	MarathonId string      `json:"marathonId"`
	Category   CategoryDto `json:"category"`
	Status     string      `json:"status"`
}

type CategoryDto struct {
	Id          int                   `json:"id"`
	Name        string                `json:"name"`
	Estimate    string                `json:"estimate"`
	Description string                `json:"description"`
	Video       string                `json:"video"`
	Type        string                `json:"type"`
	GameId      int                   `json:"gameId"`
	UserId      int                   `json:"userId"`
	Opponents   []OpponentCategoryDto `json:"opponents"`
}

type OpponentCategoryDto struct {
	Id    int        `json:"id"`
	User  ProfileDto `json:"user"`
	Video string     `json:"video"`
	// TODO: availabilities?
}
