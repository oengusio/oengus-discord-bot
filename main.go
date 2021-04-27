package main

import (
    "fmt"
    "os"
    "os/signal"
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

	usd.Game = &discordgo.Game{
		Name: "your marathons",
		Type: 5, // Competing in
		URL:  "",
	}

	s.UpdateStatusComplex(*usd)
}
// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by bots
	if m.Author.Bot {
		return
	}

    switch m.Content {
    case "<@559625844197163008>", "<@!559625844197163008>":
        s.ChannelMessageSend(m.ChannelID, "My commands can be viewed with `o!help`")
        break
    case "o!help":
        s.ChannelMessageSend(m.ChannelID, "Current command list:\n`o!invite`: Get an invite link for the bot")
        break
    case "o!invite":
        s.ChannelMessageSend(m.ChannelID, "Invite me with this link: <https://oengus.fun/bot>")
        break
    }
}
