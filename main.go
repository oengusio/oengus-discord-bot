package main

import (
    "fmt"
    "os"
    "os/signal"
    "strings"
    "syscall"

    "github.com/bwmarrin/discordgo"
)

var botToken = os.Getenv("BOT_TOKEN")

func main() {

    // Create a new Discord session using the provided bot token.
    dg, err := discordgo.New("Bot " + botToken)
    if err != nil {
        fmt.Println("error creating Discord session,", err)
        return
    }

    // configure the tracking of the state
    dg.State.MaxMessageCount = 0
    dg.State.TrackChannels = false
    dg.State.TrackEmojis = false
    dg.State.TrackMembers = false
    dg.State.TrackRoles = false
    dg.State.TrackVoice = false
    dg.State.TrackPresences = false

    dg.AddHandler(ready)
    dg.AddHandler(messageCreate)

    // In this example, we only care about receiving message events.
    dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessages)

    // Open a websocket connection to Discord and begin listening.
    err = dg.Open()
    if err != nil {
        fmt.Println("error opening connection,", err)
        return
    }

    // Wait here until CTRL-C or other term signal is received.
    fmt.Println("Bot is now running.  Press CTRL-C to exit.")
    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
    <-sc

    // Cleanly close down the Discord session.
    dg.Close()
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
    usd := &discordgo.UpdateStatusData{
        Status: "online",
    }

    usd.Activities = []*discordgo.Activity{
        {
            Name: "o!help",
            Type: 2, // Listening to
            URL:  "",
        },
    }

    s.UpdateStatusComplex(*usd)
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
            "`o!discord`: Gives the invite to the oengus discord server.\n" +
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
