package config

import (
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	TokenDiscord           string        `json:"tokenDiscord"`
	TokenLLM               string        `json:"tokenLLM"`
	HistoryTimer           time.Duration `json:"historyTimer"`
	BaseURL                string        `json:"base_url"`
	Model                  string        `json:"model"`
	SystemPrompt           string        `json:"system_prompt"`
	MaxTokens              int           `json:"max_tokens"`
	Temperature            float32       `json:"temperature"`
	RegisterCommands       bool          `json:"register_slash_commands"`
	MessagesNumberFromUser int           `json:"number_of_messages_from_user_without_cooldown"`
	CooldownTime           time.Duration `json:"cooldown_time"`
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
