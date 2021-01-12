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

	if m.Content == "o!invite" {
		s.ChannelMessageSend(m.ChannelID, "Invite me with this link: <https://discord.com/api/oauth2/authorize?client_id=559625844197163008&permissions=68608&scope=bot>")
	}
}
