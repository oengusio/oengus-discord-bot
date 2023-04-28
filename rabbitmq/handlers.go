package rabbitmq

import (
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"log"
	"oenugs-bot/api"
)

var eventHandlers = map[string]func(dg *discordgo.Session, data *api.WebhookData, params *api.BotHookParams){
	"SUBMISSION_ADD": handleSubmissionAdd,
}

func parseObject(rawJson []byte) (*api.WebhookData, error) {
	var data *api.WebhookData

	jsonErr := json.Unmarshal(rawJson, &data)
	if jsonErr != nil {
		log.Println(jsonErr)
		return nil, jsonErr
	}

	return data, nil
}

func handleIncomingEvent(rawJson []byte, dg *discordgo.Session) error {
	data, e := parseObject(rawJson)

	if e != nil {
		return e
	}

	params, e2 := api.GetBotParamsFromUrl(data.Url)

	if e2 != nil {
		return e2
	}

	if handler, ok := eventHandlers[data.Event]; ok {
		handler(dg, data, params)
	}

	return nil
}

func handleSubmissionAdd(dg *discordgo.Session, data *api.WebhookData, params *api.BotHookParams) {
	// TODO: get marathon name for code

	for _, game := range data.Submission.Games {
		sendNewGame(dg, game, params.NewSub, params.MarathonId, "")
	}
}

func sendNewGame(dg *discordgo.Session, game api.Game, channelId, marathonId, marathonName string) {
	//
}
