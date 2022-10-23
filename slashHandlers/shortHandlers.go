package slashHandlers

import (
	"fmt"
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
							Name:  "Yearly marathon stats",
							Value: "**Marathons in 2018**: 2\n**Marathons in 2019**: 14\n**Marathons in 2020**: 159\n**Marathons in 2021**: 247",
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
			Content: "Invite me with this link: <https://oengus.fun/bot>",
		},
	})
}

func DiscordInvite(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			// Flags:
			Content: "You can join the Oengus discord by clicking this link: <https://oengus.fun/discord>",
		},
	})
}
