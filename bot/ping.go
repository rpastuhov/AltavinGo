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

		elapsedSeconds := time.Since(start).Seconds()
		heartbeatLatencySeconds := s.HeartbeatLatency().Seconds()

		content := fmt.Sprintf("Pong!\nGateway Ping: `%.2f`,\nREST Ping: `%.2f`",
			elapsedSeconds, heartbeatLatencySeconds)

		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
	},
}
