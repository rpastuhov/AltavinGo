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

func main() {

	// Open the log file
	file, err := os.OpenFile("bot.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Failed to open log file: ", err)
	}
	defer file.Close()

	mv := io.MultiWriter(file, os.Stdout)

	// Set the log output to the file
	log.SetOutput(mv)

	// Read the configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal("Failed to read config: ", err)
	}

	// Create a new bot and register handlers
	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		log.Fatal("Failed creating Discord session: ", err)
	}

	b := bot.NewBot(dg, cfg)
	b.RegisterHandlers()

	if err = dg.Open(); err != nil {
		log.Fatal("Error opening connection,", err)
	}
	// defer dg.Close()

	if false {
		name, err := b.RegisterSlashCommands(dg)
		if err != nil {
			log.Fatalf("Cannot create '%v' command: %v", name, err)
		}
	}

	log.Println("Bot is now running")

	// Set up a ticker to remove old histories
	delay := cfg.TimerDelay * time.Minute
	ticker := time.NewTicker(delay)
	go func() {
		for {
			<-ticker.C
			cfg.ApiConfig.RemoveOldHistories(-delay)
		}
	}()

	// Set up a signal channel
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGTERM, os.Interrupt)
	<-sc

}
