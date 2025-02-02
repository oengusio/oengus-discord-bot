package main

import (
	"fmt"
	"log"
	"oenugs-bot/rabbitmq"
	"oenugs-bot/slashHandlers"
	"oenugs-bot/utils"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	oengusDiscord  = "601082577729880092"
	DuncteId       = "191231307290771456"
	CommandGuildId = utils.GetEnv("COMMAND_GUILD_ID", "")
	BotToken       = os.Getenv("BOT_TOKEN")
	RemoveCommands = utils.GetEnv("REMOVE_COMMANDS_ON_EXIT", "false")
	UpdateCommands = utils.GetEnv("UPDATE_SLASH_COMMANDS", "false")

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
			Name:        "role-management",
			Description: "Allows marathon ***Moderators*** to add and remove roles from runners",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "assign",
					Description: "Assign roles to accepted runners that have their discord linked to oengus",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "marathon",
							Description: "The id of your marathon",
							Required:    true,
						},
						{

							Type:        discordgo.ApplicationCommandOptionRole,
							Name:        "role",
							Description: "The role that the accepted runners need to get",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "Remove the runner rule from runners that have their discord linked to oengus",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "marathon",
							Description: "The id of your marathon",
							Required:    true,
						},
						{

							Type:        discordgo.ApplicationCommandOptionRole,
							Name:        "role",
							Description: "The role that needs to be removed from the accepted runners",
							Required:    true,
						},
					},
				},
			},
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
		"stats":           slashHandlers.OengusStats,
		"invite":          slashHandlers.BotInvite,
		"discord":         slashHandlers.DiscordInvite,
		"marathonstats":   slashHandlers.MarathonStats,
		"role-management": slashHandlers.HandleRoleManagement,
	}
)

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

	dg.AddHandlerOnce(ready)

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
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

	if RemoveCommands == "true" {
		fmt.Println("Removing commands")

		_, err := dg.ApplicationCommandBulkOverwrite(
			dg.State.User.ID,
			CommandGuildId,
			[]*discordgo.ApplicationCommand{},
		)

		if err != nil {
			fmt.Println("Error clearing commands", err)
			return
		}
	}
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	usd := &discordgo.UpdateStatusData{
		Status: "online",
	}

	usd.Activities = []*discordgo.Activity{
		{
			Name: "your marathon",
			Type: discordgo.ActivityTypeWatching, // Watching
			URL:  "",
		},
	}

	s.UpdateStatusComplex(*usd)

	fmt.Println("Bot is ready!")

	go rabbitmq.StartListening(s)
}
