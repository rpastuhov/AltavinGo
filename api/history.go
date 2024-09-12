package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

const fileName = "chats.json"

type Message struct {
	Role    string `json:"role"` // the role of the message, either system, user, assistant, or tool
	Context string `json:"content"`
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

func NewChat(channelID, systemPrompt, modelName string, stream bool, maxTokens int, temperature float32) *Chat {
	if chat, ok := activeChats.data[channelID]; ok {
		return chat
	}

	messages, err := LoadChatsFromFile(channelID)
	if err != nil {
		log.Printf("[WARNING]: loading chats from file: %s", err)
		messages = []Message{}
	}

	chat := &Chat{
		Model:       modelName,
		Stream:      stream,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		Messages: []Message{
			{
				Role:    "system",
				Context: systemPrompt,
			},
		},
	}

	if len(messages) > 0 {
		chat.Messages = append(messages, chat.Messages...)
	}

	activeChats.data[channelID] = chat
	return chat
}

func (chat *Chat) AddToChat(role, context string) *Chat {
	activeChats.Lock()
	defer activeChats.Unlock()

	chat.LastRequest = time.Now()
	chat.Messages = append(chat.Messages, Message{
		Role:    role,
		Context: context,
	})

	return chat
}

func UnloadInactiveChats(historyTimer time.Duration) error {
	activeChats.Lock()
	defer activeChats.Unlock()

	chatsToSave := make(map[string][]Message)
	now := time.Now()

	for channelID, chat := range activeChats.data {
		if now.Sub(chat.LastRequest) > historyTimer {
			log.Printf("[INFO]: unload inactive chat: %s", channelID)
			chatsToSave[channelID] = chat.Messages[1:]
			delete(activeChats.data, channelID)
		}
	}

	if len(chatsToSave) == 0 {
		return nil
	}

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("creating file: %s", err)
	}
	defer file.Close()

	existingChats := make(map[string][]Message)
	if err := json.NewDecoder(file).Decode(&existingChats); err != nil && err != io.EOF {
		return fmt.Errorf("decoding data: %w", err)
	}

	for channelID, messages := range chatsToSave {
		existingChats[channelID] = messages
	}

	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("seeking file: %s", err)
	}

	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("truncating file: %s", err)
	}

	if err := json.NewEncoder(file).Encode(existingChats); err != nil {
		return fmt.Errorf("encoding data: %w", err)
	}

	return nil
}

func LoadChatsFromFile(channelID string) ([]Message, error) {
	activeChats.Lock()
	defer activeChats.Unlock()

	file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("opening file: %s", err)
	}
	defer file.Close()

	data := make(map[string][]Message)
	if err := json.NewDecoder(file).Decode(&data); err != nil && err != io.EOF {
		return nil, fmt.Errorf("decoding data: %w", err)
	}

	if messages, ok := data[channelID]; ok {
		return messages, nil
	}

	return nil, nil
}

func ChatReset(channelID string) error {
	activeChats.Lock()
	defer activeChats.Unlock()

	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("opening file: %s", err)
	}
	defer file.Close()

	var data map[string][]Message
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return fmt.Errorf("decoding data: %w", err)
	}

	delete(data, channelID)
	delete(activeChats.data, channelID)

	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("seeking file: %s", err)
	}

	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("truncating file: %s", err)
	}

	if err := json.NewEncoder(file).Encode(data); err != nil {
		return fmt.Errorf("encoding data: %w", err)
	}

	return nil
}

func GetChatHistory(channelID string) string {
	messages, err := LoadChatsFromFile(channelID)
	if err != nil {
		log.Printf("[WARNING]: loading chats from file: %s", err)
		messages = []Message{}
	}

	chat, ok := activeChats.data[channelID]
	if !ok {
		chat = &Chat{
			Messages: []Message{},
		}
	}

	if len(messages) > 0 {
		chat.Messages = append(messages, chat.Messages...)
	}

	if len(chat.Messages) == 0 {
		return ""
	}

	var history strings.Builder

	fmt.Fprintf(&history, "channel ID: %s\n", channelID)
	for id, msg := range chat.Messages {
		if msg.Role == "system" {
			continue
		}
		fmt.Fprintf(&history, "%d. [%s]: %s\n", id, msg.Role, msg.Context)
	}

	return string(history.String())
}
