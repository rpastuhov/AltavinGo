package events

import (
	"fmt"
	"log"
	"ollama-discord/api"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const prompt string = `
You are an AI bot for Discord. Your task is to respond to user messages as part of communication on the server.

Here are some important details you need to consider:

1. Your response should be relevant, helpful and polite.
2. If the text of the user's message ("Content") contains a question or request for help, try to give a clear and informative answer.
3. If the user is responding to a specific message ("Referenced Message"), consider the context of that message when composing your response.
4. Your answer must be less than 2000 characters.
5. Your response will be used for json response, so avoid using multi-line structures and formats that may complicate parsing.
6. Answer in Russian or in the language in which the user wrote.

---

Content: %s
Referenced Message: %s
---

Now, based on this prompt, answer the user’s request.
`

func isBotMention(s *discordgo.Session, m *discordgo.Message) bool {
	for _, user := range m.Mentions {
		if user.ID == s.State.User.ID {
			return true
		}
	}
	return false
}

func replacedBotMention(s *discordgo.Session, m *discordgo.Message) {
	m.Content = strings.ReplaceAll(m.Content, "<@"+s.State.User.ID+">", "<@Your mention>")
}

func getReferenceContent(s *discordgo.Session, msg *discordgo.Message) string {
	if msg.MessageReference == nil {
		return ""
	}
	ref := msg.MessageReference
	mes, err := s.ChannelMessage(ref.ChannelID, ref.MessageID)
	if err != nil || mes.Content == "" {
		return ""
	}

	return mes.Content
}

func GenerateReply(s *discordgo.Session, m *discordgo.MessageCreate, api *api.ApiConfig) {
	log.Println("Message")

	if m.Author.ID == s.State.User.ID || !isBotMention(s, m.Message) {
		return
	}

	replacedBotMention(s, m.Message)
	content := fmt.Sprintf(prompt, m.Content, getReferenceContent(s, m.Message))

	res, err := api.Generate(content, m.ChannelID)
	if err != nil {
		log.Println("Error generating response: ", err)

		s.ChannelMessageSendReply(m.ChannelID,
			"Oh, something happened, write me again😊",
			m.Reference(),
		)

		return
	}

	api.AddToHistory(m.ChannelID, res.Context)

	if len(res.Response) > 2000 {
		res.Response = res.Response[:2000]
	}

	response := fmt.Sprintf("%s\n\nResponse duration: %.2fs",
		res.Response,
		time.Duration(res.TotalDuration).Seconds(),
	)

	s.ChannelMessageSendReply(m.ChannelID, response, m.Reference())
}
