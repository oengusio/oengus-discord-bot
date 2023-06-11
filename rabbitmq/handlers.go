package rabbitmq

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"log"
	"oenugs-bot/api"
	"oenugs-bot/utils"
	"strings"
	"time"
)

// TODO: replace bot webhook with settings.
var shortUrl = "https://oengus.fun"
var eventHandlers = map[string]func(dg *discordgo.Session, data api.WebhookData, params api.BotHookParams){
	// TODO: donation (When we support it again)
	"SUBMISSION_ADD":    handleSubmissionAdd,
	"SUBMISSION_EDIT":   handleSubmissionEdit,
	"SUBMISSION_DELETE": handleSubmissionDelete,
	"GAME_DELETE":       handleGameDelete,
	"CATEGORY_DELETE":   handleCategoryDelete,
	"SELECTION_DONE":    handleSelectionDone,
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

	if params.MarathonId == "" {
		return errors.New("marathon id is missing for rmq event")
	}

	if handler, ok := eventHandlers[data.Event]; ok {
		// We know the references are not null here
		handler(dg, utils.MustNonNil(data), utils.MustNonNil(params))
	}

	return nil
}

func handleSubmissionAdd(dg *discordgo.Session, data api.WebhookData, params api.BotHookParams) {
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

func handleSubmissionEdit(dg *discordgo.Session, data api.WebhookData, params api.BotHookParams) {
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

	canPostNew := params.NewSub != ""
	oldSubmission := data.OriginalSubmission
	submission := data.Submission

	// 1. Search for deleted games/categories
	for _, oldGame := range oldSubmission.Games {
		newGame := findGame(oldGame.Id, submission)
		username := submission.User.Username

		// User removed a game
		if newGame == nil {
			sendGameRemoved(dg, oldGame, username, username, params.EditSub, params.MarathonId, marathonName)
			continue
		}

		nonNilNewGame := utils.MustNonNil(newGame)

		// Check if a category was deleted
		for _, oldCategory := range oldGame.Categories {
			newCategory := findCategory(oldCategory.Id, nonNilNewGame)

			// User removed a category
			if newCategory == nil {
				sendRemovedCategoryEmbed(dg, oldGame, oldCategory, username, username, params.EditSub, params.MarathonId, marathonName)
			}
		}
	}

	// 2. Search for added/updated games/categories
	for _, newGame := range submission.Games {
		oldGame := findGame(newGame.Id, oldSubmission)

		// User as added a new game
		if oldGame == nil {
			// Cheat a little with the parameters
			if canPostNew {
				sendNewGame(dg, newGame, submission, api.BotHookParams{
					NewSub: params.NewSub,
				}, marathonName)
			}

			sendNewGame(dg, newGame, submission, api.BotHookParams{
				NewSub: params.EditSub,
			}, marathonName)
			continue
		}

		nonNilOldGame := utils.MustNonNil(oldGame)
		username := submission.User.Username

		// Check if a category was added or edited
		for _, newCategory := range newGame.Categories {
			oldCategory := findCategory(newCategory.Id, nonNilOldGame)

			// User has added a new category
			if oldCategory == nil {
				if canPostNew {
					sendNewCategoryEmbed(
						dg, newGame, newCategory,
						username, params.NewSub,
						params.MarathonId, marathonName)
				}

				sendNewCategoryEmbed(
					dg, newGame, newCategory,
					username, params.EditSub,
					params.MarathonId, marathonName)
				continue
			}

			nonNilCategory := utils.MustNonNil(oldCategory)

			// We can't compare a pointer to a non-pointer. Also ignore the Categories field in the game
			if cmp.Equal(newCategory, nonNilCategory) && cmp.Equal(newGame, nonNilOldGame, cmpopts.IgnoreFields(api.Game{}, "Categories")) {
				continue
			}

			sendUpdatedCategory(dg, newGame, nonNilOldGame, newCategory, nonNilCategory, params.EditSub, params.MarathonId, username, marathonName)
		}
	}
}

func handleSubmissionDelete(dg *discordgo.Session, data api.WebhookData, params api.BotHookParams) {
	if params.EditSub == "" {
		return
	}

	marathonName, err := api.GetMarathonName(params.MarathonId)

	if err != nil {
		fmt.Println("Failed to look up marathon name for code `" + params.MarathonId + "`: " + err.Error())
		return
	}

	deletedBy := data.DeletedBy.Username
	submitter := data.Submission.User.Username

	for _, game := range data.Submission.Games {
		sendGameRemoved(dg, game, submitter, deletedBy, params.EditSub, params.MarathonId, marathonName)
	}
}

func handleGameDelete(dg *discordgo.Session, data api.WebhookData, params api.BotHookParams) {
	if params.EditSub == "" {
		return
	}

	marathonName, err := api.GetMarathonName(params.MarathonId)

	if err != nil {
		fmt.Println("Failed to look up marathon name for code `" + params.MarathonId + "`: " + err.Error())
		return
	}

	deletedBy := data.DeletedBy.Username
	submitter := data.Submission.User.Username

	sendGameRemoved(dg, data.Game, submitter, deletedBy, params.EditSub, params.MarathonId, marathonName)
}

func handleCategoryDelete(dg *discordgo.Session, data api.WebhookData, params api.BotHookParams) {
	if params.EditSub == "" {
		return
	}

	marathonName, err := api.GetMarathonName(params.MarathonId)

	if err != nil {
		fmt.Println("Failed to look up marathon name for code `" + params.MarathonId + "`: " + err.Error())
		return
	}

	deletedBy := data.DeletedBy.Username
	submitter := data.Submission.User.Username

	sendRemovedCategoryEmbed(dg, data.Game, data.Category, submitter, deletedBy, params.EditSub, params.MarathonId, marathonName)
}

func handleSelectionDone(dg *discordgo.Session, data api.WebhookData, params api.BotHookParams) {
	if params.NewSub == "" {
		return
	}

	channelId := params.NewSub

	_, _ = dg.ChannelMessageSend(channelId, "Runs have been accepted, get ready for the announcements!")

	ticker := time.NewTicker(30 * time.Second)
	quit := make(chan struct{})

	index := 0

	go func() {
		for {
			select {
			case <-ticker.C:
				if index >= len(data.Selections) {
					close(quit)
					_, _ = dg.ChannelMessageSend(channelId, "Th-th-th-th-th-That's all, Folks.")
					return
				}

				selection := data.Selections[index]
				index++

				if selection.Status == "VALIDATED" {
					sendSelectionApprovedEmbed(dg, params.NewSub, selection)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func findGame(gameId int, sub api.Submission) *api.Game {
	for _, game := range sub.Games {
		if game.Id == gameId {
			return &game
		}
	}

	return nil
}

func findCategory(categoryId int, game api.Game) *api.Category {
	for _, category := range game.Categories {
		if category.Id == categoryId {
			return &category
		}
	}

	return nil
}

func sendNewGame(dg *discordgo.Session, game api.Game, submission api.Submission, params api.BotHookParams, marathonName string) {
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

func sendUpdatedCategory(
	dg *discordgo.Session, newGame, oldGame api.Game,
	newCategory, oldCategory api.Category,
	channelId, marathonId, username, marathonName string) {

	builder := strings.Builder{}

	newDuration := utils.ParseAndMakeDurationPretty(newCategory.Estimate)
	oldDuration := utils.ParseAndMakeDurationPretty(oldCategory.Estimate)

	builder.WriteString(fmt.Sprintf(
		"**Game:** %s\n**Category:** %s\n**Platform:** %s\n**Estimate:** %s",
		parseUpdatedString(newGame.Name, oldGame.Name),
		parseUpdatedString(newCategory.Name, oldCategory.Name),
		parseUpdatedString(newGame.Console, oldGame.Console),
		parseUpdatedString(newDuration, oldDuration),
	))

	if newCategory.Video != oldCategory.Video {
		builder.WriteString("\n**Video:** ")
		builder.WriteString(parseUpdatedString(newCategory.Video, oldCategory.Video))
	}

	if newCategory.Type != oldCategory.Type {
		builder.WriteString("\n**Run Type:** ")
		builder.WriteString(parseUpdatedString(newCategory.Type, oldCategory.Type))
	}

	if newCategory.Description != oldCategory.Description {
		builder.WriteString("\n**Category Description:** ")
		builder.WriteString(parseUpdatedString(newCategory.Description, oldCategory.Description))
	}

	if newGame.Description != oldGame.Description {
		builder.WriteString("\n**Game Description:** ")
		builder.WriteString(parseUpdatedString(newGame.Description, oldGame.Description))
	}

	_, err := dg.ChannelMessageSendEmbed(channelId, &discordgo.MessageEmbed{
		URL:         shortUrl + "/" + marathonId + "/submissions",
		Title:       utils.EscapeMarkdown(username + " updated a run in " + marathonName),
		Description: builder.String(),
	})

	if err != nil {
		fmt.Println("Failed to send a message to discord " + err.Error())
	}
}

func sendGameRemoved(dg *discordgo.Session, game api.Game,
	submitter, removedBy, channelId, marathonId, marathonName string) {
	for _, category := range game.Categories {
		sendRemovedCategoryEmbed(dg, game, category, submitter, removedBy, channelId, marathonId, marathonName)
	}
}

func sendRemovedCategoryEmbed(dg *discordgo.Session, game api.Game, cat api.Category,
	submitter, removedBy, channelId, marathonId, marathonName string) {
	var headerText string

	if submitter == removedBy {
		headerText = submitter + " deleted their own run"
	} else {
		headerText = removedBy + " deleted a run by " + submitter
	}

	_, err := dg.ChannelMessageSendEmbed(channelId, &discordgo.MessageEmbed{
		URL:   shortUrl + "/" + marathonId + "/submissions",
		Title: utils.EscapeMarkdown(headerText + " in " + marathonName),
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

func sendSelectionApprovedEmbed(dg *discordgo.Session, channelId string, selection api.SelectionDto) {
	category := selection.Category

	game, gErr := api.GetGameById(category.GameId)

	if gErr != nil {
		fmt.Println("Game lookup failed " + gErr.Error())
		return
	}

	user, uErr := api.GetUserProfile(category.UserId)

	if uErr != nil {
		fmt.Println("User lookup failed " + gErr.Error())
		return
	}

	opponentUsernames := utils.Map(category.Opponents, func(t api.OpponentCategoryDto) string {
		return t.User.Username
	})

	opponents := strings.Join(append([]string{user.Username}, opponentUsernames...), ", ")

	_, err := dg.ChannelMessageSendEmbed(channelId, &discordgo.MessageEmbed{
		URL:   shortUrl + "/" + selection.MarathonId,
		Title: "A run has been accepted!",
		Description: fmt.Sprintf(
			"**Submitted by:** %s\n**Game:** %s\n**Category:** %s\n**Estimate:** %s\n**Platform:** %s\n**Runners:** %s",
			utils.EscapeMarkdown(user.Username),
			utils.EscapeMarkdown(game.Name),
			utils.EscapeMarkdown(category.Name),
			utils.ParseAndMakeDurationPretty(category.Estimate),
			utils.EscapeMarkdown(game.Console),
			utils.EscapeMarkdown(opponents),
		),
	})

	if err != nil {
		fmt.Println("Failed to send a message to discord " + err.Error())
	}
}

func parseUpdatedString(current, old string) string {
	if current == old {
		return utils.EscapeMarkdown(current)
	}

	return utils.EscapeMarkdown(current + " (was " + old + ")")
}
