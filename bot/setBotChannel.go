package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var defaultMemberPermissions int64 = discordgo.PermissionAdministrator

var setBotChannel = command{
	data: &discordgo.ApplicationCommand{
		Name:                     "set-bot-channel",
		Description:              "Disables the ability to reply to messages in other channels.",
		DefaultMemberPermissions: &defaultMemberPermissions,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "channel",
				Description: "Channel in which the bot will be available.",
				Type:        discordgo.ApplicationCommandOptionChannel,
				ChannelTypes: []discordgo.ChannelType{
					discordgo.ChannelTypeGuildText,
				},
			},
		},
	},
	execute: func(s *discordgo.Session, i *discordgo.InteractionCreate, bot *Bot) {
		var response string
		var channelID string
		options := i.ApplicationCommandData().Options

		if options != nil {
			channel := options[0].ChannelValue(s)
			if channel == nil {
				response = "Channel not found or unavailable!"
			} else {
				channelID = channel.ID
				response = fmt.Sprintf("The bot is configured for the <#%s> channel", channelID)
			}
		} else {
			response = "Restrictions have been removed, the bot is available in all channels!"
		}

		err := bot.UpdateGuildSettings(i.GuildID, channelID)
		if err != nil {
			response = "Error while saving!!"
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	},
}
