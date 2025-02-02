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
	UserId     int         `json:"userId"`
}

type OpponentCategoryInfoDto struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	GameName string `json:"gameName"`
}

type CategoryDto struct {
	Id          int                   `json:"id"`
	Name        string                `json:"name"`
	Estimate    string                `json:"estimate"`
	Description string                `json:"description"`
	Video       string                `json:"video"`
	Type        string                `json:"type"`
	GameId      int                   `json:"gameId"`
	Opponents   []OpponentCategoryDto `json:"opponents"`
}

type OpponentCategoryDto struct {
	Id     int    `json:"id"`
	UserId int    `json:"userId"`
	Video  string `json:"video"`
	// TODO: availabilities?
}
