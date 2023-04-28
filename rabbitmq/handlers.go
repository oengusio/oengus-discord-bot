package rabbitmq

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"oenugs-bot/api"
	"oenugs-bot/utils"
)

var shortUrl = "https://oengus.fun"
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
	if params.NewSub == "" {
		return
	}

	// TODO: get marathon name for code
	marathonName, err := api.GetMarathonName(params.MarathonId)

	if err != nil {
		fmt.Println("Failed to look up marathon name for code `" + params.MarathonId + "`: " + err.Error())
		return
	}

	submission := data.Submission

	for _, game := range submission.Games {
		for _, category := range game.Categories {
			sendNewCategoryEmbed(dg, game, category, submission.User.Username, params.NewSub, params.MarathonId, marathonName)

			if params.EditSub != "" && params.EditSub != params.NewSub {
				sendNewCategoryEmbed(dg, game, category, submission.User.Username, params.EditSub, params.MarathonId, marathonName)
			}
		}
	}
}

func sendNewCategoryEmbed(dg *discordgo.Session, game api.Game, cat api.Category, submitter, channelId, marathonId, marathonName string) {
	_, err := dg.ChannelMessageSendEmbed(channelId, &discordgo.MessageEmbed{
		URL:   shortUrl + "/" + marathonId + "/submissions",
		Title: utils.EscapeMarkdown(submitter + " submitted a run to " + marathonName),
		Description: fmt.Sprintf(
			"**Game:** %s\n**Category:** %s\n**Platform:** %s\n**Estimate:** %s",
			utils.EscapeMarkdown(game.Name),
			utils.EscapeMarkdown(cat.Name),
			utils.EscapeMarkdown(game.Console),
			utils.ParseAndMakeDurationPretty(cat.Estimate),
		),
	})

	if err != nil {
		fmt.Println("Failed to send a message to discord " + err.Error())
	}
}
