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

type Message struct {
	Role    string   `json:"role"` // role: the role of the message, either system, user, assistant, or tool
	Context string   `json:"content"`
	Images  []string `json:"-"`
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
	data map[string]Chat
}{data: make(map[string]Chat)}

const fileName = "chats.json"

// type History struct {
// 	data        []int
// 	size        int
// 	lastRequest time.Time
// }

// func newHistory(size int) *History {
// 	return &History{
// 		data:        make([]int, 0),
// 		size:        size,
// 		lastRequest: time.Now(),
// 	}
// }

// func AddToHistory(channelId string, context []int) {
// 	if _, ok := api.Channels[channelId]; !ok {
// 		api.Channels[channelId] = newHistory(api.HistoryTokensSize)
// 	}

// 	h := api.Channels[channelId]
// 	l := len(context)

// 	if l > h.size {
// 		h.data = context[l-h.size : l]
// 	} else {
// 		h.data = context
// 	}

// 	h.lastRequest = time.Now()
// }

// func GetHistory(channelId string) []int {
// 	if _, ok := api.Channels[channelId]; ok {
// 		return api.Channels[channelId].data
// 	}
// 	return nil
// }

// func DeleteChannelHistories(channelId string) bool {
// 	if _, ok := api.Channels[channelId]; !ok {
// 		return false
// 	}

// 	delete(api.Channels, channelId)
// 	return true
// }

// func DeleteOldHistories(delay time.Duration) {
// 	expirationTime := time.Now().Add(delay)

// 	for channelId, history := range api.Channels {
// 		if history.lastRequest.Before(expirationTime) {
// 			api.DeleteChannelHistories(channelId)
// 		}
// 	}
// }

func UnloadInactiveChats(historyTimer time.Duration) error {
	activeChats.Lock()
	defer activeChats.Unlock()

	if err := os.MkdirAll(filepath.Dir(fileName), 0755); err != nil {
		return fmt.Errorf("[ERROR]: creating directories: %s", err)
	}

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("[ERROR]: creating file: %s", err)
	}
	defer file.Close()

	chatsToSave := make(map[string]Chat)
	now := time.Now()

	for chatID, chat := range activeChats.data {
		if now.Sub(chat.LastRequest) > historyTimer {
			chatsToSave[chatID] = chat
			delete(activeChats.data, chatID)
		}
	}

	if err := json.NewEncoder(file).Encode(chatsToSave); err != nil {
		return fmt.Errorf("[ERROR]: encoding data: %w", err)
	}

	return nil
}

func LoadChatsFromFile() error {
	activeChats.Lock()
	defer activeChats.Unlock()

	file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("[ERROR]: opening file: %s", err)
	}
	defer file.Close()

	var data map[string]Chat
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return fmt.Errorf("[ERROR]: decoding data: %w", err)
	}

	for chatID, chat := range data {
		if _, ok := activeChats.data[chatID]; !ok {
			activeChats.data[chatID] = chat
		}
	}

	return nil
}

func ChatReset(channelID string) error {
	activeChats.Lock()
	defer activeChats.Unlock()

	file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("[ERROR]: opening file: %s", err)
	}
	defer file.Close()

	var data map[string]Chat
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return fmt.Errorf("[ERROR]: decoding data: %w", err)
	}

	delete(data, channelID)
	delete(activeChats.data, channelID)

	if err := json.NewEncoder(file).Encode(data); err != nil {
		return fmt.Errorf("[ERROR]: encoding data: %w", err)
	}

	return nil
}

func AddToChat(channelID, role, context, image, modelName string) Chat {

	if role == "user" {
		if err := LoadChatsFromFile(); err != nil {
			log.Printf("[WARNING]: loading chats from file: %s", err)
		}
	}

	activeChats.Lock()
	defer activeChats.Unlock()

	chat, ok := activeChats.data[channelID]
	if !ok {
		chat = Chat{
			Model:       modelName,
			Stream:      false,
			MaxTokens:   1024,
			Temperature: 0.5,
			LastRequest: time.Now(),

			Messages: []Message{
				{
					Role:    "system",
					Context: "You are a helpful assistant.",
				},
				{
					Role:    role,
					Context: context,
					Images:  []string{image},
				},
			},
		}
		activeChats.data[channelID] = chat
	} else {
		chat.LastRequest = time.Now()
		chat.Messages = append(chat.Messages, Message{
			Role:    role,
			Context: context,
			Images:  []string{image},
		})
		activeChats.data[channelID] = chat
	}

	return chat
}

func GetChatHistory(channelID string) (string, bool) {
	if _, ok := activeChats.data[channelID]; !ok {
		return "No chat history available for this channel", false
	}

	chat := activeChats.data[channelID]
	history := ""

	for id, msg := range chat.Messages {
		history += fmt.Sprintf("%d. [%s]: %s", id, msg.Role, msg.Context)
	}

	return history, true
}
