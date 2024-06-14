package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
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

type User struct {
	RequestsCount int
	EndOfCooldown time.Time
}

type ApiConfig struct {
	ApiDomain         string `json:"api-domain"`
	Model             string `json:"model"`
	HistoryTokensSize int    `json:"history-tokens-size"`
	Channels          map[string]*History
	Users             map[string]*User
}

func (api *ApiConfig) UpdateUserCounter(userId string) bool {
	user, exist := api.Users[userId]
	if !exist {
		user = &User{}
		api.Users[userId] = user
	}

	if user.RequestsCount >= 6 {
		if user.EndOfCooldown.IsZero() {
			user.EndOfCooldown = time.Now().Add(30 * time.Minute)
		}
		return false
	}

	user.RequestsCount++

	return true
}

func (api *ApiConfig) ResetUsersCounter(delay time.Duration) {
	expirationTime := time.Now().Add(delay)

	for id, user := range api.Users {
		if user.EndOfCooldown.Before(expirationTime) {
			delete(api.Users, id)
		}
	}
}

const prompt string = `
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

func (api *ApiConfig) Generate(content, referenceContent, channelId string) (*Response, error) {
	url := api.ApiDomain + "/api/generate"

	requestBody, err := json.Marshal(Request{
		Model:   api.Model,
		Prompt:  fmt.Sprintf(prompt, content, referenceContent),
		Context: api.GetHistory(channelId),
		Stream:  false,
	})
	if err != nil {
		log.Printf("Error marshalling request: %v", err)
		return nil, err
	}

	res, err := http.Post(url, "application/json", strings.NewReader(string(requestBody)))
	if err != nil {
		log.Printf("Error sending request to API: %v, Prompt: %s, Reference: %s", err, content, referenceContent)
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		log.Printf("Error response: %s,\nPrompt: %s, Reference: %s", res.Status, content, referenceContent)
		return nil, errors.New("Response not 200 OK")
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil, err
	}

	var formatted Response

	err = json.Unmarshal(body, &formatted)
	if err != nil {
		log.Printf("Error unmarshalling response: %v", err)
		return nil, err
	}

	return &formatted, nil
}
