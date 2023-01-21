package utils

import "github.com/bwmarrin/discordgo"

func OptionsToMap(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
    optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))

    for _, opt := range options {
        optionMap[opt.Name] = opt
    }

    return optionMap
}
