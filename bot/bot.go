package bot

import (
	"AltavinGo/api"
	"AltavinGo/config"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

type User struct {
	RequestsCount int
	EndCooldown   time.Time
}

type Server struct {
	Users map[string]User
}

type Bot struct {
	Session       *discordgo.Session
	Config        *config.Config
	GuildSettings map[string]string
	Cooldowns     map[string]Server
}

func NewBot(config *config.Config) (*Bot, error) {
	settings, err := loadGuildsCfg("guilds.json")
	if err != nil {
		return nil, err
	}

	dg, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		return nil, fmt.Errorf("[ERROR]: Failed creating Discord session: %v", err)
	}

	b := &Bot{
		Session:       dg,
		Config:        config,
		GuildSettings: settings,
		Cooldowns:     make(map[string]Server),
	}

	b.RegisterHandlers()
	b.StartTimer()

	if err = dg.Open(); err != nil {
		return nil, fmt.Errorf("[ERROR]: opening connection: %v", err)
	}

	return b, err
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
		err := s.UpdateGameStatus(0, "Chat with AI")
		if err != nil {
			log.Println("[ERROR]: Failed to set user status")
		}
	})

	bot.Session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if err := SendReply(s, m, bot); err != nil {
			log.Println(err)
			if _, err := s.ChannelMessageSendReply(m.ChannelID, "Something went wrong.", m.Reference()); err != nil {
				log.Printf("[ERROR]: message sending: %v", err)
			}
		}
	})

	bot.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if c, ok := commands[i.ApplicationCommandData().Name]; ok {
			c.execute(s, i, bot)
		}
	})
}

func (bot *Bot) StartTimer() {
	delay := bot.Config.HistoryTimer * time.Minute
	ticker := time.NewTicker(delay)
	go func() {
		for {
			<-ticker.C
			api.UnloadInactiveChats(delay)
			bot.ResetUsersCounter()
		}
	}()
}

func (bot *Bot) UpdateGuildCfg(guildId string, data string) error {
	if data == "" {
		delete(bot.GuildSettings, guildId)
	} else {
		bot.GuildSettings[guildId] = data
	}
	return saveJson(bot.GuildSettings, "guilds.json")
}

func saveJson(m map[string]string, fileName string) error {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("[ERROR]: converting to json object: %v", err)
	}

	return os.WriteFile(fileName, jsonData, 0644)
}

func loadGuildsCfg(fileName string) (map[string]string, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("[ERROR]: reading file %s: %v", fileName, err)
	}

	var m map[string]string

	if err = json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("[ERROR]: json syntax problem: %v", err)
	}

	return m, nil
}

func (bot *Bot) UpdateUserCounter(serverID, userID string) bool {
	now := time.Now()

	server, serverExists := bot.Cooldowns[serverID]
	if !serverExists {
		bot.Cooldowns[serverID] = Server{Users: make(map[string]User)}
		server = bot.Cooldowns[serverID]
	}

	user, userExists := server.Users[userID]
	if !userExists {
		server.Users[userID] = User{RequestsCount: 1}
		return true
	}

	if user.EndCooldown.Before(now) {
		delete(bot.Cooldowns, userID)
		return true
	}

	if user.EndCooldown.After(now) {
		return false
	}

	if user.RequestsCount >= bot.Config.MessagesNumberFromUser {
		user.EndCooldown = now.Add(bot.Config.CooldownTime * time.Minute)
		server.Users[userID] = user
		bot.Cooldowns[userID] = server
		return false
	}

	user.RequestsCount++
	server.Users[userID] = user
	bot.Cooldowns[serverID] = server
	return true
}

func (bot *Bot) ResetUsersCounter() {
	now := time.Now()
	for serverID, server := range bot.Cooldowns {
		for userID, user := range server.Users {
			if user.EndCooldown.Before(now) {
				delete(server.Users, userID)
			}
		}
		if len(server.Users) == 0 {
			delete(bot.Cooldowns, serverID)
		}
	}
}

// expirationTime := time.Now().Add(bot.Config.CooldownTime)
// if user.EndCooldown.Before(expirationTime) {
// 	delete(bot.Cooldowns, id)
// }
