package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
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
		Prompt:  content,
		Context: api.GetHistory(channelId),
		Stream:  false,
	})
	if err != nil {
		log.Printf("Error marshalling request: %v", err)
		return nil, err
	}

	res, err := http.Post(url, "application/json", strings.NewReader(string(requestBody)))
	if err != nil {
		log.Printf("Error sending request to API: %v, Prompt: %s", err, content)
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		log.Printf("Error response: %s,\nPrompt: %s", res.Status, content)
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
