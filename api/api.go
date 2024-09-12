package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Usage struct {
	TotalTime   float64 `json:"total_time"`
	TotalTokens int     `json:"total_tokens"`
}

type Choice struct {
	FinishReason string  `json:"finish_reason"`
	Message      Message `json:"message"`
}

type Response struct {
	Choices []Choice `json:"choices"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Usage   Usage    `json:"usage"`
}

// func GetImageBase64(m *discordgo.MessageCreate) (string, error) {
// 	if len(m.Message.Attachments) == 0 {
// 		return "", nil
// 	}

// 	imageURL := m.Message.Attachments[0].URL

// 	response, err := http.Get(imageURL)
// 	if err != nil {
// 		return "", fmt.Errorf("while getting image: %v", err)
// 	}
// 	defer response.Body.Close()

// 	imageData, err := io.ReadAll(response.Body)
// 	if err != nil {
// 		return "", fmt.Errorf("reading image: %v", err)
// 	}

// 	return base64.StdEncoding.EncodeToString(imageData), nil
// }

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

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		var jsonBody map[string]interface{}
		json.Unmarshal(body, &jsonBody)

		responseForm, _ := json.Marshal(jsonBody)
		payloadForm, _ := json.Marshal(payload)

		return nil, fmt.Errorf("code: %v: response: %v payload: %v", res.StatusCode, string(responseForm), string(payloadForm))
	}

	var resUnmarshal Response
	err = json.Unmarshal(body, &resUnmarshal)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling response: %v", err)
	}

	return &resUnmarshal, nil
}
