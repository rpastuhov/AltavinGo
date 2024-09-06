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

func GetImageBase64(m *discordgo.MessageCreate) (string, error) {
	if len(m.Message.Attachments) == 0 {
		return "", nil
	}

	imageURL := m.Message.Attachments[0].URL

	response, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("[ERROR]: while getting image: %v", err)
	}
	defer response.Body.Close()

	imageData, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("[ERROR]: reading image: %v", err)
	}

	return base64.StdEncoding.EncodeToString(imageData), nil
}

func Generate(payload Chat, prompt, baseURL string) (*Response, error) {

	// url := fmt.Sprintf("http://%s/api/chat", baseURL)
	url := fmt.Sprintf("https://%s/v1/chat/completions", baseURL)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("[ERROR]: marshalling request: %v", err)
	}

	// res, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	// if err != nil {
	// 	return nil, fmt.Errorf("[ERROR]: sending request to API: %v, Prompt: %s", err, prompt)
	// }

	// defer res.Body.Close()

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("[ERROR]: sending request to API: %v, Prompt: %s", err, prompt)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer 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")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[ERROR]: sending request to API: %v", err)
	}
	defer res.Body.Close()

	// if res.StatusCode != http.StatusOK {
	// 	return nil, fmt.Errorf("[ERROR]: response code: %v", res.StatusCode)
	// }

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("[ERROR]: reading response body: %v", err)
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
		return nil, fmt.Errorf("[ERROR]: unmarshalling response: %v", err)
	}

	return &resUnmarshal, nil
}
