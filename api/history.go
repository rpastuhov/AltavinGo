package api

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const fileName = "chats.json"

type Message struct {
	Role    string    `json:"role"` // the role of the message, either system, user, assistant, or tool
	Context string    `json:"content"`
	Images  *[]string `json:"images,omitempty"`
}

type Chat struct {
	Model       string    `json:"model"`
	Stream      bool      `json:"stream"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float32   `json:"temperature"`
	LastRequest time.Time `json:"-"`
}

var activeChats = struct {
	sync.RWMutex
	data map[string]*Chat
}{data: make(map[string]*Chat)}

func UnloadInactiveChats(historyTimer time.Duration) error {
	activeChats.Lock()
	defer activeChats.Unlock()

	if err := os.MkdirAll(filepath.Dir(fileName), 0755); err != nil {
		return fmt.Errorf("creating directories: %s", err)
	}

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("creating file: %s", err)
	}
	defer file.Close()

	chatsToSave := make(map[string]*Chat)
	now := time.Now()

	for channelID, chat := range activeChats.data {
		if now.Sub(chat.LastRequest) > historyTimer {
			chatsToSave[channelID] = chat
			delete(activeChats.data, channelID)
		}
	}

	if err := json.NewEncoder(file).Encode(chatsToSave); err != nil {
		return fmt.Errorf("encoding data: %w", err)
	}

	return nil
}

func LoadChatsFromFile(channelID string) (bool, error) {
	activeChats.Lock()
	defer activeChats.Unlock()

	file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	if err != nil {
		return false, fmt.Errorf("opening file: %s", err)
	}
	defer file.Close()

	var data map[string]*Chat
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return false, fmt.Errorf("decoding data: %w", err)
	}

	if chat, ok := data[channelID]; ok {
		activeChats.data[channelID] = chat
		return true, nil
	}

	return false, nil
}

func ChatReset(channelID string) error {
	activeChats.Lock()
	defer activeChats.Unlock()

	file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening file: %s", err)
	}
	defer file.Close()

	var data map[string]Chat
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return fmt.Errorf("decoding data: %w", err)
	}

	delete(data, channelID)
	delete(activeChats.data, channelID)

	if err := json.NewEncoder(file).Encode(data); err != nil {
		return fmt.Errorf("encoding data: %w", err)
	}

	return nil
}

func NewChat(channelID, systemPrompt, modelName string, stream bool, maxTokens int, temperature float32) *Chat {
	if _, ok := activeChats.data[channelID]; !ok {
		chatAdded, err := LoadChatsFromFile(channelID)
		if err != nil {
			log.Printf("[WARNING]: loading chats from file: %s", err)
		}

		if chatAdded {
			return activeChats.data[channelID]
		}
	}

	return &Chat{
		Model:       modelName,
		Stream:      stream,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		Messages: []Message{
			{
				Role:    "system",
				Context: systemPrompt,
				Images:  nil,
			},
		},
	}
}

func (chat *Chat) AddToChat(role, context, image string) *Chat {
	activeChats.Lock()
	defer activeChats.Unlock()

	var images *[]string
	if image != "" {
		images = &[]string{image}
	}

	chat.LastRequest = time.Now()
	chat.Messages = append(chat.Messages, Message{
		Role:    role,
		Context: context,
		Images:  images,
	})

	return chat
}

func GetChatHistory(channelID string) string {
	if _, ok := activeChats.data[channelID]; !ok {
		return ""
	}

	chat := activeChats.data[channelID]
	history := ""

	for id, msg := range chat.Messages {
		history += fmt.Sprintf("%d. [%s]: %s", id, msg.Role, msg.Context)
	}

	return history
}
