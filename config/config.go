package config

import (
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	TokenDiscord       string        `json:"token_discord"`
	TokenLLM           string        `json:"token_llm"`
	HistoryTimer       time.Duration `json:"history_timer"`
	HistoryMaxMessages int           `json:"history_max_messages"`
	BaseURL            string        `json:"base_url"`
	Model              string        `json:"model"`
	SystemPrompt       string        `json:"system_prompt"`
	MaxTokens          int           `json:"max_tokens"`
	Temperature        float32       `json:"temperature"`
	RegisterCommands   bool          `json:"register_slash_commands"`
	MaxUserRequests    int           `json:"max_user_requests"`
	CooldownTime       time.Duration `json:"cooldown_time"`
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
