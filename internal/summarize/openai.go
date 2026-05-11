package summarize

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

const (
	DefaultModel = openai.GPT4oMini
	maxTokens    = 8192
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
	choice := resp.Choices[0]
	slog.Info("openai chat response",
		"model", c.model,
		"finish_reason", choice.FinishReason,
		"prompt_tokens", resp.Usage.PromptTokens,
		"completion_tokens", resp.Usage.CompletionTokens,
		"total_tokens", resp.Usage.TotalTokens,
	)
	if choice.FinishReason == openai.FinishReasonLength {
		slog.Warn("summary may be truncated; consider raising max_tokens or using --summary-model gpt-4o")
	}
	out := strings.TrimSpace(choice.Message.Content)
	if out == "" {
		return "", fmt.Errorf("openai returned empty content")
	}
	return out, nil
}
