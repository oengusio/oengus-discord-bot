package rabbitmq

import (
    "fmt"
    "oenugs-bot/api"
    "oenugs-bot/globals"
    "time"

    "github.com/bwmarrin/discordgo"
)

func handleSubmissionStatusChanged(dg *discordgo.Session, data api.WebhookData, params api.BotHookParams) {
    if data.SubmissionStatus.Open {
        // opened
        var messageToSend = fmt.Sprintf(
            "Submissions for %s have been opened! Visit <%s/%s> to submit your runs",
            data.SubmissionStatus.MarathonName,
            globals.ShortUrl,
            params.MarathonId,
        )

        if data.SubmissionStatus.ClosesAt != "" {
            // TODO: error is always not nil
            var t, _ = time.Parse(time.RFC3339, data.SubmissionStatus.ClosesAt)

            messageToSend += fmt.Sprintf("\n\nSubmissions are open until <t:%d:f>", t.Unix())
        }

        dg.ChannelMessageSend(params.NewSub, messageToSend)

        return
    }

    // Item is closed

    var messageToSend = fmt.Sprintf("Submissions for %s are now closed", data.SubmissionStatus.MarathonName)

    dg.ChannelMessageSend(params.NewSub, messageToSend)
}
