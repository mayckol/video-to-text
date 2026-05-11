package summarize

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

const (
	DefaultModel = openai.GPT4oMini
	maxTokens    = 4096
)

type Client struct {
	api   *openai.Client
	model string
}

func NewClient(apiKey, model string) *Client {
	if model == "" {
		model = DefaultModel
	}
	return &Client{api: openai.NewClient(apiKey), model: model}
}

func (c *Client) Summarize(ctx context.Context, prompt string) (string, error) {
	resp, err := c.api.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     c.model,
		MaxTokens: maxTokens,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
	})
	if err != nil {
		return "", fmt.Errorf("openai chat: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai returned no choices")
	}
	out := strings.TrimSpace(resp.Choices[0].Message.Content)
	if out == "" {
		return "", fmt.Errorf("openai returned empty content")
	}
	return out, nil
}
