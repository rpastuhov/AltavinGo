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
	s := strings.ReplaceAll(text, "\n", " ")
	return s
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
		log.Printf("[INFO]: %s/%s/%s: Redirecting to allowed channel", m.Author.Username, m.Author.GlobalName, m.GuildID)
		redirectMessage := fmt.Sprintf("I can't answer you here, go to <#%s>.", guildSettings)

		if _, err := s.ChannelMessageSendReply(m.ChannelID, redirectMessage, m.Reference()); err != nil {
			return true, fmt.Errorf("sending message: %v", err)
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
		bot.Config.MaxUserRequests, cooldown)

	log.Printf("[INFO]: %s/%s/%s: Cooldown triggered", m.Author.Username, m.Author.GlobalName, m.GuildID)

	if _, err := s.ChannelMessageSendReply(m.ChannelID, timestamp, m.Reference()); err != nil {
		return true, fmt.Errorf("sending message: %v", err)
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

	g, err := s.Guild(m.GuildID)
	if err != nil {
		return fmt.Errorf("get guild: %v", err)
	}

	if err := s.ChannelTyping(m.ChannelID); err != nil {
		return fmt.Errorf("channel typing: %v", err)
	}

	prompt := strings.TrimSpace(strings.ReplaceAll(m.Content, "<@"+s.State.User.ID+">", ""))

	chat := api.NewChat(m.ChannelID, bot.Config.SystemPrompt, bot.Config.Model, false, bot.Config.MaxTokens, bot.Config.Temperature)
	payload := chat.AddToChat("user", prompt)

	log.Printf("[INFO]: %s/%s/%s/%s: Processing: %s", m.Author.Username, m.Author.GlobalName, g.Name, g.ID, oneLine(prompt))

	res, err := api.Generate(payload, prompt, bot.Config.BaseURL, bot.Config.TokenLLM)
	if err != nil {
		return fmt.Errorf("generating response: %v", err)
	}

	context := res.Choices[0].Message.Context
	if len(context) > 2000 {
		context = context[:2000]
	}

	chat.AddToChat("assistant", context)

	log.Printf("[INFO]: %s/%s/%s/%s: Response (%d tokens): %s", m.Author.Username, m.Author.GlobalName, g.Name, g.ID, res.Usage.TotalTokens, oneLine(context))

	if _, err := s.ChannelMessageSendReply(m.ChannelID, replaceMentions(s, m.GuildID, context), m.Reference()); err != nil {
		return fmt.Errorf("message sending: %v", err)
	}

	return nil
}
