package slashHandlers

import (
	"fmt"
	"oenugs-bot/globals"

	"github.com/bwmarrin/discordgo"
)

func OengusStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			// Flags:
			// Content: "Hey there! Congratulations, you just executed your first slash command",
			Embeds: []*discordgo.MessageEmbed{
				{
					Fields: []*discordgo.MessageEmbedField{
						{
							Name: "Bot stats",
							Value: fmt.Sprintf(
								"**Guilds (cached)**: %d",
								len(s.State.Guilds),
							),
						},
						{
							Name: "Yearly marathon stats",
							Value: "**Marathons in 2018**: 2\n" +
								"**Marathons in 2019**: 14\n" +
								"**Marathons in 2020**: 159\n" +
								"**Marathons in 2021**: 247\n" +
								"**Marathons in 2022**: 260\n" +
								"**Marathons in 2023**: 274\n" +
								"**Marathons in 2024**: 241",
						},
					},
				},
			},
		},
	})
}

func BotInvite(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			// Flags:
			Content: "Invite me with this link: <" + globals.ShortUrl + "/bot>",
		},
	})
}

func DiscordInvite(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			// Flags:
			Content: "You can join the Oengus discord by clicking this link: <" + globals.ShortUrl + "/discord>",
		},
	})
}
