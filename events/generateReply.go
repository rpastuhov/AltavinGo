package events

import (
	"fmt"
	"ollama-discord/api"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func isBotMention(s *discordgo.Session, m *discordgo.Message) bool {
	for _, user := range m.Mentions {

		if user.ID == s.State.User.ID {
			return true
		}
	}
	return false
}

func replacedBotMention(s *discordgo.Session, m *discordgo.Message) {
	for _, user := range m.Mentions {
		if user.ID == s.State.User.ID {
			m.Content = strings.ReplaceAll(m.Content, "<@"+user.ID+">", "<Your mention>")
		}
	}
}

func GenerateReply(s *discordgo.Session, m *discordgo.MessageCreate, api *api.ApiConfig) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if mention := isBotMention(s, m.Message); !mention {
		return
	}

	replacedBotMention(s, m.Message)

	res, err := api.Generate(m.Message.Content, m.ChannelID)
	if err != nil {
		fmt.Println("Error generating response: ", err)
		s.ChannelMessageSendReply(m.ChannelID,
			"Oh, something happened, write me again😊",
			m.Reference(),
		)
		return
	}

	go api.AddToHistory(m.ChannelID, res.Context)

	if len(res.Response) > 2000 {
		res.Response = res.Response[:2000]
	}

	res.Response += fmt.Sprint("\n\nResponse duration: ", time.Duration(res.TotalDuration))

	s.ChannelMessageSendReply(m.ChannelID, res.Response, m.Reference())
}
