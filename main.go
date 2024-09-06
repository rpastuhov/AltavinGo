package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"AltavinGo/bot"
	"AltavinGo/config"
)

// Open the log file
func setupLogging() (*os.File, error) {
	file, err := os.OpenFile("bot.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("[ERROR]: Failed to open log file: %v", err)
	}

	mv := io.MultiWriter(file, os.Stdout)
	log.SetOutput(mv)

	return file, nil
}

// Read the config file
func readConfig() (*config.Config, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("[ERROR]: Failed to read config: %v", err)
	}
	return cfg, nil
}

// Set up a signal channel
func waitForSignal() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func main() {
	file, err := setupLogging()
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	cfg, err := readConfig()
	if err != nil {
		log.Fatal(err)
	}

	bot, err := bot.NewBot(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer bot.Session.Close()

	if cfg.RegisterCommands {
		err := bot.RegisterSlashCommands()
		if err != nil {
			log.Printf("[WARNING]: cannot register command: %v", err)
		}
	}

	waitForSignal()
}
