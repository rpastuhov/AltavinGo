package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

type Choice struct {
	FinishReason string  `json:"finish_reason"`
	Message      Message `json:"message"`
}

type Response struct {
	Choices []Choice `json:"choices"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
}

func GetImageBase64(m *discordgo.MessageCreate) (string, error) {
	if len(m.Message.Attachments) == 0 {
		return "", nil
	}

	imageURL := m.Message.Attachments[0].URL

	response, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("while getting image: %v", err)
	}
	defer response.Body.Close()

	imageData, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("reading image: %v", err)
	}

	return base64.StdEncoding.EncodeToString(imageData), nil
}

func Generate(payload *Chat, prompt, baseURL, token string) (*Response, error) {
	url := fmt.Sprintf("https://%s/v1/chat/completions", baseURL)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("sending request to API: %v, Prompt: %s", err, prompt)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request to API: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response code: %v", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	// Выводим содержимое тела ответа в формате JSON
	fmt.Println("Response Body (JSON):")
	var jsonBody map[string]interface{}
	json.Unmarshal(body, &jsonBody)
	jsonFormatted, _ := json.MarshalIndent(jsonBody, "", "  ")
	fmt.Println(string(jsonFormatted))

	var resUnmarshal Response
	err = json.Unmarshal(body, &resUnmarshal)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling response: %v", err)
	}

	return &resUnmarshal, nil
}
