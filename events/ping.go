package events

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

const pong string = "Pong!"

func Ping(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "ping" {
		msg, _ := s.ChannelMessageSend(m.ChannelID, pong)
		s.ChannelMessageEdit(
			m.ChannelID,
			msg.ID,
			fmt.Sprintf("%s %s", pong, msg.Timestamp.Sub(m.Timestamp)))
	}

}
