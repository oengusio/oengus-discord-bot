package rabbitmq

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"oenugs-bot/api"
	"oenugs-bot/globals"
	"oenugs-bot/utils"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// TODO: replace bot webhook with settings.
var shortUrl = globals.ShortUrl
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

	myString := string(rawJson[:])

	log.Println(myString)

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

func handleSelectionDone(dg *discordgo.Session, data api.WebhookData, params api.BotHookParams) {
	if params.NewSub == "" {
		return
	}

	channelId := params.NewSub

	_, _ = dg.ChannelMessageSend(channelId, "Runs have been accepted, get ready for the announcements!")

	ticker := time.NewTicker(30 * time.Second)
	quit := make(chan struct{})
	// Make sure we have all our accepted submissions
	filteredSubmissions := utils.Filter(data.Selections, func(selection api.SelectionDto) bool {
		return selection.Status == "VALIDATED"
	})

	index := 0

	go func() {
		for {
			select {
			case <-ticker.C:
				if index >= len(filteredSubmissions) {
					close(quit)
					_, _ = dg.ChannelMessageSend(channelId, "Th-th-th-th-th-That's all, Folks.")
					return
				}

				selection := filteredSubmissions[index]
				index++

				sendSelectionApprovedEmbed(dg, params.NewSub, selection)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func sendSelectionApprovedEmbed(dg *discordgo.Session, channelId string, selection api.SelectionDto) {
	category := selection.Category

	game, gErr := api.GetGameById(category.GameId)

	if gErr != nil {
		fmt.Println("Game lookup failed " + gErr.Error())
		return
	}

	user, uErr := api.GetUserProfile(selection.UserId)

	if uErr != nil {
		fmt.Println("User lookup failed " + gErr.Error())
		return
	}

	opponentUsernames, oppErr := api.GetOpponentUsernames(category.Id)

	if oppErr != nil {
		fmt.Println("Opponent lookup failed " + oppErr.Error())
		return
	}

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
