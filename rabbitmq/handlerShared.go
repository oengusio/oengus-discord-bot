package rabbitmq

import (
	"oenugs-bot/api"
	"oenugs-bot/utils"
)

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

func findOpponent(opponentId int, submission api.Submission) *api.Opponent {
	for _, opponent := range submission.Opponents {
		if opponent.Id == opponentId {
			return &opponent
		}
	}

	return nil
}

func parseUpdatedString(current, old string) string {
	if current == old {
		return utils.EscapeMarkdown(current)
	}

	return utils.EscapeMarkdown(current + " (was " + old + ")")
}
