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

// Open the log file
func setupLogging() *os.File {
	file, err := os.OpenFile("bot.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Failed to open log file: ", err)
	}

	mv := io.MultiWriter(file, os.Stdout)
	log.SetOutput(mv)

	return file
}

// Read the config file
func readConfig() *config.Config {
	cfg, err := config.New()
	if err != nil {
		log.Fatal("Failed to read config: ", err)
	}
	return cfg
}

// Create a new bot session
func createBotSession(cfg *config.Config) *bot.Bot {
	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		log.Fatal("Failed creating Discord session: ", err)
	}

	b, err := bot.NewBot(dg, cfg)
	if err != nil {
		log.Fatalf("bot initialization error: %v", err)
	}
	b.RegisterHandlers()

	if err = dg.Open(); err != nil {
		log.Fatal("error opening connection: ", err)
	}

	return b
}

// Set up a ticker to remove old histories
func startTicker(cfg *config.Config) {
	delay := cfg.HistoryTimer * time.Minute
	ticker := time.NewTicker(delay)
	go func() {
		for {
			<-ticker.C
			// cfg.ApiConfig.DeleteOldHistories(-delay)
			// cfg.ApiConfig.ResetUsersCounter(-delay)
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
	file := setupLogging()
	defer file.Close()

	cfg := readConfig()
	bot := createBotSession(cfg)
	defer bot.Session.Close()

	if false {
		err := bot.RegisterSlashCommands()
		if err != nil {
			log.Fatal("cannot register command: ", err)
		}
	}

	startTicker(cfg)
	waitForSignal()
}
