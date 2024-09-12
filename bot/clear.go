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

		g, err := s.Guild(i.GuildID)
		if err != nil {
			log.Printf("[ERROR]: get guild: %v", err)
			return
		}

		if err := api.ChatReset(i.ChannelID); err != nil {
			log.Printf("[ERROR]: chat reset: %s: %v", i.ChannelID, err)
		} else {
			log.Printf("[INFO]: %s/%s/%s/%s: Chat has been reset", i.Member.User.Username, i.Member.User.GlobalName, g.Name, g.ID)
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
