package rabbitmq

import (
	"fmt"
	"oenugs-bot/api"
	"oenugs-bot/utils"
	"strings"

	"github.com/bwmarrin/discordgo"
)

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
