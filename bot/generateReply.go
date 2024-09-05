package bot

import (
	"fmt"
	"log"
	"ollama-discord/api"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

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

func isBotMention(s *discordgo.Session, m *discordgo.Message) bool {
	for _, user := range m.Mentions {
		if user.ID == s.State.User.ID {
			return true
		}
	}
	return false
}

func pingInNotAllowedChannels(s *discordgo.Session, m *discordgo.MessageCreate, bot *Bot) (bool, error) {
	if guildSettings, ok := bot.GuildSettings[m.GuildID]; ok && guildSettings != m.ChannelID {

		message := fmt.Sprintf("I can't answer you here, go to <#%s>.", guildSettings)

		if _, err := s.ChannelMessageSendReply(m.ChannelID, message, m.Reference()); err != nil {
			return true, fmt.Errorf("error sending message: %v", err)
		}

		return true, nil
	}

	return false, nil
}

func SendReply(s *discordgo.Session, m *discordgo.MessageCreate, bot *Bot) error {
	if m.Author.ID == s.State.User.ID || m.Mentions[0].ID == s.State.User.ID {
		return nil
	}

	if ok, err := pingInNotAllowedChannels(s, m, bot); !ok {
		if err != nil {
			return err
		}
		return nil
	}

	err := s.ChannelTyping(m.ChannelID)
	if err != nil {
		return fmt.Errorf("error channel typing: %v", err)
	}

	prompt := strings.TrimSpace(strings.ReplaceAll(m.Content, "<@"+s.State.User.ID+">", ""))

	image, err := api.GetImageBase64(m)
	if err != nil {
		return err
	}

	payload := api.AddToChat(m.ChannelID, prompt, image, bot.Config.Model)
	log.Printf("[API]: Processing '%s' for %s (%s)", prompt, m.Author.ID, m.Message.Member.Nick)

	// if !bot.Config.ApiConfig.UpdateUserCounter(m.Author.ID) {
	// 	cooldown := bot.Config.ApiConfig.Users[m.Author.ID].EndOfCooldown.Unix()
	// 	timestamp := fmt.Sprintf("You have reached your request limit, the next request can be made <t:%d:R>.", cooldown)
	// 	s.ChannelMessageSendReply(m.ChannelID, timestamp, m.Reference())
	// 	return nil
	// }

	res, err := api.Generate(payload, prompt, bot.Config.BaseURL)

	if err != nil {

		return fmt.Errorf("Error generating response: ", err)

		// log.Println("Error generating response: ", err)

		// s.ChannelMessageSendReply(m.ChannelID,
		// 	"Oh, something happened, write me again😊",
		// 	m.Reference(),
		// )
		// return
	}

	// bot.Config.ApiConfig.AddToHistory(m.ChannelID, res.Context)

	// if len(res.Response) > 2000 {
	// 	res.Response = res.Response[:2000]
	// }

	response := fmt.Sprintf("%s\n\nResponse duration: %.2fs",
		replaceMentions(s, m.GuildID, res.Response),
		time.Duration(res.TotalDuration).Seconds(),
	)

	s.ChannelMessageSendReply(m.ChannelID, response, m.Reference())

	return nil
}
