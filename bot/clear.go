package bot

import (
	"AltavinGo/api"
	"log"

	"github.com/bwmarrin/discordgo"
)

var clear = command{
	data: &discordgo.ApplicationCommand{
		Name:        "clear",
		Description: "Clears chat history for this channel.",
	},
	execute: func(s *discordgo.Session, i *discordgo.InteractionCreate, bot *Bot) {
		if err := api.ChatReset(i.ChannelID); err == nil {
			log.Printf("[ERROR]: chat reset: %v", err)
		} else {
			log.Printf("[INFO]: Chat has been reset by %s (%s): %s", i.Member.User.Username, i.Member.User.GlobalName, i.GuildID)
		}

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Chat has been reset",
			},
		}

		if err := s.InteractionRespond(i.Interaction, response); err != nil {
			log.Printf("[ERROR]: interaction reply sending: %v", err)
		}
	},
}
