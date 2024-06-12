package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ollama-discord/bot"
	"ollama-discord/config"

	"github.com/bwmarrin/discordgo"
)

// Получить контекст из сообщение на которое отвечаю
// s.ChannelMessage(m.Message.Reference.ChannelID, m.Message.Reference.MessageID)

// Open the log file
func setupLogging() {
	file, err := os.OpenFile("bot.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Failed to open log file: ", err)
	}
	defer file.Close()

	mv := io.MultiWriter(file, os.Stdout)
	log.SetOutput(mv)
}

// Read the config file
func readConfig() *config.Config {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal("Failed to read config: ", err)
	}
	return cfg
}

// Create a new bot session
func createBotSession(token string) *discordgo.Session {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("Failed creating Discord session: ", err)
	}

	if err = dg.Open(); err != nil {
		log.Fatal("Error opening connection,", err)
	}
	defer dg.Close()

	return dg
}

// Set up a ticker to remove old histories
func startTicker(cfg *config.Config) {
	delay := cfg.TimerDelay * time.Minute
	ticker := time.NewTicker(delay)
	go func() {
		for {
			<-ticker.C
			cfg.ApiConfig.DeleteOldHistories(-delay)
		}
	}()
}

// Set up a signal channel
func waitForSignal() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func main() {
	setupLogging()

	cfg := readConfig()
	dg := createBotSession(cfg.Token)

	b := bot.NewBot(dg, cfg)
	b.RegisterHandlers()

	if false {
		err := b.RegisterSlashCommands()
		if err != nil {
			log.Fatal("Cannot register command: ", err)
		}
	}

	log.Println("Bot is now running")

	startTicker(cfg)
	waitForSignal()
}
