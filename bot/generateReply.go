package bot

import (
	"AltavinGo/api"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func DisplayName(m *discordgo.Member) string {
	if m.Nick != "" {
		return m.Nick
	}
	return m.User.GlobalName
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

func pingInNotAllowedChannels(s *discordgo.Session, m *discordgo.MessageCreate, bot *Bot) (bool, error) {
	if guildSettings, ok := bot.GuildSettings[m.GuildID]; ok && guildSettings != m.ChannelID {
		log.Printf("[INFO]: Redirecting %s (%s) to allowed channel in guild %s", m.Author.Username, m.Author.GlobalName, m.GuildID)
		redirectMessage := fmt.Sprintf("I can't answer you here, go to <#%s>.", guildSettings)

		if _, err := s.ChannelMessageSendReply(m.ChannelID, redirectMessage, m.Reference()); err != nil {
			return true, fmt.Errorf("[ERROR]: sending message: %v", err)
		}

		return true, nil
	}

	return false, nil
}

func cooldownCheck(s *discordgo.Session, m *discordgo.MessageCreate, bot *Bot) (bool, error) {
	if bot.UpdateUserCounter(m.GuildID, m.Author.ID) {
		return false, nil
	}

	cooldown := bot.Cooldowns[m.GuildID].Users[m.Author.ID].EndCooldown.Unix()
	timestamp := fmt.Sprintf(
		"You have reached the %d message request limit on this server!\nThe next request will be available <t:%d:R>.",
		bot.Config.MessagesNumberFromUser, cooldown)

	log.Printf("[INFO]: Cooldown triggered for user %s (%s) in guild %s", m.Author.Username, m.Author.GlobalName, m.GuildID)

	if _, err := s.ChannelMessageSendReply(m.ChannelID, timestamp, m.Reference()); err != nil {
		return true, fmt.Errorf("[ERROR]: sending message: %v", err)
	}
	return false, nil
}

func SendReply(s *discordgo.Session, m *discordgo.MessageCreate, bot *Bot) error {
	if m.Author.ID == s.State.User.ID || !isBotMention(s, m) {
		return nil
	}

	if ok, err := pingInNotAllowedChannels(s, m, bot); ok {
		if err != nil {
			return err
		}
		return nil
	}

	if ok, err := cooldownCheck(s, m, bot); ok {
		if err != nil {
			return err
		}
		return nil
	}

	if err := s.ChannelTyping(m.ChannelID); err != nil {
		return fmt.Errorf("[ERROR]: channel typing: %v", err)
	}

	prompt := strings.TrimSpace(strings.ReplaceAll(m.Content, "<@"+s.State.User.ID+">", ""))

	image, err := api.GetImageBase64(m)
	if err != nil {
		return err
	}

	payload := api.AddToChat(m.ChannelID, "user", prompt, image, bot.Config.Model)

	log.Printf("[INFO]: Processing: '%s' for %s (%s)", prompt, m.Author.Username, m.Author.GlobalName)

	res, err := api.Generate(payload, prompt, bot.Config.BaseURL)
	if err != nil {
		return fmt.Errorf("[ERROR]: generating response: %v", err)
	}

	context := res.Choices[0].Message.Context

	if len(context) > 2000 {
		context = context[:2000]
	}

	api.AddToChat(m.ChannelID, "assistant", context, "", bot.Config.Model)

	log.Printf("[INFO]: Response: '%s' for %s (%s)", context, m.Author.Username, m.Author.GlobalName)

	if _, err := s.ChannelMessageSendReply(
		m.ChannelID,
		replaceMentions(s, m.GuildID, context),
		m.Reference(),
	); err != nil {
		return fmt.Errorf("[ERROR]: message sending: %v", err)
	}

	return nil
}
