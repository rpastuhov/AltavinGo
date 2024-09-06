package bot

import (
	"AltavinGo/api"
	"log"

	"github.com/bwmarrin/discordgo"
)

var clear = command{
	data: &discordgo.ApplicationCommand{
		Name:        "clear",
		Description: "Clears context history in this channel",
	},
	execute: func(s *discordgo.Session, i *discordgo.InteractionCreate, bot *Bot) {
		if err := api.ChatReset(i.ChannelID); err == nil {
			log.Printf("[ERROR]: chat reset: %v", err)
		}

		log.Printf("Chat has been reset for %s (%s)", i.Member.User.Username, i.Member.User.GlobalName)

		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Chat has been reset",
			},
		}); err != nil {
			log.Printf("[ERROR]: interaction reply sending: %v", err)
		}
	},
}
