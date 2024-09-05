package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type Request struct {
	Model   string `json:"model"`
	Prompt  string `json:"prompt"`
	Context []int  `json:"context"`
	Stream  bool   `json:"stream"`
}

type Response struct {
	Response      string `json:"response"`
	Context       []int  `json:"context"`
	TotalDuration int    `json:"total_duration"`
}

type Message struct {
	Role    string `json:"role"` // role: the role of the message, either system, user, assistant, or tool
	Context string `json:"content"`
	Images  []string
}

type Chat struct {
	Model    string    `json:"model"`
	Stream   bool      `json:"stream"`
	Messages []Message `json:"messages"`
}

var activeChats = struct {
	sync.RWMutex
	data map[string]Chat
}{data: make(map[string]Chat)}

// type User struct {
// 	RequestsCount int
// 	EndOfCooldown time.Time
// }

// type ApiCfg struct {
// 	Users map[string]*User
// }

// func (api *ApiCfg) UpdateUserCounter(userId string) bool {
// 	user, exist := api.Users[userId]
// 	if !exist {
// 		user = &User{}
// 		api.Users[userId] = user
// 	}

// 	if user.RequestsCount >= 6 {
// 		if user.EndOfCooldown.IsZero() {
// 			user.EndOfCooldown = time.Now().Add(30 * time.Minute)
// 		}
// 		return false
// 	}

// 	user.RequestsCount++

// 	return true
// }

// func (api *ApiCfg) ResetUsersCounter(delay time.Duration) {
// 	expirationTime := time.Now().Add(delay)

// 	for id, user := range api.Users {
// 		if user.EndOfCooldown.Before(expirationTime) {
// 			delete(api.Users, id)
// 		}
// 	}
// }

const system string = `
You are an AI bot for Discord. Your task is to respond to user messages as part of communication on the server.

Here are some important details you need to consider:

1. Your response should be relevant, helpful and polite.
2. If the text of the user's message ("Content") contains a question or request for help, try to give a clear and informative answer.
3. If the user is responding to a specific message ("Referenced Message"), consider the context of that message when composing your response.
4. Your answer must be less than 2000 characters.
5. Your response will be used for json response, so avoid using multi-line structures and formats that may complicate parsing.
6. Answer in Russian or in the language in which the user wrote.

---

Content: %s
Referenced Message: %s
---

Now, based on this prompt, answer the user's request.
`

// func (api *ApiCfg) GenerateOld(content, referenceContent, channelId string) (*Response, error) {
// 	url := api.ApiDomain + "/api/generate"

// 	requestBody, err := json.Marshal(Request{
// 		Model:   api.Model,
// 		Prompt:  fmt.Sprintf(prompt, content, referenceContent),
// 		Context: api.GetHistory(channelId),
// 		Stream:  false,
// 	})
// 	if err != nil {
// 		log.Printf("Error marshalling request: %v", err)
// 		return nil, err
// 	}

// 	res, err := http.Post(url, "application/json", strings.NewReader(string(requestBody)))
// 	if err != nil {
// 		log.Printf("Error sending request to API: %v, Prompt: %s, Reference: %s", err, content, referenceContent)
// 		return nil, err
// 	}

// 	if res.StatusCode != http.StatusOK {
// 		log.Printf("Error response: %s,\nPrompt: %s, Reference: %s", res.Status, content, referenceContent)
// 		return nil, errors.New("Response not 200 OK")
// 	}

// 	defer res.Body.Close()

// 	body, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		log.Printf("Error reading response body: %v", err)
// 		return nil, err
// 	}

// 	var formatted Response

// 	err = json.Unmarshal(body, &formatted)
// 	if err != nil {
// 		log.Printf("Error unmarshalling response: %v", err)
// 		return nil, err
// 	}

// 	return &formatted, nil
// }

// func Generate(payload []interface{}, prompt string) (*Response, error) {

// 	url := "http://localhost:11434/api/chat"

// 	jsonPayload, err := json.Marshal(payload)
// 	if err != nil {
// 		return nil, fmt.Errorf("Error marshalling request: %v", err)
// 	}

// 	res, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
// 	if err != nil {
// 		return nil, fmt.Errorf("Error sending request to API: %v, Prompt: %s", err, prompt)
// 	}

// 	defer res.Body.Close()

// 	if res.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("Error response code: %v", err)
// 	}

// 	body, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("Error reading response body: %v", err)
// 	}

// 	var resp Response

// 	err = json.Unmarshal(body, &resp)
// 	if err != nil {
// 		return nil, fmt.Errorf("Error unmarshalling response: %v", err)
// 	}

// 	return &resp, nil
// }

func AddToChat(channelID, prompt, image, modelName string) Chat {
	activeChats.Lock()
	defer activeChats.Unlock()

	if _, ok := activeChats.data[channelID]; !ok {
		activeChats.data[channelID] = Chat{
			Model:  modelName,
			Stream: false,
			Messages: []Message{
				{
					Role:    "user",
					Context: prompt,
					Images:  []string{image},
				},
			},
		}

		return activeChats.data[channelID]
	}

	chat := activeChats.data[channelID]

	chat.Messages = append(activeChats.data[channelID].Messages, Message{
		Role:    "user",
		Context: prompt,
		Images:  []string{image},
	})

	activeChats.data[channelID] = chat

	return activeChats.data[channelID]

}

func GetImageBase64(m *discordgo.MessageCreate) (string, error) {
	if len(m.Message.Attachments) < 0 {
		return "", nil
	}

	imageURL := m.Message.Attachments[0].URL

	response, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("Error while getting image: %v", err)
	}
	defer response.Body.Close()

	imageData, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading image: %v", err)
	}

	return base64.StdEncoding.EncodeToString(imageData), nil
}

func Generate(payload Chat, prompt, baseURL string) (*Response, error) {

	url := fmt.Sprintf("http://%s/api/chat", baseURL)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("Error marshalling request: %v", err)
	}

	res, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("Error sending request to API: %v, Prompt: %s", err, prompt)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error response code: %v", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %v", err)
	}

	var resp Response

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling response: %v", err)
	}

	return &resp, nil
}
