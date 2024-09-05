package config

import (
	"encoding/json"
	"os"
	"time"
)

// type Config struct {
// 	Token      string        `json:"token"`
// 	TimerDelay time.Duration `json:"timer-delay"`
// 	ApiConfig  api.ApiConfig `json:"api-config"`
// }

type Config struct {
	Token             string        `json:"token"`
	HistoryTimer      time.Duration `json:"history_timer"`
	HistoryTokensSize int           `json:"histor_tokens_size"`
	BaseURL           string        `json:"base_url"`
	Model             string        `json:"model"`
	SystemPrompt      string        `json:"system_prompt"`
	RegisterCommands  string        `json:"register_slash_commands"`
}

func New() (*Config, error) {
	configData, err := os.ReadFile("config.json")
	if err != nil {
		return nil, err
	}

	var config Config

	if err = json.Unmarshal(configData, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// func NewConfig() (*Config, error) {
// 	configData, err := os.ReadFile("config.json")
// 	if err != nil {
// 		return nil, err
// 	}

// 	var config Config

// 	if err = json.Unmarshal(configData, &config); err != nil {
// 		return nil, err
// 	}

// 	config.ApiConfig.Channels = make(map[string]*api.History)
// 	config.ApiConfig.Users = make(map[string]*api.User)

// 	return &config, nil
// }
