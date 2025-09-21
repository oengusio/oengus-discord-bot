package rabbitmq

import (
	"fmt"
	"oenugs-bot/api"
	"oenugs-bot/utils"

	"github.com/bwmarrin/discordgo"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

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

	for _, opponent := range submission.Opponents {
		sendNewOpponentEmbed(dg, params, submission.User.Username, opponent)
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
					MarathonId: params.MarathonId,
					NewSub:     params.NewSub,
				}, marathonName)
			}

			sendNewGame(dg, newGame, submission, api.BotHookParams{
				MarathonId: params.MarathonId,
				NewSub:     params.EditSub,
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

	// 3. Search for added opponents
	for _, newOpponent := range submission.Opponents {
		oldOpponent := findOpponent(newOpponent.Id, oldSubmission)

		if oldOpponent == nil {
			sendNewOpponentEmbed(dg, params, submission.User.Username, newOpponent)
		}
	}

	// 4. Search for deleted opponents
	for _, oldOpponent := range oldSubmission.Opponents {
		newOpponent := findOpponent(oldOpponent.Id, submission)

		if newOpponent == nil {
			sendOpponentRemovedEmbed(dg, params, submission.User.Username, oldOpponent)
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

func sendNewOpponentEmbed(dg *discordgo.Session, params api.BotHookParams, submitter string, opponent api.Opponent) {
	opponentInfo, _ := api.GetOpponentCategoryById(opponent.CategoryId)

	go sendRaceJoinedEmbed(dg, params.NewSub, params.MarathonId, submitter, opponentInfo)

	if params.EditSub != "" {
		go sendRaceJoinedEmbed(dg, params.EditSub, params.MarathonId, submitter, opponentInfo)
	}
}

func sendRaceJoinedEmbed(dg *discordgo.Session, channelId, marathonId, submitter string, categoryInfo api.OpponentCategoryInfoDto) {
	_, err := dg.ChannelMessageSendEmbed(channelId, &discordgo.MessageEmbed{
		URL:   shortUrl + "/" + marathonId + "/submissions",
		Title: utils.EscapeMarkdown(submitter + " joined the multiplayer run for " + utils.EscapeMarkdown(categoryInfo.GameName)),
		Description: fmt.Sprintf(
			"**Category:** %s\n**Estimate:** %s\n",
			utils.EscapeMarkdown(categoryInfo.Name),
			utils.ParseAndMakeDurationPretty(categoryInfo.Estimate),
		),
	})

	if err != nil {
		fmt.Println("Failed to send a message to discord " + err.Error())
	}
}

func sendOpponentRemovedEmbed(dg *discordgo.Session, params api.BotHookParams, submitter string, opponent api.Opponent) {
	categoryInfo, _ := api.GetOpponentCategoryById(opponent.CategoryId)

	_, err := dg.ChannelMessageSendEmbed(params.EditSub, &discordgo.MessageEmbed{
		URL:   shortUrl + "/" + params.MarathonId + "/submissions",
		Title: utils.EscapeMarkdown(submitter + " left the multiplayer run for " + utils.EscapeMarkdown(categoryInfo.GameName)),
		Description: fmt.Sprintf(
			"**Category:** %s\n**Estimate:** %s\n",
			utils.EscapeMarkdown(categoryInfo.Name),
			utils.ParseAndMakeDurationPretty(categoryInfo.Estimate),
		),
	})

	if err != nil {
		fmt.Println("Failed to send a message to discord " + err.Error())
	}
}
