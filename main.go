package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "strings"
    "syscall"

    "github.com/bwmarrin/discordgo"
)

var (
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
        "stats": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
        },
        "invite": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
            s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    // Flags:
                    Content: "Invite me with this link: <https://oengus.fun/bot>",
                },
            })
        },
        "discord": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
            s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    // Flags:
                    Content: "You can join the Oengus discord by clicking this link: <https://oengus.fun/discord>",
                },
            })
        },
        "marathonstats": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
            // https://oengus.io/api/marathons/{marathon}/stats
            s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    Flags:   1 << 6,
                    Content: "WIP",
                },
            })
        },
        "remove-runner-roles": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
            marathonId := i.ApplicationCommandData().Options[0].StringValue()

            s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    // Flags:
                    Content: fmt.Sprintf("Removing role set in %s from users", marathonId),
                },
            })

            removeRoleFromRunners(s, "caching", EsaDiscord, esaRunnerRole)
        },
        "test": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
            assignRoleToRunnersESA(s)
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
    dg.AddHandler(messageCreate) // TODO: remove

    oengusDiscord := "601082577729880092"

    dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
            // TODO: temp for testing
            if i.GuildID == EsaDiscord || i.GuildID == BsgDiscord || i.GuildID == oengusDiscord {
                h(s, i)
            }
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

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    // Ignore all messages created by bots
    if m.Author.Bot || m.Author.System {
        return
    }

    // TODO: slash commands
    switch strings.ToLower(m.Content) {
    case "<@559625844197163008>", "<@!559625844197163008>":
        s.ChannelMessageSend(m.ChannelID, "My commands can be viewed with `o!help`")
        break
    case "o!help":
        s.ChannelMessageSend(m.ChannelID, "Current command list:\n"+
            "`o!invite`: Get an invite link for the bot\n"+
            "`o!discord`: Gives the invite to the oengus discord server.\n"+
            "`o!stats`: Show some some dev stats")
        break
    case "o!invite":
        s.ChannelMessageSend(m.ChannelID, "Invite me with this link: <https://oengus.fun/bot>")
        break
    case "o!discord":
        s.ChannelMessageSend(m.ChannelID, "You can join the Oengus discord by clicking this link: <https://oengus.fun/discord>")
        break
    case "o!stats":
        s.ChannelMessageSendEmbed(m.ChannelID, &discordgo.MessageEmbed{
            Title: "Oengus bot stats (WIP)",
            Description: fmt.Sprintf(
                "**Guilds (cached)**: %d",
                len(s.State.Guilds),
            ),
        })
        break
    }
}
