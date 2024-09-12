package bot

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

var ping = command{
	data: &discordgo.ApplicationCommand{
		Name:        "ping",
		Description: "Check bot latency!",
	},
	execute: func(s *discordgo.Session, i *discordgo.InteractionCreate, bot *Bot) {

		g, err := s.Guild(i.GuildID)
		if err != nil {
			log.Printf("[ERROR]: get guild: %v", err)
			return
		}

		log.Printf("[INFO]: %s/%s/%s/%s: Checks the bot's ping", i.Member.User.Username, i.Member.User.GlobalName, g.Name, g.ID)

		start := time.Now()

		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Pinging...",
			},
		}); err != nil {
			log.Printf("[ERROR]: interaction reply sending: %v", err)
			return
		}

		elapsedSeconds := time.Since(start).Seconds() * 1000
		heartbeatLatencySeconds := s.HeartbeatLatency().Seconds() * 1000

		content := fmt.Sprintf(
			"Pong!\nGateway Ping: `%.2f`,\nREST Ping: `%0.2f`",
			elapsedSeconds,
			heartbeatLatencySeconds,
		)

		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		}); err != nil {
			log.Printf("[ERROR]: editing interaction response: %v", err)
		}
	},
}
