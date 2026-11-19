package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Completer mengirim prompt ke OpenAI Chat Completions.
type Completer interface {
	Complete(ctx context.Context, system, user string) (string, error)
}

type OpenAIClient struct {
	Endpoint string
	Key      string
	Model    string
	HTTP     *http.Client
}

func NewOpenAIClient(endpoint, key, model string) *OpenAIClient {
	return &OpenAIClient{Endpoint: endpoint, Key: key, Model: model, HTTP: &http.Client{Timeout: 120 * time.Second}}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (o *OpenAIClient) Complete(ctx context.Context, system, user string) (string, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"model":       o.Model,
		"temperature": 0.7,
		"messages":    []chatMessage{{Role: "system", Content: system}, {Role: "user", Content: user}},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(o.Endpoint, "/")+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.Key)

	resp, err := o.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("OpenAI mengembalikan status %d", resp.StatusCode)
	}

	var body struct {
		Choices []struct {
			Message chatMessage `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	if len(body.Choices) == 0 {
		return "", errors.New("respons OpenAI kosong")
	}
	return body.Choices[0].Message.Content, nil
}
