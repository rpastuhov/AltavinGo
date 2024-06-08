package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
)

// const role string = "Responses must always be less than 2000 characters. "

const role string = `Your task is to interact with users who mention you by responding to
their messages. Always use the same language as the user and try to
communicate in a way that makes you seem human. Responses must
always be less than 2000 characters. Keep in mind that your response is used in the API
response in json format. Here is the text of the message in
which you were mentioned:
`

type Request struct {
	Model   string `json:"model"`
	Prompt  string `json:"prompt"`
	Context []int  `json:"context"`
	Stream  bool   `json:"stream"`
	Format  string `json:"format"`
}

type Response struct {
	Response      string `json:"response"`
	Context       []int  `json:"context"`
	TotalDuration int    `json:"total_duration"`
}

type ApiConfig struct {
	ApiDomain         string `json:"api-domain"`
	Model             string `json:"model"`
	HistoryTokensSize int    `json:"history-tokens-size"`
	Channels          map[string]*History
}

func (api *ApiConfig) Generate(content, channelId string) (*Response, error) {
	url := api.ApiDomain + "/api/generate"

	requestBody, err := json.Marshal(Request{
		Model:   api.Model,
		Prompt:  role + content,
		Context: api.GetHistory(channelId),
		Stream:  false,
		Format:  "",
	})
	if err != nil {
		log.Println("Error marshal Request: ", err)
		return nil, err
	}

	res, err := http.Post(url, "application/json", strings.NewReader(string(requestBody)))
	if err != nil {
		log.Printf("Error sending request to API: %v,\nPrompt: %s\n", err, content)
		return nil, err
	}

	if res.Status != "200 OK" {
		log.Printf("Error response: %s,\nPrompt: %s\n", res.Status, content)
		return nil, errors.New("Response not 200 OK")
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("Error reading from response body: ", err)
		return nil, err
	}

	var formatted Response

	err = json.Unmarshal(body, &formatted)
	if err != nil {
		log.Println("Error unmarshal Response:", err)
		return nil, err
	}

	return &formatted, nil
}
