package main

import (
	"fmt"
	"log"
	"oenugs-bot/slashHandlers"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	oengusDiscord  = "601082577729880092"
	CommandGuildId = getEnv("COMMAND_GUILD_ID", "")
	BotToken       = os.Getenv("BOT_TOKEN")
	RemoveCommands = getEnv("REMOVE_COMMANDS_ON_EXIT", "false")
	UpdateCommands = getEnv("UPDATE_SLASH_COMMANDS", "false")

	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "stats",
			Description: "Shows some statistics about the bot",
		},
		{
			Name:        "invite",
			Description: "Shows the link to invite the bot",
		},
		{
			Name:        "discord",
			Description: "Shows the link to the oengus discord",
		},
		{
			Name:        "remove-runner-roles",
			Description: "Removes the role assigned to your runners",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "marathon",
					Description: "The id of a marathon to fetch the runner role for",
					Required:    true,
				},
			},
		},
		{
			Name:        "test",
			Description: "test command",
		},
		{
			Name:        "marathonstats",
			Description: "Shows statistics about a marathon",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "marathon",
					Description: "The id of a marathon",
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"stats":         slashHandlers.OengusStats,
		"invite":        slashHandlers.BotInvite,
		"discord":       slashHandlers.DiscordInvite,
		"marathonstats": slashHandlers.MarathonStats,
		"remove-runner-roles": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.GuildID != oengusDiscord {
				return
			}

			marathonId := i.ApplicationCommandData().Options[0].StringValue()

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					// Flags:
					Content: fmt.Sprintf("Removing role set in marathon `%s` from users", marathonId),
				},
			})

			removeRoleFromRunners(s, "caching", EsaDiscord, esaRunnerRole)
		},
		"test": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.GuildID != oengusDiscord {
				return
			}

			assignRoleToRunnersESA(s, i)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					// Flags:
					Content: "check console",
				},
			})
		},
		// TODO: remove runner roles
		//  - A command that has a role as input and removes it from all members
	}
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// configure the tracking of the state
	dg.StateEnabled = true
	dg.State.MaxMessageCount = 0
	dg.State.TrackChannels = false
	dg.State.TrackEmojis = false
	dg.State.TrackMembers = true
	dg.State.TrackRoles = true
	dg.State.TrackVoice = false
	dg.State.TrackPresences = false

	dg.AddHandler(ready)

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			// TODO: temp for testing
			h(s, i)
		}
	})

	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMembers

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	if RemoveCommands == "true" {
		_, err := dg.ApplicationCommandBulkOverwrite(
			dg.State.User.ID,
			CommandGuildId,
			[]*discordgo.ApplicationCommand{},
		)

		if err != nil {
			fmt.Println("Error clearing commands", err)
			return
		}

		defer dg.Close()
		return
	}

	if UpdateCommands == "true" {
		for _, v := range commands {
			_, err := dg.ApplicationCommandCreate(dg.State.User.ID, CommandGuildId, v)
			if err != nil {
				log.Panicf("Cannot create '%v' command: %v", v.Name, err)
			}
		}
	}

	// Cleanly close down the Discord session.
	defer dg.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	fmt.Println("Gracefully shutting down")
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	usd := &discordgo.UpdateStatusData{
		Status: "online",
	}

	usd.Activities = []*discordgo.Activity{
		{
			Name: "your marathon",
			Type: 3, // Watching
			URL:  "",
		},
	}

	s.UpdateStatusComplex(*usd)

	fmt.Println("Bot is ready!")
}
