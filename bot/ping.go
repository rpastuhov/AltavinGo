package bot

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

var ping = command{
	data: &discordgo.ApplicationCommand{
		Name:        "ping",
		Description: "Check bot latency!",
	},
	execute: func(s *discordgo.Session, i *discordgo.InteractionCreate, bot *Bot) {
		start := time.Now()

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Pinging...",
			},
		})

		elapsed := time.Since(start)
		content := fmt.Sprintf("Pong!\nGateway Ping: `%v`,\nREST Ping: `%v`", elapsed, s.HeartbeatLatency())

		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
	},
}
