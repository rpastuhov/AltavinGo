package config

import (
	"encoding/json"
	"ollama-discord/api"
	"os"
	"time"
)

type Config struct {
	Token      string        `json:"token"`
	TimerDelay time.Duration `json:"timer-delay"`
	ApiConfig  api.ApiConfig `json:"api-config"`
}

func NewConfig() (*Config, error) {
	configData, err := os.ReadFile("config.json")
	if err != nil {
		return nil, err
	}

	var config Config

	if err = json.Unmarshal(configData, &config); err != nil {
		return nil, err
	}

	config.ApiConfig.Channels = make(map[string]*api.History)
	config.ApiConfig.Users = make(map[string]*api.User)

	return &config, nil
}
