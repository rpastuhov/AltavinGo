package bot

import (
	"encoding/json"
	"log"
	"ollama-discord/config"
	"os"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	Session       *discordgo.Session
	Config        *config.Config
	GuildSettings map[string]string
}

func NewBot(session *discordgo.Session, config *config.Config) *Bot {
	settings, err := LoadJson("guilds.json")
	if err != nil {
		log.Fatal(err)
	}

	return &Bot{
		Session:       session,
		Config:        config,
		GuildSettings: settings,
	}
}

func (bot *Bot) RegisterSlashCommands() error {
	var data []*discordgo.ApplicationCommand
	for _, v := range commands {
		data = append(data, v.data)
		log.Printf("Command %v add\n", v.data.Name)
	}

	_, err := bot.Session.ApplicationCommandBulkOverwrite(bot.Session.State.User.ID, "", data)
	if err != nil {
		return err
	}
	log.Printf("%d slash-commands registered\n", len(data))
	return nil
}

func (bot *Bot) RegisterHandlers() {
	bot.Session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Loging as %s#%s\n", r.User.Username, r.User.Discriminator)
		s.UpdateGameStatus(0, "Chat with AI")
	})

	bot.Session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		GenerateReply(s, m, bot)
	})

	bot.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if c, ok := commands[i.ApplicationCommandData().Name]; ok {
			c.execute(s, i, bot)
		}
	})
}

func (bot *Bot) UpdateGuildSettings(guildId string, data string) error {
	if data == "" {
		delete(bot.GuildSettings, guildId)
	} else {
		bot.GuildSettings[guildId] = data
	}
	return SaveJson(bot.GuildSettings, "guilds.json")
}

func SaveJson(m map[string]string, fileName string) error {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return os.WriteFile(fileName, jsonData, 0644)
}

func LoadJson(fileName string) (map[string]string, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		if err := SaveJson(map[string]string{}, fileName); err != nil {
			return nil, err
		}

		data, err = os.ReadFile(fileName)
		if err != nil {
			return nil, err
		}
	}

	var m map[string]string

	if err = json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return m, nil
}
