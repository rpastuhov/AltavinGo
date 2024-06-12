package bot

import "github.com/bwmarrin/discordgo"

var clear = command{
	data: &discordgo.ApplicationCommand{
		Name:        "clear",
		Description: "Clears context history in this channel",
	},
	execute: func(s *discordgo.Session, i *discordgo.InteractionCreate, bot *Bot) {
		message := "History cleared!"

		if !bot.Config.ApiConfig.DeleteChannelHistories(i.ChannelID) {
			message = "History is already empty!"
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: message,
			},
		})
	},
}
