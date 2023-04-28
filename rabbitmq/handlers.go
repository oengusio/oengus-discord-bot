package rabbitmq

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/google/go-cmp/cmp"
	"log"
	"oenugs-bot/api"
	"oenugs-bot/utils"
)

var shortUrl = "https://oengus.fun"
var eventHandlers = map[string]func(dg *discordgo.Session, data *api.WebhookData, params *api.BotHookParams){
	"SUBMISSION_ADD":  handleSubmissionAdd,
	"SUBMISSION_EDIT": handleSubmissionEdit,
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

	marathonName, err := api.GetMarathonName(params.MarathonId)

	if err != nil {
		fmt.Println("Failed to look up marathon name for code `" + params.MarathonId + "`: " + err.Error())
		return
	}

	submission := data.Submission

	for _, game := range submission.Games {
		sendNewGame(dg, game, submission, params, marathonName)
	}
}

func handleSubmissionEdit(dg *discordgo.Session, data *api.WebhookData, params *api.BotHookParams) {
	if params.EditSub == "" {
		return
	}

	if cmp.Equal(data.OriginalSubmission, data.Submission) {
		return
	}

	marathonName, err := api.GetMarathonName(params.MarathonId)

	if err != nil {
		fmt.Println("Failed to look up marathon name for code `" + params.MarathonId + "`: " + err.Error())
		return
	}

	fmt.Println("THEY ARE NOT EQUAL!!!!!!")

	canPostNew := params.NewSub != ""
	submission := data.Submission

	for _, newGame := range submission.Games {
		oldGame := findOldGame(newGame.Id, data.OriginalSubmission)

		if oldGame == nil {
			// Cheat a little with the parameters
			if canPostNew {
				sendNewGame(dg, newGame, submission, &api.BotHookParams{
					NewSub: params.NewSub,
				}, marathonName)
			}

			sendNewGame(dg, newGame, submission, &api.BotHookParams{
				NewSub: params.EditSub,
			}, marathonName)
			continue
		}
	}

}

func findOldGame(gameId int, sub api.Submission) *api.Game {
	for _, game := range sub.Games {
		if game.Id == gameId {
			return &game
		}
	}

	return nil
}

func sendNewGame(dg *discordgo.Session, game api.Game, submission api.Submission, params *api.BotHookParams, marathonName string) {
	for _, category := range game.Categories {
		sendNewCategoryEmbed(dg, game, category, submission.User.Username, params.NewSub, params.MarathonId, marathonName)

		if params.EditSub != "" && params.EditSub != params.NewSub {
			sendNewCategoryEmbed(dg, game, category, submission.User.Username, params.EditSub, params.MarathonId, marathonName)
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
