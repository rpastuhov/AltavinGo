package bot

import (
	"AltavinGo/api"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var history = command{
	data: &discordgo.ApplicationCommand{
		Name:        "history",
		Description: "The history of your chats with the bot.",
	},
	execute: func(s *discordgo.Session, i *discordgo.InteractionCreate, bot *Bot) {
		g, err := s.Guild(i.GuildID)
		if err != nil {
			log.Printf("[ERROR]: get guild: %v", err)
			return
		}

		history := api.GetChatHistory(i.ChannelID)
		log.Printf("[INFO]: %s/%s/%s/%s: Requesting a chat history", i.Member.User.Username, i.Member.User.GlobalName, g.Name, g.ID)

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{},
		}

		if history == "" {
			response.Data.Content = "No chat history available for this channel"
		} else {
			response.Data.Files = []*discordgo.File{
				{
					Name:   "chat_history.txt",
					Reader: strings.NewReader(history),
				},
			}
		}

		if err := s.InteractionRespond(i.Interaction, response); err != nil {
			log.Printf("[ERROR]: interaction reply sending: %v", err)
		}
	},
}
