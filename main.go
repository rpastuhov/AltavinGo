package main

import (
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

	file, err := os.OpenFile("bot.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Failed to open log file: ", err)
	}
	defer file.Close()

	log.SetOutput(file)

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal("Failed to read config: ", err)
	}

	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		log.Fatal("Failed creating Discord session: ", err)
	}

	b := bot.NewBot(dg, cfg)
	b.RegisterHandlers()

	if err = dg.Open(); err != nil {
		log.Fatal("Error opening connection,", err)
	}
	defer dg.Close()

	log.Println("Bot is now running")

	delay := cfg.TimerDelay * time.Minute
	ticker := time.NewTicker(delay)
	go func() {
		for {
			<-ticker.C
			cfg.ApiConfig.RemoveOldHistories(-delay)
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGTERM, os.Interrupt)
	<-sc

}
