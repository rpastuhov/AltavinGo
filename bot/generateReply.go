package bot

import (
	"AltavinGo/api"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func oneLine(text string) string {
	return strings.ReplaceAll(text, "\n", " ")
}

func isBotMention(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	for _, user := range m.Mentions {
		if user.ID == s.State.User.ID {
			return true
		}
	}
	return false
}

func replaceMentions(s *discordgo.Session, guildID string, message string) string {
	message = strings.ReplaceAll(message, "@everyone", "everyone")
	message = strings.ReplaceAll(message, "@here", "here")

	patterns := map[*regexp.Regexp]func(string) string{

		regexp.MustCompile(`<@!?(\d+)>`): func(id string) string {
			member, err := s.GuildMember(guildID, id)
			if err != nil {
				return "@" + id
			}
			if member.Nick != "" {
				return "@" + member.Nick
			}
			return "@" + member.User.Username
		},

		regexp.MustCompile(`<@&(\d+)>`): func(id string) string {
			role, err := s.State.Role(guildID, id)
			if err != nil {
				return "@" + id
			}
			return "@" + role.Name
		},

		regexp.MustCompile(`<#(\d+)>`): func(id string) string {
			channel, err := s.State.Channel(id)
			if err != nil {
				return "#" + id
			}
			return "#" + channel.Name
		},
	}

	for pattern, handler := range patterns {
		message = pattern.ReplaceAllStringFunc(message, func(match string) string {
			id := pattern.FindStringSubmatch(match)[1]
			return handler(id)
		})
	}

	return message
}

func checkRestrictions(channelID, guildID, authorID string, bot *Bot) (bool, string) {
	if guildID != "" {
		if guildSettings, ok := bot.GuildSettings[guildID]; ok && guildSettings != channelID {
			log.Printf("[INFO]: %s redirected to allowed channel in guild %s", authorID, guildID)
			return false, fmt.Sprintf("I can't answer you here, go to <#%s>.", guildSettings)
		}
	}

	if !bot.UpdateUserCounter(authorID) {
		return true, ""
	}

	cooldown := bot.Cooldowns[authorID].EndCooldown.Unix()
	timestampMessage := fmt.Sprintf(
		"You have reached the %d message request limit!\nThe next request will be available <t:%d:R>.",
		bot.Config.MaxUserRequests, cooldown)

	log.Printf("[INFO]: %s cooldown triggered", authorID)
	return false, timestampMessage
}

func processRequest(s *discordgo.Session, channelID, guildID, authorID, username, content string, bot *Bot) (string, error) {
	prompt := fmt.Sprintf("%s: %s", username, strings.TrimSpace(content))

	chat := api.NewChat(channelID, bot.Config.SystemPrompt, bot.Config.Model, false, bot.Config.MaxTokens, bot.Config.Temperature)
	payload := chat.AddToChat("user", prompt, bot.Config.HistoryMaxMessages)

	log.Printf("[INFO]: %s/%s/%s: Processing: %s", username, authorID, channelID, oneLine(prompt))

	res, err := api.Generate(payload, prompt, bot.Config.BaseURL, bot.Config.TokenLLM)
	if err != nil {
		return "", fmt.Errorf("generating response: %v", err)
	}

	context := res.Choices[0].Message.Context
	if len(context) > 2000 {
		context = context[:2000]
	}

	chat.AddToChat("assistant", context, bot.Config.HistoryMaxMessages)

	log.Printf("[INFO]: %s/%s/%s: Response (%d tokens): %s", username, authorID, channelID, res.Usage.TotalTokens, oneLine(context))

	return replaceMentions(s, guildID, context), nil
}

func SendReply(s *discordgo.Session, m *discordgo.MessageCreate, bot *Bot) error {
	if m.Author.ID == s.State.User.ID {
		return nil
	}

	isDM := m.GuildID == ""
	if !isDM && !isBotMention(s, m) {
		return nil
	}

	if allowed, message := checkRestrictions(m.ChannelID, m.GuildID, m.Author.ID, bot); !allowed {
		s.ChannelMessageSendReply(m.ChannelID, message, m.Reference())
		return nil
	}

	s.ChannelTyping(m.ChannelID)

	content := strings.TrimSpace(strings.ReplaceAll(m.Content, "<@"+s.State.User.ID+">", ""))
	response, err := processRequest(s, m.ChannelID, m.GuildID, m.Author.ID, m.Author.Username, content, bot)
	if err != nil {
		s.ChannelMessageSendReply(m.ChannelID, "Sorry, I encountered an error while processing your request.", m.Reference())
		return err
	}

	_, err = s.ChannelMessageSendReply(m.ChannelID, response, m.Reference())
	return err
}

var chat = command{
	data: &discordgo.ApplicationCommand{
		Name:             "chat",
		Description:      "Chat with the AI assistant",
		IntegrationTypes: &integrationTypes,
		Contexts:         &contexts,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "message",
				Description: "Your message to the AI",
				Required:    true,
			},
		},
	},
	execute: func(s *discordgo.Session, i *discordgo.InteractionCreate, bot *Bot) {
		message := i.ApplicationCommandData().Options[0].StringValue()

		var userID, username string
		if i.Member != nil {
			userID = i.Member.User.ID
			username = i.Member.User.Username
		} else if i.User != nil {
			userID = i.User.ID
			username = i.User.Username
		}

		if allowed, message := checkRestrictions(i.ChannelID, i.GuildID, userID, bot); !allowed {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: message,
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		}); err != nil {
			log.Printf("[ERROR]: interaction respond: %v", err)
			return
		}

		response, err := processRequest(s, i.ChannelID, i.GuildID, userID, username, message, bot)
		if err != nil {
			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Sorry, an error occurred while processing your request.",
			})
			log.Printf("[ERROR]: interaction chat sending: %v", err)
			return
		}

		_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: response,
		})
		if err != nil {
			log.Printf("[ERROR]: followup message: %v", err)
		}
	},
}
