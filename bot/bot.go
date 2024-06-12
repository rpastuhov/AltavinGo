package bot

import (
	"fmt"
	"log"
	"ollama-discord/config"
	"ollama-discord/events"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	Session *discordgo.Session
	Config  *config.Config
}

func NewBot(session *discordgo.Session, config *config.Config) *Bot {
	return &Bot{
		Session: session,
		Config:  config,
	}
}

func (bot *Bot) RegisterSlashCommands() error {
	for _, v := range commands {
		_, err := bot.Session.ApplicationCommandCreate(
			bot.Session.State.User.ID, "", v.data,
		)
		if err != nil {
			return err
		}
		log.Printf("Command %v registered\n", v.data.Name)
	}
	return nil
}

func (bot *Bot) RegisterHandlers() {
	bot.Session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Printf("Loging as %s#%s\n", r.User.Username, r.User.Discriminator)
		s.UpdateGameStatus(0, "Chat with AI")
	})

	bot.Session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		events.GenerateReply(s, m, &bot.Config.ApiConfig)
		events.Ping(s, m)
	})

	bot.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if c, ok := commands[i.ApplicationCommandData().Name]; ok {
			c.execute(s, i, bot)
		}
	})
}
