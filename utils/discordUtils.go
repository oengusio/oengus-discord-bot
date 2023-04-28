package utils

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

func OptionsToMap(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))

	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	return optionMap
}

var replacer = strings.NewReplacer(
	"*", "\\*",
	"_", "\\_",
	"`", "\\`",
	">", "\\>",
	"||", "\\||",
)

func EscapeMarkdown(input string) string {
	return replacer.Replace(input)
}
